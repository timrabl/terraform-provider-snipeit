package inventory

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	inventoryapi "github.com/timrabl/terraform-provider-snipeit/internal/api/inventory"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var _ resource.Resource = &AccessoryCheckoutResource{}

// NewAccessoryCheckoutResource returns a new snipeit_accessory_checkout resource.
func NewAccessoryCheckoutResource() resource.Resource {
	return &AccessoryCheckoutResource{}
}

// AccessoryCheckoutResource models the checkout of one accessory unit to a
// user. Create checks out, Delete checks the unit back in via its checkout
// (pivot) row id.
type AccessoryCheckoutResource struct {
	svc *inventoryapi.Service
}

// AccessoryCheckoutResourceModel describes the resource data model.
type AccessoryCheckoutResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	AccessoryID types.Int64  `tfsdk:"accessory_id"`
	UserID      types.Int64  `tfsdk:"user_id"`
	Note        types.String `tfsdk:"note"`
}

func (r *AccessoryCheckoutResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_accessory_checkout"
}

func (r *AccessoryCheckoutResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks one unit of an accessory out to a user. Creating the resource " +
			"performs the checkout, destroying it checks the unit back in. All attributes force " +
			"replacement. Snipe-IT v8 only supports users as accessory checkout targets via the API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Id of the checkout (pivot) row, used for checkin.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"accessory_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the accessory to check out.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the user receiving the accessory.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "Note recorded with the checkout.",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *AccessoryCheckoutResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = inventoryapi.New(c)
	}
}

func (r *AccessoryCheckoutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data AccessoryCheckoutResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	accessoryID := data.AccessoryID.ValueInt64()
	userID := data.UserID.ValueInt64()

	// Snapshot existing pivot ids so the new row can be identified afterwards
	// (the checkout response has a null payload).
	before, err := r.svc.ListAccessoryCheckouts(ctx, accessoryID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list accessory checkouts", err.Error())
		return
	}
	known := make(map[int64]bool, len(before.Rows))
	for _, row := range before.Rows {
		known[row.Id] = true
	}

	body := map[string]any{"assigned_user": userID}
	if !data.Note.IsNull() && !data.Note.IsUnknown() {
		body["note"] = data.Note.ValueString()
	}
	if err := r.svc.CheckoutAccessory(ctx, accessoryID, body); err != nil {
		resp.Diagnostics.AddError("Unable to check out accessory", err.Error())
		return
	}

	after, err := r.svc.ListAccessoryCheckouts(ctx, accessoryID)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list accessory checkouts after checkout", err.Error())
		return
	}
	var pivotID int64
	for _, row := range after.Rows {
		if known[row.Id] || row.AssignedTo == nil || row.AssignedTo.Id != userID {
			continue
		}
		if row.Id > pivotID {
			pivotID = row.Id
		}
	}
	if pivotID == 0 {
		resp.Diagnostics.AddError("Unable to identify accessory checkout",
			"The checkout succeeded but its pivot row could not be found.")
		return
	}

	data.ID = types.Int64Value(pivotID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *AccessoryCheckoutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AccessoryCheckoutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	list, err := r.svc.ListAccessoryCheckouts(ctx, data.AccessoryID.ValueInt64())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read accessory checkout", err.Error())
		return
	}
	for _, row := range list.Rows {
		if row.Id == data.ID.ValueInt64() {
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
			return
		}
	}
	// Checked in outside of Terraform.
	resp.State.RemoveResource(ctx)
}

func (r *AccessoryCheckoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes RequiresReplace; Update is never called.
	resp.Diagnostics.AddError("Unexpected update", "snipeit_accessory_checkout does not support in-place updates.")
}

func (r *AccessoryCheckoutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data AccessoryCheckoutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Checkin addresses the pivot row id, not the accessory id.
	err := r.svc.CheckinAccessory(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) && !isAlreadyCheckedInError(err) {
		resp.Diagnostics.AddError("Unable to check in accessory", err.Error())
	}
}
