package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/terraform-provider-checkmk/internal/client"
	"github.com/terraform-provider-checkmk/internal/common"
	"github.com/terraform-provider-checkmk/internal/configuration"
	"github.com/terraform-provider-checkmk/internal/groups"
	"github.com/terraform-provider-checkmk/internal/hosts"
	"github.com/terraform-provider-checkmk/internal/labels"
	"github.com/terraform-provider-checkmk/internal/rules"
	"github.com/terraform-provider-checkmk/internal/rules/wrappers"
	"github.com/terraform-provider-checkmk/internal/users"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider              = &checkmkProvider{}
	_ provider.ProviderWithFunctions = &checkmkProvider{}
)

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &checkmkProvider{
			version: version,
		}
	}
}

// checkmkProvider is the provider implementation.
type checkmkProvider struct {
	version string
}

// checkmkProviderModel describes the provider data model.
type checkmkProviderModel struct {
	URL                   types.String `tfsdk:"url"`
	Username              types.String `tfsdk:"username"`
	Password              types.String `tfsdk:"password"`
	Activate              types.String `tfsdk:"activate"`
	ForceForeignChanges   types.Bool   `tfsdk:"force_foreign_changes"`
	ActivationWaitTime    types.Int64  `tfsdk:"activation_wait_time"`
	StrictResourceLocking types.Bool   `tfsdk:"strict_resource_locking"`
	RequestTimeout        types.Int64  `tfsdk:"request_timeout"`
	MaxRetries            types.Int64  `tfsdk:"max_retries"`
	InsecureSkipVerify    types.Bool   `tfsdk:"insecure_skip_verify"`
}

// ProviderData is an alias to common.ProviderData for backwards compatibility
type ProviderData = common.ProviderData

// Features is an alias to common.Features for backwards compatibility
type Features = common.Features

// NewFeatures creates feature flags based on detected version
var NewFeatures = common.NewFeatures

// Metadata returns the provider type name.
func (p *checkmkProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "checkmk"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *checkmkProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with CheckMK monitoring system.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Description: "CheckMK server URL (e.g., http://localhost:5000/test). " +
					"May also be provided via CHECKMK_URL environment variable.",
				Optional: true,
			},
			"username": schema.StringAttribute{
				Description: "CheckMK automation user username. " +
					"May also be provided via CHECKMK_USERNAME environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"password": schema.StringAttribute{
				Description: "CheckMK automation user password. " +
					"May also be provided via CHECKMK_PASSWORD environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"activate": schema.StringAttribute{
				Description: "Change activation mode. Options: 'auto', 'manual' (default). " +
					"'auto': Automatically activate changes after create/update/delete. " +
					"'manual': Changes are staged but require manual activation in CheckMK UI.",
				Optional: true,
			},
			"force_foreign_changes": schema.BoolAttribute{
				Description: "Whether to activate changes made by other users/tools (default: true). " +
					"When true, activates ALL pending changes (including those made externally). " +
					"When false, activation may fail if there are pending changes made outside Terraform.",
				Optional: true,
			},
			"activation_wait_time": schema.Int64Attribute{
				Description: "Time in seconds to wait after activation for changes to propagate (default: 5). " +
					"Increase this value if you experience eventual consistency issues. " +
					"Only applies when activate = \"auto\".",
				Optional: true,
			},
			"strict_resource_locking": schema.BoolAttribute{
				Description: "Enable strict ETag-based locking (default: false). " +
					"When false, uses If-Match: * to bypass ETag validation for both activation and resource operations (Python approach). " +
					"When true, fetches and validates ETags for activation endpoint and resource operations (enforces concurrency control).",
				Optional: true,
			},
			"request_timeout": schema.Int64Attribute{
				Description: "HTTP request timeout in seconds (default: 60). " +
					"Increase this value for slow networks or large API responses.",
				Optional: true,
			},
			"max_retries": schema.Int64Attribute{
				Description: "Maximum number of retries for failed API requests (default: 3). " +
					"Retries occur on transient errors (429, 500, 502, 503, 504) with exponential backoff.",
				Optional: true,
			},
			"insecure_skip_verify": schema.BoolAttribute{
				Description: "Skip TLS certificate verification (default: false). " +
					"WARNING: Only use this for testing with self-signed certificates. " +
					"This makes the connection insecure and should not be used in production.",
				Optional: true,
			},
		},
	}
}

// Configure prepares a CheckMK API client for data sources and resources.
func (p *checkmkProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config checkmkProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values from environment variables
	url := os.Getenv("CHECKMK_URL")
	username := os.Getenv("CHECKMK_USERNAME")
	password := os.Getenv("CHECKMK_PASSWORD")

	// Override with provider configuration if set
	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	// Validate required configuration
	if url == "" {
		resp.Diagnostics.AddError(
			"Missing CheckMK URL",
			"The provider cannot create the CheckMK API client as there is a missing or empty value for the CheckMK URL. "+
				"Set the url value in the configuration or use the CHECKMK_URL environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddError(
			"Missing CheckMK Username",
			"The provider cannot create the CheckMK API client as there is a missing or empty value for the CheckMK username. "+
				"Set the username value in the configuration or use the CHECKMK_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddError(
			"Missing CheckMK Password",
			"The provider cannot create the CheckMK API client as there is a missing or empty value for the CheckMK password. "+
				"Set the password value in the configuration or use the CHECKMK_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Build client options
	opts := &client.ClientOptions{}

	// Request timeout (default: 60 seconds)
	if !config.RequestTimeout.IsNull() {
		opts.RequestTimeout = time.Duration(config.RequestTimeout.ValueInt64()) * time.Second
	}

	// Max retries (default: 3)
	if !config.MaxRetries.IsNull() {
		opts.MaxRetries = int(config.MaxRetries.ValueInt64())
	}

	// Insecure skip verify (default: false)
	if !config.InsecureSkipVerify.IsNull() {
		opts.InsecureSkipVerify = config.InsecureSkipVerify.ValueBool()
	}

	// Create CheckMK API client with options
	c, err := client.NewClientWithOptions(url, username, password, opts)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create CheckMK API Client",
			"An unexpected error occurred when creating the CheckMK API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"CheckMK Client Error: "+err.Error(),
		)
		return
	}

	// Set provider version for User-Agent header
	c.SetProviderVersion(p.version)

	// Get activate setting (default: "manual")
	activate := "manual"
	if !config.Activate.IsNull() {
		activate = config.Activate.ValueString()
	}

	// Validate activate parameter
	if activate != "auto" && activate != "manual" {
		resp.Diagnostics.AddError(
			"Invalid Activation Mode",
			fmt.Sprintf("The 'activate' parameter must be either 'auto' or 'manual', got: %s", activate),
		)
		return
	}

	// Get force_foreign_changes setting (default: true)
	forceForeignChanges := true
	if !config.ForceForeignChanges.IsNull() {
		forceForeignChanges = config.ForceForeignChanges.ValueBool()
	}

	// Get activation_wait_time setting (default: 5 seconds)
	activationWaitTime := 5
	if !config.ActivationWaitTime.IsNull() {
		activationWaitTime = int(config.ActivationWaitTime.ValueInt64())
	}

	// Get strict_resource_locking setting (default: false)
	strictResourceLocking := false
	if !config.StrictResourceLocking.IsNull() {
		strictResourceLocking = config.StrictResourceLocking.ValueBool()
	}

	// Initialize feature flags based on detected version
	features := common.NewFeatures(c.Version)

	// Make the CheckMK client and configuration available during DataSource and Resource
	// type Configure methods.
	providerData := &common.ProviderData{
		Client:                c,
		Features:              features,
		Activate:              activate,
		ForceForeignChanges:   forceForeignChanges,
		ActivationWaitTime:    activationWaitTime,
		StrictResourceLocking: strictResourceLocking,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

// DataSources defines the data sources implemented in the provider.
func (p *checkmkProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	var datasources []func() datasource.DataSource

	// Add data sources from each package
	datasources = append(datasources, configuration.DataSources()...)
	datasources = append(datasources, users.DataSources()...)
	datasources = append(datasources, hosts.DataSources()...)
	datasources = append(datasources, groups.DataSources()...)
	datasources = append(datasources, labels.DataSources()...)
	datasources = append(datasources, rules.DataSources()...)

	return datasources
}

// Resources defines the resources implemented in the provider.
func (p *checkmkProvider) Resources(_ context.Context) []func() resource.Resource {
	var resources []func() resource.Resource

	// Add provider-level resources
	resources = append(resources, NewActivationResource)

	// Add resources from each package
	resources = append(resources, configuration.Resources()...)
	resources = append(resources, users.Resources()...)
	resources = append(resources, hosts.Resources()...)
	resources = append(resources, groups.Resources()...)
	resources = append(resources, labels.Resources()...)
	resources = append(resources, rules.Resources()...)
	// Rule wrappers are registered separately to avoid import cycles
	// (wrappers imports rules for condition utilities)
	resources = append(resources, wrappers.AllResources()...)

	return resources
}

// Functions defines the provider functions.
func (p *checkmkProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		NewToPythonFunction,
	}
}
