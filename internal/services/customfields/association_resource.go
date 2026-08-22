package customfields

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	customfieldsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/customfields"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &FieldFieldsetAssociationResource{}
	_ resource.ResourceWithImportState = &FieldFieldsetAssociationResource{}
)

// NewFieldFieldsetAssociationResource returns a new
// snipeit_field_fieldset_association resource.
func NewFieldFieldsetAssociationResource() resource.Resource {
	return &FieldFieldsetAssociationResource{}
}

// FieldFieldsetAssociationResource manages the membership of a custom field
// in a fieldset.
type FieldFieldsetAssociationResource struct {
	svc *customfieldsapi.Service
}

// FieldFieldsetAssociationResourceModel describes the resource data model.
type FieldFieldsetAssociationResourceModel struct {
	ID         types.String `tfsdk:"id"`
	FieldID    types.Int64  `tfsdk:"field_id"`
	FieldsetID types.Int64  `tfsdk:"fieldset_id"`
}

func (r *FieldFieldsetAssociationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_field_fieldset_association"
}

func (r *FieldFieldsetAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Associates a custom field (`snipeit_field`) with a fieldset " +
			"(`snipeit_fieldset`). Deleting the association removes the field from the fieldset " +
			"without deleting either object.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Synthetic id in the form `field_id:fieldset_id`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"field_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the custom field. Changing this forces a new association.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"fieldset_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the fieldset. Changing this forces a new association.",
				Required:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
		},
	}
}

func (r *FieldFieldsetAssociationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = customfieldsapi.New(c)
	}
}

// associationExists reports whether the field is currently part of the
// fieldset. The dedicated /fieldsets/{id}/fields endpoint 404s on v8.0.4, so
// membership is read from the fieldset detail response.
func (r *FieldFieldsetAssociationResource) associationExists(ctx context.Context, fieldID, fieldsetID int64) (bool, error) {
	api, err := r.svc.GetFieldset(ctx, fieldsetID)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	for _, f := range api.Fields.Rows {
		if f.Id == fieldID {
			return true, nil
		}
	}
	return false, nil
}

func (r *FieldFieldsetAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data FieldFieldsetAssociationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	fieldID := data.FieldID.ValueInt64()
	fieldsetID := data.FieldsetID.ValueInt64()
	if err := r.svc.AssociateField(ctx, fieldID, fieldsetID); err != nil {
		resp.Diagnostics.AddError("Unable to associate field with fieldset", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d:%d", fieldID, fieldsetID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *FieldFieldsetAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data FieldFieldsetAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	exists, err := r.associationExists(ctx, data.FieldID.ValueInt64(), data.FieldsetID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Unable to read field/fieldset association", err.Error())
		return
	}
	if !exists {
		resp.State.RemoveResource(ctx)
		return
	}
	data.ID = types.StringValue(fmt.Sprintf("%d:%d", data.FieldID.ValueInt64(), data.FieldsetID.ValueInt64()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update never runs: both attributes force replacement.
func (r *FieldFieldsetAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Unexpected update of snipeit_field_fieldset_association",
		"All attributes require replacement; this is a bug in the provider.",
	)
}

func (r *FieldFieldsetAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data FieldFieldsetAssociationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DisassociateField(ctx, data.FieldID.ValueInt64(), data.FieldsetID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to disassociate field from fieldset", err.Error())
	}
}

func (r *FieldFieldsetAssociationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	fieldID, fieldsetID, err := splitFieldAssociationID(req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID",
			fmt.Sprintf("Expected \"field_id:fieldset_id\", got %q.", req.ID))
		return
	}

	data := FieldFieldsetAssociationResourceModel{
		ID:         types.StringValue(fmt.Sprintf("%d:%d", fieldID, fieldsetID)),
		FieldID:    types.Int64Value(fieldID),
		FieldsetID: types.Int64Value(fieldsetID),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// splitFieldAssociationID parses "field_id:fieldset_id" (also accepting "/"
// as separator).
func splitFieldAssociationID(id string) (int64, int64, error) {
	sep := ":"
	if !strings.Contains(id, sep) {
		sep = "/"
	}
	parts := strings.Split(id, sep)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("expected two parts, got %d", len(parts))
	}
	fieldID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	fieldsetID, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, 0, err
	}
	return fieldID, fieldsetID, nil
}
