package assets

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	assetsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/assets"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &CategoryResource{}
	_ resource.ResourceWithImportState = &CategoryResource{}
)

// NewCategoryResource returns a new snipeit_category resource.
func NewCategoryResource() resource.Resource {
	return &CategoryResource{}
}

// CategoryResource manages a Snipe-IT category.
type CategoryResource struct {
	svc *assetsapi.Service
}

// CategoryResourceModel describes the resource data model.
type CategoryResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	CategoryType      types.String `tfsdk:"category_type"`
	EULAText          types.String `tfsdk:"eula_text"`
	UseDefaultEULA    types.Bool   `tfsdk:"use_default_eula"`
	RequireAcceptance types.Bool   `tfsdk:"require_acceptance"`
	CheckinEmail      types.Bool   `tfsdk:"checkin_email"`
	Notes             types.String `tfsdk:"notes"`
}

func (r *CategoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_category"
}

func (r *CategoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a category in Snipe-IT. Categories classify assets, accessories, " +
			"consumables, components and licenses.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the category.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the category. Must be unique per category type.",
				Required:            true,
			},
			"category_type": schema.StringAttribute{
				MarkdownDescription: "Type of the category. One of `asset`, `accessory`, `consumable`, " +
					"`component` or `license`. Changing this forces a new category.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf("asset", "accessory", "consumable", "component", "license"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"eula_text": schema.StringAttribute{
				MarkdownDescription: "EULA text (Markdown) shown for items in this category.",
				Optional:            true,
			},
			"use_default_eula": schema.BoolAttribute{
				MarkdownDescription: "Use the instance-wide default EULA instead of `eula_text`.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"require_acceptance": schema.BoolAttribute{
				MarkdownDescription: "Require users to confirm acceptance of items in this category.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"checkin_email": schema.BoolAttribute{
				MarkdownDescription: "Send email to users on checkin/checkout of items in this category.",
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

func (r *CategoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = assetsapi.New(c)
	}
}

func (m *CategoryResourceModel) toBody() map[string]any {
	body := map[string]any{
		"name":          m.Name.ValueString(),
		"category_type": m.CategoryType.ValueString(),
	}
	tfutil.BodyString(body, "eula_text", m.EULAText)
	tfutil.BodyString(body, "notes", m.Notes)
	tfutil.BodyOptBool(body, "use_default_eula", m.UseDefaultEULA)
	tfutil.BodyOptBool(body, "require_acceptance", m.RequireAcceptance)
	tfutil.BodyOptBool(body, "checkin_email", m.CheckinEmail)
	return body
}

func (m *CategoryResourceModel) fromAPI(api *assetsapi.Category) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.CategoryType = types.StringValue(strings.ToLower(api.CategoryType))
	m.Notes = tfutil.StateStringPtr(api.Notes)
	m.UseDefaultEULA = types.BoolValue(bool(api.UseDefaultEula))
	m.RequireAcceptance = types.BoolValue(bool(api.RequireAcceptance))
	m.CheckinEmail = types.BoolValue(bool(api.CheckinEmail))
	// The list/detail serializer does not echo eula_text back; keep the
	// configured value so it round-trips (m.EULAText left untouched).
}

func (r *CategoryResource) read(ctx context.Context, id int64, data *CategoryResourceModel) error {
	api, err := r.svc.GetCategory(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *CategoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CategoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateCategory(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create category", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read category after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CategoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read category", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CategoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CategoryResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateCategory(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update category", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read category after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CategoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CategoryResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteCategory(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete category", err.Error())
	}
}

func (r *CategoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
