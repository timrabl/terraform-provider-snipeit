package assets

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	assetsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/assets"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &HardwareResource{}
	_ resource.ResourceWithImportState = &HardwareResource{}
)

// NewHardwareResource returns a new snipeit_hardware resource.
func NewHardwareResource() resource.Resource {
	return &HardwareResource{}
}

// HardwareResource manages a Snipe-IT asset.
type HardwareResource struct {
	svc *assetsapi.Service
}

// HardwareResourceModel describes the resource data model.
type HardwareResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	AssetTag       types.String `tfsdk:"asset_tag"`
	ModelID        types.Int64  `tfsdk:"model_id"`
	StatusID       types.Int64  `tfsdk:"status_id"`
	Name           types.String `tfsdk:"name"`
	Serial         types.String `tfsdk:"serial"`
	OrderNumber    types.String `tfsdk:"order_number"`
	PurchaseDate   types.String `tfsdk:"purchase_date"`
	WarrantyMonths types.Int64  `tfsdk:"warranty_months"`
	CompanyID      types.Int64  `tfsdk:"company_id"`
	SupplierID     types.Int64  `tfsdk:"supplier_id"`
	LocationID     types.Int64  `tfsdk:"location_id"`
	Requestable    types.Bool   `tfsdk:"requestable"`
	Notes          types.String `tfsdk:"notes"`
}

func (r *HardwareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hardware"
}

func (r *HardwareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an asset (hardware) in Snipe-IT.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the asset.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"asset_tag": schema.StringAttribute{
				MarkdownDescription: "Unique asset tag.",
				Required:            true,
			},
			"model_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the asset model.",
				Required:            true,
			},
			"status_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the status label.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name of the asset.",
				Optional:            true,
			},
			"serial": schema.StringAttribute{
				MarkdownDescription: "Serial number of the asset.",
				Optional:            true,
			},
			"order_number": schema.StringAttribute{
				MarkdownDescription: "Order number of the purchase.",
				Optional:            true,
			},
			"purchase_date": schema.StringAttribute{
				MarkdownDescription: "Purchase date in `YYYY-MM-DD` format.",
				Optional:            true,
			},
			"warranty_months": schema.Int64Attribute{
				MarkdownDescription: "Warranty duration in months.",
				Optional:            true,
			},
			"company_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the company owning this asset.",
				Optional:            true,
			},
			"supplier_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the supplier of this asset.",
				Optional:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the default (ready-to-deploy) location of this asset.",
				Optional:            true,
			},
			"requestable": schema.BoolAttribute{
				MarkdownDescription: "Whether this asset can be requested by users.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Free-form notes.",
				Optional:            true,
			},
		},
	}
}

func (r *HardwareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = assetsapi.New(c)
	}
}

func (m *HardwareResourceModel) toBody() map[string]any {
	body := map[string]any{
		"asset_tag": m.AssetTag.ValueString(),
		"model_id":  m.ModelID.ValueInt64(),
		"status_id": m.StatusID.ValueInt64(),
	}
	tfutil.BodyString(body, "name", m.Name)
	tfutil.BodyString(body, "serial", m.Serial)
	tfutil.BodyString(body, "order_number", m.OrderNumber)
	tfutil.BodyString(body, "notes", m.Notes)
	tfutil.BodyNullableString(body, "purchase_date", m.PurchaseDate)
	tfutil.BodyNullableInt(body, "warranty_months", m.WarrantyMonths)
	tfutil.BodyNullableInt(body, "company_id", m.CompanyID)
	tfutil.BodyNullableInt(body, "supplier_id", m.SupplierID)
	// The API expects the default location under rtd_location_id.
	tfutil.BodyNullableInt(body, "rtd_location_id", m.LocationID)
	tfutil.BodyOptBool(body, "requestable", m.Requestable)
	return body
}

func (m *HardwareResourceModel) fromAPI(api *assetsapi.Hardware) {
	m.ID = types.Int64Value(api.Id)
	m.AssetTag = types.StringValue(api.AssetTag)
	m.ModelID = types.Int64Value(api.Model.IDOrZero())
	m.StatusID = types.Int64Value(api.StatusLabel.IDOrZero())
	m.Name = tfutil.StateStringPtr(api.Name)
	m.Serial = tfutil.StateStringPtr(api.Serial)
	m.OrderNumber = tfutil.StateStringPtr(api.OrderNumber)
	m.Notes = tfutil.StateStringPtr(api.Notes)
	m.CompanyID = tfutil.StateRefID(api.Company)
	m.SupplierID = tfutil.StateRefID(api.Supplier)
	m.LocationID = tfutil.StateRefID(api.RtdLocation)
	m.WarrantyMonths = tfutil.StateOptInt(int64(api.WarrantyMonths))
	m.Requestable = types.BoolValue(bool(api.Requestable))
	if api.PurchaseDate != nil && api.PurchaseDate.Date != "" {
		m.PurchaseDate = types.StringValue(api.PurchaseDate.Date)
	} else {
		m.PurchaseDate = types.StringNull()
	}
}

func (r *HardwareResource) read(ctx context.Context, id int64, data *HardwareResourceModel) error {
	api, err := r.svc.GetHardware(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *HardwareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HardwareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateHardware(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create asset", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read asset after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HardwareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HardwareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read asset", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HardwareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data HardwareResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateHardware(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update asset", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read asset after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HardwareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HardwareResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteHardware(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete asset", err.Error())
	}
}

func (r *HardwareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
