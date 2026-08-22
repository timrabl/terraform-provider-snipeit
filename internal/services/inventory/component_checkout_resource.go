package inventory

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	inventoryapi "github.com/timrabl/terraform-provider-snipeit/internal/api/inventory"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var _ resource.Resource = &ComponentCheckoutResource{}

// NewComponentCheckoutResource returns a new snipeit_component_checkout resource.
func NewComponentCheckoutResource() resource.Resource {
	return &ComponentCheckoutResource{}
}

// ComponentCheckoutResource models the checkout of component units to an
// asset. Create checks out, Delete checks the units back in via the
// component-asset (pivot) row id.
type ComponentCheckoutResource struct {
	svc *inventoryapi.Service
}

// ComponentCheckoutResourceModel describes the resource data model.
type ComponentCheckoutResourceModel struct {
	ID          types.Int64 `tfsdk:"id"`
	ComponentID types.Int64 `tfsdk:"component_id"`
	AssetID     types.Int64 `tfsdk:"asset_id"`
	Qty         types.Int64 `tfsdk:"qty"`
}

func (r *ComponentCheckoutResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component_checkout"
}

func (r *ComponentCheckoutResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks units of a component out to an asset. Creating the resource " +
			"performs the checkout, destroying it checks the same quantity back in. All attributes " +
			"force replacement.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Id of the component-asset (pivot) row, used for checkin.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"component_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the component to check out.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"asset_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the asset receiving the component units.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"qty": schema.Int64Attribute{
				MarkdownDescription: "Number of units to check out.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *ComponentCheckoutResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = inventoryapi.New(c)
	}
}

func (r *ComponentCheckoutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ComponentCheckoutResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	componentID := data.ComponentID.ValueInt64()
	assetID := data.AssetID.ValueInt64()

	before, err := r.svc.ListComponentAssets(ctx, componentID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list component checkouts", err.Error())
		return
	}
	known := make(map[int64]bool, len(before.Rows))
	for _, row := range before.Rows {
		known[row.AssignedPivotId] = true
	}

	body := map[string]any{
		"assigned_to":  assetID,
		"assigned_qty": data.Qty.ValueInt64(),
	}
	if err := r.svc.CheckoutComponent(ctx, componentID, body); err != nil {
		resp.Diagnostics.AddError("Unable to check out component", err.Error())
		return
	}

	after, err := r.svc.ListComponentAssets(ctx, componentID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list component checkouts after checkout", err.Error())
		return
	}
	var pivotID int64
	for _, row := range after.Rows {
		if known[row.AssignedPivotId] || row.Id != assetID {
			continue
		}
		if row.AssignedPivotId > pivotID {
			pivotID = row.AssignedPivotId
		}
	}
	if pivotID == 0 {
		resp.Diagnostics.AddError("Unable to identify component checkout",
			"The checkout succeeded but its pivot row could not be found.")
		return
	}

	data.ID = types.Int64Value(pivotID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ComponentCheckoutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ComponentCheckoutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.svc.ListComponentAssets(ctx, data.ComponentID.ValueInt64())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read component checkout", err.Error())
		return
	}
	for _, row := range list.Rows {
		if row.AssignedPivotId == data.ID.ValueInt64() {
			data.Qty = types.Int64Value(int64(row.Qty))
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	// Checked in outside of Terraform.
	resp.State.RemoveResource(ctx)
}

func (r *ComponentCheckoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes RequiresReplace; Update is never called.
	resp.Diagnostics.AddError("Unexpected update", "snipeit_component_checkout does not support in-place updates.")
}

func (r *ComponentCheckoutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ComponentCheckoutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Checkin addresses the pivot row id, not the component id.
	body := map[string]any{"checkin_qty": data.Qty.ValueInt64()}
	err := r.svc.CheckinComponent(ctx, data.ID.ValueInt64(), body)
	if err != nil && !errors.Is(err, client.ErrNotFound) && !isAlreadyCheckedInError(err) {
		resp.Diagnostics.AddError("Unable to check in component", err.Error())
	}
}
