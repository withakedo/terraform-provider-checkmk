package hosts

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/withakedo/terraform-provider-checkmk/internal/client"
	"github.com/withakedo/terraform-provider-checkmk/internal/common"
)

var (
	_ resource.Resource                   = &HostsBulkResource{}
	_ resource.ResourceWithValidateConfig = &HostsBulkResource{}
)

func NewHostsBulkResource() resource.Resource {
	return &HostsBulkResource{}
}

// HostsBulkResource manages many CheckMK hosts through a single Terraform
// resource, using CheckMK's bulk-create/bulk-update/bulk-delete endpoints
// instead of one API call per host. This is a performance-oriented
// alternative to declaring many individual checkmk_host resources with
// for_each: a create/update/delete here is 1 API call regardless of how
// many hosts are listed, instead of N.
//
// CheckMK's bulk endpoints are not transactional: a partially-failing
// bulk-create or bulk-update can leave some hosts created/updated on the
// CheckMK side even though Terraform sees the whole operation as failed
// (and therefore does not record it in state). The error message lists
// which host names succeeded vs failed; reconcile manually.
type HostsBulkResource struct {
	providerData *common.ProviderData
}

// HostsBulkResourceModel describes the resource data model.
type HostsBulkResourceModel struct {
	ID                  types.String         `tfsdk:"id"`
	Host                []HostBulkEntryModel `tfsdk:"host"`
	Activate            types.String         `tfsdk:"activate"`
	ForceForeignChanges types.Bool           `tfsdk:"force_foreign_changes"`
}

// HostBulkEntryModel describes a single host within a checkmk_hosts_bulk resource.
type HostBulkEntryModel struct {
	HostName   types.String `tfsdk:"host_name"`
	Folder     types.String `tfsdk:"folder"`
	Attributes types.Map    `tfsdk:"attributes"`
}

func (r *HostsBulkResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hosts_bulk"
}

func (r *HostsBulkResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Manages many CheckMK hosts through CheckMK's bulk-create/bulk-update/
bulk-delete endpoints, so a single apply issues 1 API call per operation regardless of host count,
instead of the N calls ` + "`checkmk_host`" + ` with ` + "`for_each`" + ` would make. Prefer this
resource over many ` + "`checkmk_host`" + ` instances when you have a large, uniformly-managed
fleet of hosts and API call volume during apply is a bottleneck.

CheckMK's bulk endpoints are not transactional: if one host in the batch fails validation, the
whole call fails, but hosts that were already processed before the failing one are not rolled
back. A failed apply may therefore leave some hosts created/updated on the CheckMK side without
Terraform recording it in state. The error message lists which host names succeeded and which
failed; reconcile manually (fix the failing entries and re-apply, or import an unexpectedly
already-created host into a separate ` + "`checkmk_host`" + ` resource).

Does not support moving a host between folders after creation (matching ` + "`checkmk_host`" + `);
changing ` + "`folder`" + ` on an existing entry has no effect.
`,
		Attributes: map[string]schema.Attribute{
			"id": common.ComputedIDAttribute("Unique identifier for this bulk resource instance."),
			"host": schema.ListNestedAttribute{
				MarkdownDescription: "The hosts to manage. Each entry's `host_name` must be unique within this resource.",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"host_name": schema.StringAttribute{
							MarkdownDescription: "The hostname or IP address of the host.",
							Required:            true,
						},
						"folder": schema.StringAttribute{
							MarkdownDescription: "The folder path where the host will be created. Default is '/' (root folder).",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"attributes": schema.MapAttribute{
							MarkdownDescription: "Host attributes as key-value pairs (see `checkmk_host` for details). " +
								"All attributes are replaced on update (full replacement strategy).",
							ElementType: types.StringType,
							Optional:    true,
						},
					},
				},
			},
			"activate":              common.ActivateAttribute(),
			"force_foreign_changes": common.ForceForeignChangesAttribute(),
		},
	}
}

func (r *HostsBulkResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.providerData = common.ConfigureResource(req, resp)
}

func (r *HostsBulkResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HostsBulkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildBaseConfig(r.providerData, data.Activate, data.ForceForeignChanges, types.BoolNull())
	promoter := common.NewAttributePromoter(r.providerData, "checkmk_host")

	entries := make([]client.HostCreateRequest, 0, len(data.Host))
	for _, h := range data.Host {
		entries = append(entries, client.HostCreateRequest{
			HostName:   h.HostName.ValueString(),
			Folder:     resolveFolder(h.Folder),
			Attributes: attributesToAPI(promoter, h.Attributes),
		})
	}

	tflog.Info(ctx, "Bulk-creating CheckMK hosts", map[string]interface{}{"count": len(entries)})

	result, err := r.providerData.Client.BulkCreateHosts(ctx, &client.BulkCreateHostRequest{Entries: entries})
	if err != nil {
		common.AddClientError(resp, "bulk-create", "hosts", err)
		return
	}

	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "host"); err != nil {
		common.AddActivationWarning(resp, "Hosts", "bulk-created", err)
	}

	data.ID = types.StringValue(bulkHostsID(entryHostNames(data.Host)))
	applyResolvedFolders(data.Host, result.Value)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostsBulkResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HostsBulkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	promoter := common.NewAttributePromoter(r.providerData, "checkmk_host")
	refreshed := make([]HostBulkEntryModel, 0, len(data.Host))

	for _, h := range data.Host {
		host, err := r.providerData.Client.GetHost(ctx, h.HostName.ValueString())
		if err != nil {
			if apiErr, ok := err.(*client.APIError); ok && apiErr.Status == 404 {
				// Host no longer exists on the CheckMK side; drop it so the
				// next plan shows it needs to be recreated.
				continue
			}
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read host %q: %s", h.HostName.ValueString(), err))
			return
		}

		h.Folder = types.StringValue(host.Extensions.Folder)

		if len(host.Extensions.Attributes) > 0 && !h.Attributes.IsNull() {
			attrMap := make(map[string]string)
			for key := range h.Attributes.Elements() {
				if value, exists := host.Extensions.Attributes[promoter.APIKey(key)]; exists {
					if strValue, ok := value.(string); ok {
						attrMap[key] = strValue
					}
				}
			}
			var diags diag.Diagnostics
			h.Attributes, diags = types.MapValueFrom(ctx, types.StringType, attrMap)
			resp.Diagnostics.Append(diags...)
		}

		refreshed = append(refreshed, h)
	}

	if len(refreshed) == 0 && len(data.Host) > 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	data.Host = refreshed
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HostsBulkResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state HostsBulkResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildBaseConfig(r.providerData, plan.Activate, plan.ForceForeignChanges, types.BoolNull())
	promoter := common.NewAttributePromoter(r.providerData, "checkmk_host")

	oldNames := make(map[string]bool, len(state.Host))
	for _, h := range state.Host {
		oldNames[h.HostName.ValueString()] = true
	}

	var toCreate []client.HostCreateRequest
	var toUpdate []client.BulkUpdateHostEntry
	newNames := make(map[string]bool, len(plan.Host))

	for _, h := range plan.Host {
		name := h.HostName.ValueString()
		newNames[name] = true
		attrs := attributesToAPI(promoter, h.Attributes)

		if oldNames[name] {
			toUpdate = append(toUpdate, client.BulkUpdateHostEntry{HostName: name, Attributes: attrs})
		} else {
			toCreate = append(toCreate, client.HostCreateRequest{HostName: name, Folder: resolveFolder(h.Folder), Attributes: attrs})
		}
	}

	var toDelete []string
	for name := range oldNames {
		if !newNames[name] {
			toDelete = append(toDelete, name)
		}
	}

	var created []client.Host
	if len(toCreate) > 0 {
		tflog.Info(ctx, "Bulk-creating CheckMK hosts", map[string]interface{}{"count": len(toCreate)})
		result, err := r.providerData.Client.BulkCreateHosts(ctx, &client.BulkCreateHostRequest{Entries: toCreate})
		if err != nil {
			common.AddClientError(resp, "bulk-create", "hosts", err)
			return
		}
		created = result.Value
	}

	if len(toUpdate) > 0 {
		tflog.Info(ctx, "Bulk-updating CheckMK hosts", map[string]interface{}{"count": len(toUpdate)})
		if _, err := r.providerData.Client.BulkUpdateHosts(ctx, &client.BulkUpdateHostRequest{Entries: toUpdate}); err != nil {
			common.AddClientError(resp, "bulk-update", "hosts", err)
			return
		}
	}

	if len(toDelete) > 0 {
		tflog.Info(ctx, "Bulk-deleting CheckMK hosts", map[string]interface{}{"count": len(toDelete)})
		if err := r.providerData.Client.BulkDeleteHosts(ctx, toDelete); err != nil {
			common.AddClientError(resp, "bulk-delete", "hosts", err)
			return
		}
	}

	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "host"); err != nil {
		common.AddActivationWarning(resp, "Hosts", "updated", err)
	}

	plan.ID = types.StringValue(bulkHostsID(entryHostNames(plan.Host)))
	applyResolvedFolders(plan.Host, created)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *HostsBulkResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HostsBulkResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := common.BuildBaseConfig(r.providerData, data.Activate, data.ForceForeignChanges, types.BoolNull())
	names := entryHostNames(data.Host)

	tflog.Info(ctx, "Bulk-deleting CheckMK hosts", map[string]interface{}{"count": len(names)})

	if err := r.providerData.Client.BulkDeleteHosts(ctx, names); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to bulk-delete hosts: %s", err))
		return
	}

	if err := common.TrackAndActivate(ctx, r.providerData, cfg, "host"); err != nil {
		common.AddActivationWarning(resp, "Hosts", "deleted", err)
	}
}

// ValidateConfig validates the resource configuration against the generated
// OpenAPI types for the connected CheckMK version, and rejects duplicate
// host_name entries within the same resource.
func (r *HostsBulkResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	if r.providerData == nil {
		return
	}

	var data HostsBulkResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	seen := make(map[string]bool, len(data.Host))
	validator := common.NewAttributeValidator(r.providerData)
	for i, h := range data.Host {
		name := h.HostName.ValueString()
		if name != "" {
			if seen[name] {
				resp.Diagnostics.AddAttributeError(
					path.Root("host").AtListIndex(i).AtName("host_name"),
					"Duplicate host_name",
					fmt.Sprintf("Host %q is listed more than once in this checkmk_hosts_bulk resource.", name),
				)
			}
			seen[name] = true
		}
		resp.Diagnostics.Append(validator.ValidateHostAttributes(ctx, h.Attributes, path.Root("host").AtListIndex(i).AtName("attributes"))...)
	}
}

func resolveFolder(folder types.String) string {
	if folder.IsNull() || folder.IsUnknown() {
		return "/"
	}
	return folder.ValueString()
}

func attributesToAPI(promoter *common.AttributePromoter, m types.Map) map[string]interface{} {
	out := make(map[string]interface{})
	if m.IsNull() {
		return out
	}
	for key, value := range m.Elements() {
		if strValue, ok := value.(types.String); ok {
			out[promoter.APIKey(key)] = strValue.ValueString()
		}
	}
	return out
}

func entryHostNames(hosts []HostBulkEntryModel) []string {
	names := make([]string, len(hosts))
	for i, h := range hosts {
		names[i] = h.HostName.ValueString()
	}
	return names
}

// applyResolvedFolders fills in the resolved folder for entries that left
// it unset in config, using the freshly-created/updated host objects
// returned by the bulk API.
func applyResolvedFolders(hosts []HostBulkEntryModel, result []client.Host) {
	byName := make(map[string]client.Host, len(result))
	for _, h := range result {
		byName[h.ID] = h
	}
	for i := range hosts {
		if hosts[i].Folder.IsNull() || hosts[i].Folder.IsUnknown() {
			if h, ok := byName[hosts[i].HostName.ValueString()]; ok {
				hosts[i].Folder = types.StringValue(h.Extensions.Folder)
			} else {
				hosts[i].Folder = types.StringValue("/")
			}
		}
	}
}

// bulkHostsID derives a stable synthetic id for this resource instance from
// its managed host names, since a bulk resource has no natural single
// identifier of its own.
func bulkHostsID(hostNames []string) string {
	sorted := append([]string(nil), hostNames...)
	sort.Strings(sorted)
	sum := sha256.Sum256([]byte(strings.Join(sorted, "\n")))
	return "hosts-bulk-" + hex.EncodeToString(sum[:])[:16]
}
