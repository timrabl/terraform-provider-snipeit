package organization

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	organizationapi "github.com/timrabl/terraform-provider-snipeit/internal/api/organization"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &SupplierResource{}
	_ resource.ResourceWithImportState = &SupplierResource{}
)

// NewSupplierResource returns a new snipeit_supplier resource.
func NewSupplierResource() resource.Resource {
	return &SupplierResource{}
}

// SupplierResource manages a Snipe-IT supplier.
type SupplierResource struct {
	svc *organizationapi.Service
}

// SupplierResourceModel describes the resource data model.
type SupplierResourceModel struct {
	ID       types.Int64  `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Address  types.String `tfsdk:"address"`
	Address2 types.String `tfsdk:"address2"`
	City     types.String `tfsdk:"city"`
	State    types.String `tfsdk:"state"`
	Country  types.String `tfsdk:"country"`
	Zip      types.String `tfsdk:"zip"`
	Phone    types.String `tfsdk:"phone"`
	Fax      types.String `tfsdk:"fax"`
	Email    types.String `tfsdk:"email"`
	Contact  types.String `tfsdk:"contact"`
	URL      types.String `tfsdk:"url"`
	Notes    types.String `tfsdk:"notes"`
}

func (r *SupplierResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_supplier"
}

func (r *SupplierResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	optString := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{MarkdownDescription: desc, Optional: true}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a supplier in Snipe-IT.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the supplier.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the supplier. Must be unique.",
				Required:            true,
			},
			"address":  optString("Street address of the supplier."),
			"address2": optString("Street address of the supplier, line 2."),
			"city":     optString("City of the supplier."),
			"state":    optString("State/province of the supplier."),
			"country":  optString("Country of the supplier (two-letter code recommended)."),
			"zip":      optString("Postal code of the supplier."),
			"phone":    optString("Phone number of the supplier."),
			"fax":      optString("Fax number of the supplier."),
			"email":    optString("Email address of the supplier."),
			"contact":  optString("Contact person at the supplier."),
			"url":      optString("Website of the supplier."),
			"notes":    optString("Free-form notes."),
		},
	}
}

func (r *SupplierResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = organizationapi.New(c)
	}
}

func (m *SupplierResourceModel) toBody() map[string]any {
	body := map[string]any{"name": m.Name.ValueString()}
	tfutil.BodyString(body, "address", m.Address)
	tfutil.BodyString(body, "address2", m.Address2)
	tfutil.BodyString(body, "city", m.City)
	tfutil.BodyString(body, "state", m.State)
	tfutil.BodyString(body, "country", m.Country)
	tfutil.BodyString(body, "zip", m.Zip)
	tfutil.BodyString(body, "phone", m.Phone)
	tfutil.BodyString(body, "fax", m.Fax)
	tfutil.BodyString(body, "email", m.Email)
	tfutil.BodyString(body, "contact", m.Contact)
	tfutil.BodyString(body, "url", m.URL)
	tfutil.BodyString(body, "notes", m.Notes)
	return body
}

func (m *SupplierResourceModel) fromAPI(api *organizationapi.Supplier) {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Address = tfutil.StateStringPtr(api.Address)
	m.Address2 = tfutil.StateStringPtr(api.Address2)
	m.City = tfutil.StateStringPtr(api.City)
	m.State = tfutil.StateStringPtr(api.State)
	m.Country = tfutil.StateStringPtr(api.Country)
	m.Zip = tfutil.StateStringPtr(api.Zip)
	m.Phone = tfutil.StateStringPtr(api.Phone)
	m.Fax = tfutil.StateStringPtr(api.Fax)
	m.Email = tfutil.StateStringPtr(api.Email)
	m.Contact = tfutil.StateStringPtr(api.Contact)
	m.URL = tfutil.StateStringPtr(api.Url)
	m.Notes = tfutil.StateStringPtr(api.Notes)
}

func (r *SupplierResource) read(ctx context.Context, id int64, data *SupplierResourceModel) error {
	api, err := r.svc.GetSupplier(ctx, id)
	if err != nil {
		return err
	}
	data.fromAPI(api)
	return nil
}

func (r *SupplierResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SupplierResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, err := r.svc.CreateSupplier(ctx, data.toBody())
	if err != nil {
		resp.Diagnostics.AddError("Unable to create supplier", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read supplier after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SupplierResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SupplierResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read supplier", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SupplierResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SupplierResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := data.ID.ValueInt64()
	if err := r.svc.UpdateSupplier(ctx, id, data.toBody()); err != nil {
		resp.Diagnostics.AddError("Unable to update supplier", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read supplier after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SupplierResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SupplierResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteSupplier(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete supplier", err.Error())
	}
}

func (r *SupplierResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
