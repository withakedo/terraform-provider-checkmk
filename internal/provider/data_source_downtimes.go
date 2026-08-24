package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/withakedo/terraform-provider-checkmk/internal/common"
)

var _ datasource.DataSource = &DowntimesDataSource{}

func NewDowntimesDataSource() datasource.DataSource {
	return &DowntimesDataSource{}
}

// DowntimesDataSource lists the downtimes currently active on a host (both
// host-level and service-level downtimes), so out-of-band or
// externally-scheduled downtimes can be observed from Terraform -
// checkmk_downtime itself has no persistent server-side object to read back
// (see its docs), so this is the way to detect that kind of drift.
type DowntimesDataSource struct {
	providerData *common.ProviderData
}

// DowntimesDataSourceModel describes the data source's data model.
type DowntimesDataSourceModel struct {
	ID        types.String         `tfsdk:"id"`
	HostName  types.String         `tfsdk:"host_name"`
	Downtimes []DowntimeEntryModel `tfsdk:"downtimes"`
}

// DowntimeEntryModel describes a single downtime within the list.
type DowntimeEntryModel struct {
	ID                  types.String `tfsdk:"id"`
	SiteID              types.String `tfsdk:"site_id"`
	Author              types.String `tfsdk:"author"`
	StartTime           types.String `tfsdk:"start_time"`
	EndTime             types.String `tfsdk:"end_time"`
	Recurring           types.Bool   `tfsdk:"recurring"`
	Comment             types.String `tfsdk:"comment"`
	IsService           types.Bool   `tfsdk:"is_service"`
	ServiceDescription  types.String `tfsdk:"service_description"`
	ModeType            types.String `tfsdk:"mode_type"`
	ModeDurationMinutes types.Int64  `tfsdk:"mode_duration_minutes"`
}

func (d *DowntimesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_downtimes"
}

func (d *DowntimesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists downtimes currently active on a host, covering both host-level and " +
			"service-level downtimes. Useful for detecting downtimes scheduled outside Terraform (in the " +
			"CheckMK UI, by another automation, or by `checkmk_downtime` resources managed elsewhere).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Same as `host_name`.",
				Computed:            true,
			},
			"host_name": schema.StringAttribute{
				MarkdownDescription: "The host to list downtimes for.",
				Required:            true,
			},
			"downtimes": schema.ListNestedAttribute{
				MarkdownDescription: "The active downtimes found for this host.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                    schema.StringAttribute{Computed: true, MarkdownDescription: "CheckMK's internal id for the downtime."},
						"site_id":               schema.StringAttribute{Computed: true, MarkdownDescription: "The site id of the downtime."},
						"author":                schema.StringAttribute{Computed: true, MarkdownDescription: "Who scheduled the downtime."},
						"start_time":            schema.StringAttribute{Computed: true, MarkdownDescription: "Start time (RFC3339)."},
						"end_time":              schema.StringAttribute{Computed: true, MarkdownDescription: "End time (RFC3339)."},
						"recurring":             schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the downtime is time-repetitive."},
						"comment":               schema.StringAttribute{Computed: true, MarkdownDescription: "The downtime's comment."},
						"is_service":            schema.BoolAttribute{Computed: true, MarkdownDescription: "True if this is a service downtime, false if it's a host downtime."},
						"service_description":   schema.StringAttribute{Computed: true, MarkdownDescription: "The service name, if `is_service` is true."},
						"mode_type":             schema.StringAttribute{Computed: true, MarkdownDescription: "`fixed` or `flexible`."},
						"mode_duration_minutes": schema.Int64Attribute{Computed: true, MarkdownDescription: "For `flexible` mode, the duration in minutes."},
					},
				},
			},
		},
	}
}

func (d *DowntimesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	providerData, ok := req.ProviderData.(*common.ProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *common.ProviderData, got: %T", req.ProviderData),
		)
		return
	}
	d.providerData = providerData
}

func (d *DowntimesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DowntimesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	hostName := data.HostName.ValueString()
	infos, err := d.providerData.Client.ListDowntimes(ctx, hostName)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list downtimes: %s", err))
		return
	}

	entries := make([]DowntimeEntryModel, len(infos))
	for i, info := range infos {
		entries[i] = DowntimeEntryModel{
			ID:                  types.StringValue(info.ID),
			SiteID:              types.StringValue(info.SiteID),
			Author:              types.StringValue(info.Author),
			StartTime:           types.StringValue(info.StartTime),
			EndTime:             types.StringValue(info.EndTime),
			Recurring:           types.BoolValue(info.Recurring),
			Comment:             types.StringValue(info.Comment),
			IsService:           types.BoolValue(info.IsService),
			ServiceDescription:  types.StringValue(info.ServiceDescription),
			ModeType:            types.StringValue(info.Mode.Type),
			ModeDurationMinutes: types.Int64Value(info.Mode.DurationMinutes),
		}
	}

	data.ID = types.StringValue(hostName)
	data.Downtimes = entries

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
