package inventory

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	inventoryapi "github.com/timrabl/terraform-provider-snipeit/internal/api/inventory"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &HardwareCheckoutResource{}
	_ resource.ResourceWithImportState = &HardwareCheckoutResource{}
)

// NewHardwareCheckoutResource returns a new snipeit_hardware_checkout resource.
func NewHardwareCheckoutResource() resource.Resource {
	return &HardwareCheckoutResource{}
}

// HardwareCheckoutResource models the checkout of an asset to a user, asset or
// location. Create checks the asset out, Delete checks it back in. There is no
// in-place update; every change replaces the resource (checkin + checkout).
type HardwareCheckoutResource struct {
	svc *inventoryapi.Service
}

// HardwareCheckoutResourceModel describes the resource data model.
type HardwareCheckoutResourceModel struct {
	ID             types.Int64  `tfsdk:"id"`
	AssetID        types.Int64  `tfsdk:"asset_id"`
	CheckoutToType types.String `tfsdk:"checkout_to_type"`
	AssignedID     types.Int64  `tfsdk:"assigned_id"`
	Note           types.String `tfsdk:"note"`
}

// isAlreadyCheckedInError reports whether an API error indicates the target
// was already checked in ("That asset is already checked in."), which checkout
// resources treat as a successful delete.
func isAlreadyCheckedInError(err error) bool {
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	return strings.Contains(strings.ToLower(apiErr.Messages), "already checked in")
}

func (r *HardwareCheckoutResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_hardware_checkout"
}

func (r *HardwareCheckoutResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Checks an asset out to a user, another asset, or a location. " +
			"Creating the resource performs the checkout, destroying it checks the asset back in. " +
			"All attributes force replacement (checkin followed by a new checkout).",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Id of the checked-out asset (same as `asset_id`).",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"asset_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the asset to check out.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"checkout_to_type": schema.StringAttribute{
				MarkdownDescription: "Target type: `user`, `asset` or `location`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("user", "asset", "location"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"assigned_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the target user, asset or location.",
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

func (r *HardwareCheckoutResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = inventoryapi.New(c)
	}
}

func (r *HardwareCheckoutResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data HardwareCheckoutResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	targetType := data.CheckoutToType.ValueString()
	body := map[string]any{"checkout_to_type": targetType}
	switch targetType {
	case "user":
		body["assigned_user"] = data.AssignedID.ValueInt64()
	case "asset":
		body["assigned_asset"] = data.AssignedID.ValueInt64()
	case "location":
		body["assigned_location"] = data.AssignedID.ValueInt64()
	}
	if !data.Note.IsNull() && !data.Note.IsUnknown() {
		body["note"] = data.Note.ValueString()
	}

	assetID := data.AssetID.ValueInt64()
	if err := r.svc.CheckoutHardware(ctx, assetID, body); err != nil {
		resp.Diagnostics.AddError("Unable to check out asset", err.Error())
		return
	}
	data.ID = types.Int64Value(assetID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HardwareCheckoutResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data HardwareCheckoutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	api, err := r.svc.GetHardwareAssignment(ctx, data.AssetID.ValueInt64())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read asset assignment", err.Error())
		return
	}
	// Checked in outside of Terraform (or reassigned): drop from state so the
	// next apply re-creates the checkout.
	if api.AssignedTo == nil ||
		api.AssignedTo.Id != data.AssignedID.ValueInt64() ||
		api.AssignedTo.Type != data.CheckoutToType.ValueString() {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *HardwareCheckoutResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All attributes RequiresReplace; Update is never called.
	resp.Diagnostics.AddError("Unexpected update", "snipeit_hardware_checkout does not support in-place updates.")
}

func (r *HardwareCheckoutResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data HardwareCheckoutResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.CheckinHardware(ctx, data.AssetID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) && !isAlreadyCheckedInError(err) {
		resp.Diagnostics.AddError("Unable to check in asset", err.Error())
	}
}

func (r *HardwareCheckoutResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.AddError("Import not supported",
		"snipeit_hardware_checkout cannot be imported; re-create the checkout via Terraform instead.")
}
