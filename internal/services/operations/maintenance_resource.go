package operations

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	operationsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/operations"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &MaintenanceResource{}
	_ resource.ResourceWithImportState = &MaintenanceResource{}
)

// NewMaintenanceResource returns a new snipeit_maintenance resource.
func NewMaintenanceResource() resource.Resource {
	return &MaintenanceResource{}
}

// MaintenanceResource manages an asset maintenance in Snipe-IT.
type MaintenanceResource struct {
	svc *operationsapi.Service
}

// MaintenanceResourceModel describes the resource data model.
type MaintenanceResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	AssetID        types.Int64  `tfsdk:"asset_id"`
	SupplierID     types.Int64  `tfsdk:"supplier_id"`
	Type           types.String `tfsdk:"maintenance_type"`
	Title          types.String `tfsdk:"title"`
	StartDate      types.String `tfsdk:"start_date"`
	CompletionDate types.String `tfsdk:"completion_date"`
	IsWarranty     types.Bool   `tfsdk:"is_warranty"`
	Notes          types.String `tfsdk:"notes"`
}

func (r *MaintenanceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance"
}

func (r *MaintenanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an asset maintenance record in Snipe-IT.\n\n" +
			"~> The `cost` field is intentionally not supported: the API returns it formatted " +
			"according to the instance locale, which cannot round-trip stably through Terraform state.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the maintenance.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"asset_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the asset the maintenance belongs to.",
				Required:            true,
			},
			"supplier_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the supplier performing the maintenance.",
				Required:            true,
			},
			"maintenance_type": schema.StringAttribute{
				MarkdownDescription: "Type of the maintenance. Common values: `Maintenance`, `Repair`, " +
					"`Upgrade`, `PAT test`, `Calibration`, `Software Support`, `Hardware Support`. " +
					"The API does not validate this value.",
				Required: true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Title of the maintenance.",
				Required:            true,
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "Start date in `YYYY-MM-DD` format.",
				Required:            true,
			},
			"completion_date": schema.StringAttribute{
				MarkdownDescription: "Completion date in `YYYY-MM-DD` format.",
				Optional:            true,
			},
			"is_warranty": schema.BoolAttribute{
				MarkdownDescription: "Whether the maintenance is covered by warranty.",
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

func (r *MaintenanceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = operationsapi.New(c)
	}
}

func (m *MaintenanceResourceModel) toBody() map[string]any {
	body := map[string]any{
		"asset_id":               m.AssetID.ValueInt64(),
		"supplier_id":            m.SupplierID.ValueInt64(),
		"asset_maintenance_type": m.Type.ValueString(),
		"title":                  m.Title.ValueString(),
		"start_date":             m.StartDate.ValueString(),
	}
	tfutil.BodyNullableString(body, "completion_date", m.CompletionDate)
	tfutil.BodyString(body, "notes", m.Notes)
	tfutil.BodyOptBool(body, "is_warranty", m.IsWarranty)
	return body
}

func (m *MaintenanceResourceModel) fromAPI(api *operationsapi.Maintenance) {
	m.ID = types.Int64Value(api.Id)
	m.AssetID = types.Int64Value(api.Asset.IDOrZero())
	m.SupplierID = types.Int64Value(api.Supplier.IDOrZero())
	m.Type = types.StringValue(api.AssetMaintenanceType)
	m.Title = types.StringValue(api.Title)
	m.IsWarranty = types.BoolValue(bool(api.IsWarranty))
	m.Notes = tfutil.StateStringPtr(api.Notes)
	if api.StartDate != nil && api.StartDate.Date != "" {
		m.StartDate = types.StringValue(api.StartDate.Date)
	} else {
		m.StartDate = types.StringNull()
	}
	if api.CompletionDate != nil && api.CompletionDate.Date != "" {
		m.CompletionDate = types.StringValue(api.CompletionDate.Date)
	} else {
		m.CompletionDate = types.StringNull()
	}
}

func (r *MaintenanceResource) read(ctx context.Context, id int64, data *MaintenanceResourceModel) error {
	api, err := r.svc.GetMaintenance(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *MaintenanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MaintenanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateMaintenance(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create maintenance", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read maintenance after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaintenanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MaintenanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read maintenance", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaintenanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MaintenanceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateMaintenance(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update maintenance", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read maintenance after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MaintenanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MaintenanceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteMaintenance(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete maintenance", err.Error())
	}
}

func (r *MaintenanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
