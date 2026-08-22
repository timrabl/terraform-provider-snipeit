package people

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	peopleapi "github.com/timrabl/terraform-provider-snipeit/internal/api/people"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &UserResource{}
	_ resource.ResourceWithImportState = &UserResource{}
)

// NewUserResource returns a new snipeit_user resource.
func NewUserResource() resource.Resource {
	return &UserResource{}
}

// UserResource manages a Snipe-IT user.
type UserResource struct {
	svc *peopleapi.Service
}

// UserResourceModel describes the resource data model.
type UserResourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Username     types.String `tfsdk:"username"`
	FirstName    types.String `tfsdk:"first_name"`
	LastName     types.String `tfsdk:"last_name"`
	Password     types.String `tfsdk:"password"`
	Email        types.String `tfsdk:"email"`
	EmployeeNum  types.String `tfsdk:"employee_num"`
	Jobtitle     types.String `tfsdk:"jobtitle"`
	Phone        types.String `tfsdk:"phone"`
	Notes        types.String `tfsdk:"notes"`
	CompanyID    types.Int64  `tfsdk:"company_id"`
	DepartmentID types.Int64  `tfsdk:"department_id"`
	LocationID   types.Int64  `tfsdk:"location_id"`
	ManagerID    types.Int64  `tfsdk:"manager_id"`
	Activated    types.Bool   `tfsdk:"activated"`
	Groups       types.Set    `tfsdk:"groups"`
}

func (r *UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a user in Snipe-IT.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the user.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Login name of the user. Must be unique.",
				Required:            true,
			},
			"first_name": schema.StringAttribute{
				MarkdownDescription: "First name of the user.",
				Required:            true,
			},
			"last_name": schema.StringAttribute{
				MarkdownDescription: "Last name of the user.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password of the user. Write-only: the API never returns it, so the " +
					"configured value is kept in state as-is. Required by the API when creating a user. " +
					"After `terraform import` this attribute is null and only re-set when changed in " +
					"configuration.",
				Optional:  true,
				Sensitive: true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Email address of the user.",
				Optional:            true,
			},
			"employee_num": schema.StringAttribute{
				MarkdownDescription: "Employee number of the user.",
				Optional:            true,
			},
			"jobtitle": schema.StringAttribute{
				MarkdownDescription: "Job title of the user.",
				Optional:            true,
			},
			"phone": schema.StringAttribute{
				MarkdownDescription: "Phone number of the user.",
				Optional:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Free-form notes.",
				Optional:            true,
			},
			"company_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the company the user belongs to.",
				Optional:            true,
			},
			"department_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the department the user belongs to.",
				Optional:            true,
			},
			"location_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the location of the user.",
				Optional:            true,
			},
			"manager_id": schema.Int64Attribute{
				MarkdownDescription: "Id of the user's manager.",
				Optional:            true,
			},
			"activated": schema.BoolAttribute{
				MarkdownDescription: "Whether the user can log in.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"groups": schema.SetAttribute{
				MarkdownDescription: "Set of permission group ids the user belongs to.",
				ElementType:         types.Int64Type,
				Optional:            true,
			},
		},
	}
}

func (r *UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = peopleapi.New(c)
	}
}

func (m *UserResourceModel) toBody(ctx context.Context) (map[string]any, error) {
	body := map[string]any{
		"username":   m.Username.ValueString(),
		"first_name": m.FirstName.ValueString(),
	}
	tfutil.BodyString(body, "last_name", m.LastName)
	tfutil.BodyString(body, "email", m.Email)
	tfutil.BodyString(body, "employee_num", m.EmployeeNum)
	tfutil.BodyString(body, "jobtitle", m.Jobtitle)
	tfutil.BodyString(body, "phone", m.Phone)
	tfutil.BodyString(body, "notes", m.Notes)
	tfutil.BodyNullableInt(body, "company_id", m.CompanyID)
	tfutil.BodyNullableInt(body, "department_id", m.DepartmentID)
	tfutil.BodyNullableInt(body, "location_id", m.LocationID)
	tfutil.BodyNullableInt(body, "manager_id", m.ManagerID)
	tfutil.BodyOptBool(body, "activated", m.Activated)
	if !m.Password.IsNull() && !m.Password.IsUnknown() {
		body["password"] = m.Password.ValueString()
		body["password_confirmation"] = m.Password.ValueString()
	}
	if !m.Groups.IsUnknown() {
		if m.Groups.IsNull() {
			body["groups"] = nil
		} else {
			var ids []int64
			if diags := m.Groups.ElementsAs(ctx, &ids, false); diags.HasError() {
				return nil, fmt.Errorf("decoding groups: %v", diags)
			}
			body["groups"] = ids
		}
	}
	return body, nil
}

// fromAPI maps the API object into state. The password is intentionally left
// untouched: the API never returns it.
func (m *UserResourceModel) fromAPI(ctx context.Context, api *peopleapi.User) error {
	m.ID = types.Int64Value(api.Id)
	m.Username = types.StringValue(api.Username)
	m.FirstName = types.StringValue(api.FirstName)
	m.LastName = tfutil.StateStringPtr(api.LastName)
	m.Email = tfutil.StateStringPtr(api.Email)
	m.EmployeeNum = tfutil.StateStringPtr(api.EmployeeNum)
	m.Jobtitle = tfutil.StateStringPtr(api.Jobtitle)
	m.Phone = tfutil.StateStringPtr(api.Phone)
	m.Notes = tfutil.StateStringPtr(api.Notes)
	m.CompanyID = tfutil.StateRefID(api.Company)
	m.DepartmentID = tfutil.StateRefID(api.Department)
	m.LocationID = tfutil.StateRefID(api.Location)
	m.ManagerID = tfutil.StateRefID(api.Manager)
	m.Activated = types.BoolValue(api.Activated)

	if api.Groups == nil || len(api.Groups.Rows) == 0 {
		m.Groups = types.SetNull(types.Int64Type)
		return nil
	}
	ids := make([]int64, 0, len(api.Groups.Rows))
	for _, g := range api.Groups.Rows {
		ids = append(ids, g.ID)
	}
	set, diags := types.SetValueFrom(ctx, types.Int64Type, ids)
	if diags.HasError() {
		return fmt.Errorf("building groups set: %v", diags)
	}
	m.Groups = set
	return nil
}

func (r *UserResource) read(ctx context.Context, id int64, data *UserResourceModel) error {
	api, err := r.svc.GetUser(ctx, id)
	if err != nil {
		return err
	}
	return data.fromAPI(ctx, api)
}

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := data.toBody(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build user request", err.Error())
		return
	}
	id, err := r.svc.CreateUser(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create user", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read user after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read user", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := data.toBody(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build user request", err.Error())
		return
	}
	id := data.ID.ValueInt64()
	if err := r.svc.UpdateUser(ctx, id, body); err != nil {
		resp.Diagnostics.AddError("Unable to update user", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read user after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteUser(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete user",
			err.Error()+" (users with assigned assets, accessories or licenses cannot be deleted)")
	}
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
