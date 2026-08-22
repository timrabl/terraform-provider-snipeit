package assets

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	assetsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/assets"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var _ datasource.DataSource = &HardwareDataSource{}

// NewHardwareDataSource returns a new snipeit_hardware data source.
func NewHardwareDataSource() datasource.DataSource {
	return &HardwareDataSource{}
}

// HardwareDataSource looks up a single asset by id, asset tag or serial.
type HardwareDataSource struct {
	svc *assetsapi.Service
}

// HardwareDataSourceModel describes the data source data model.
type HardwareDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	AssetTag     types.String `tfsdk:"asset_tag"`
	Serial       types.String `tfsdk:"serial"`
	Name         types.String `tfsdk:"name"`
	ModelID      types.Int64  `tfsdk:"model_id"`
	StatusID     types.Int64  `tfsdk:"status_id"`
	CompanyID    types.Int64  `tfsdk:"company_id"`
	LocationID   types.Int64  `tfsdk:"location_id"`
	Notes        types.String `tfsdk:"notes"`
	PurchaseDate types.String `tfsdk:"purchase_date"`
}

func (d *HardwareDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hardware"
}

func (d *HardwareDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a single asset by `id`, `asset_tag` or `serial`. Exactly one " +
			"of these must be set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the asset.",
				Optional:            true,
				Computed:            true,
			},
			"asset_tag": schema.StringAttribute{
				MarkdownDescription: "Asset tag to look up.",
				Optional:            true,
				Computed:            true,
			},
			"serial": schema.StringAttribute{
				MarkdownDescription: "Serial number to look up.",
				Optional:            true,
				Computed:            true,
			},
			"name":          schema.StringAttribute{MarkdownDescription: "Display name of the asset.", Computed: true},
			"model_id":      schema.Int64Attribute{MarkdownDescription: "Id of the asset model.", Computed: true},
			"status_id":     schema.Int64Attribute{MarkdownDescription: "Id of the status label.", Computed: true},
			"company_id":    schema.Int64Attribute{MarkdownDescription: "Id of the owning company.", Computed: true},
			"location_id":   schema.Int64Attribute{MarkdownDescription: "Id of the default location.", Computed: true},
			"notes":         schema.StringAttribute{MarkdownDescription: "Free-form notes.", Computed: true},
			"purchase_date": schema.StringAttribute{MarkdownDescription: "Purchase date (`YYYY-MM-DD`).", Computed: true},
		},
	}
}

func (d *HardwareDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		d.svc = assetsapi.New(c)
	}
}

func (d *HardwareDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data HardwareDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var api *assetsapi.Hardware
	var err error
	switch {
	case !data.ID.IsNull():
		api, err = d.svc.GetHardware(ctx, data.ID.ValueInt64())
	case !data.AssetTag.IsNull():
		api, err = d.svc.GetHardwareByTag(ctx, data.AssetTag.ValueString())
	case !data.Serial.IsNull():
		var list *assetsapi.HardwareList
		list, err = d.svc.FindHardwareBySerial(ctx, data.Serial.ValueString())
		if err == nil {
			if len(list.Rows) == 0 {
				resp.Diagnostics.AddError("Unable to look up asset", "No matching asset found.")
				return
			}
			api = &list.Rows[0]
		}
	default:
		resp.Diagnostics.AddError("Missing lookup key",
			"Exactly one of id, asset_tag or serial must be set.")
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to look up asset", err.Error())
		return
	}

	data.ID = types.Int64Value(api.Id)
	data.AssetTag = types.StringValue(api.AssetTag)
	data.Serial = tfutil.StateStringPtr(api.Serial)
	data.Name = tfutil.StateStringPtr(api.Name)
	data.ModelID = types.Int64Value(api.Model.IDOrZero())
	data.StatusID = types.Int64Value(api.StatusLabel.IDOrZero())
	data.CompanyID = tfutil.StateRefID(api.Company)
	data.LocationID = tfutil.StateRefID(api.RtdLocation)
	data.Notes = tfutil.StateStringPtr(api.Notes)
	if api.PurchaseDate != nil && api.PurchaseDate.Date != "" {
		data.PurchaseDate = types.StringValue(api.PurchaseDate.Date)
	} else {
		data.PurchaseDate = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
