package people

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	peopleapi "github.com/timrabl/terraform-provider-snipeit/internal/api/people"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

// UserDataSourceModel describes the user data source data model (no password:
// the API never returns it).
type UserDataSourceModel struct {
	ID           types.Int64  `tfsdk:"id"`
	Username     types.String `tfsdk:"username"`
	FirstName    types.String `tfsdk:"first_name"`
	LastName     types.String `tfsdk:"last_name"`
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

// userDataSourceAttributes returns the computed attribute set shared by the
// snipeit_user and snipeit_user_me data sources.
func userDataSourceAttributes() map[string]dsschema.Attribute {
	return map[string]dsschema.Attribute{
		"first_name":    tfutil.DSString("First name of the user."),
		"last_name":     tfutil.DSString("Last name of the user."),
		"employee_num":  tfutil.DSString("Employee number of the user."),
		"jobtitle":      tfutil.DSString("Job title of the user."),
		"phone":         tfutil.DSString("Phone number of the user."),
		"notes":         tfutil.DSString("Free-form notes."),
		"company_id":    tfutil.DSInt64("Id of the company the user belongs to."),
		"department_id": tfutil.DSInt64("Id of the department the user belongs to."),
		"location_id":   tfutil.DSInt64("Id of the location of the user."),
		"manager_id":    tfutil.DSInt64("Id of the user's manager."),
		"activated":     tfutil.DSBool("Whether the user can log in."),
		"groups": dsschema.SetAttribute{
			MarkdownDescription: "Set of permission group ids the user belongs to.",
			ElementType:         types.Int64Type,
			Computed:            true,
		},
	}
}

// userFromAPI maps a User object into the data source model.
func userFromAPI(ctx context.Context, api *peopleapi.User, m *UserDataSourceModel) error {
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

// NewUserDataSource returns a new snipeit_user data source.
func NewUserDataSource() datasource.DataSource {
	attrs := userDataSourceAttributes()
	attrs["id"] = tfutil.DSID("user")
	attrs["username"] = tfutil.DSLookupString("Login name. Set it to look up by exact username.")
	attrs["email"] = tfutil.DSLookupString("Email address. Set it to look up by exact email.")

	return tfutil.NewLookupDataSource(tfutil.LookupCfg[peopleapi.User, UserDataSourceModel]{
		TypeSuffix: "_user",
		Path:       "/users",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single user by `id`, exact `username` or exact `email`.",
			Attributes:          attrs,
		},
		IDOf: func(m *UserDataSourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[UserDataSourceModel]{
			{
				Attr:  "username",
				Get:   func(m *UserDataSourceModel) types.String { return m.Username },
				Match: func(r tfutil.ListRow) string { return r.Username },
			},
			{
				Attr:  "email",
				Get:   func(m *UserDataSourceModel) types.String { return m.Email },
				Match: func(r tfutil.ListRow) string { return r.Email },
			},
		},
		FromAPI: userFromAPI,
	})
}

// NewGroupDataSource returns a new snipeit_group data source.
func NewGroupDataSource() datasource.DataSource {
	return tfutil.NewLookupDataSource(tfutil.LookupCfg[peopleapi.Group, GroupResourceModel]{
		TypeSuffix: "_group",
		Path:       "/groups",
		Schema: dsschema.Schema{
			MarkdownDescription: "Looks up a single permission group by `id` or exact `name`.",
			Attributes: map[string]dsschema.Attribute{
				"id":    tfutil.DSID("group"),
				"name":  tfutil.DSLookupString("Name of the group. Set it to look up by exact name."),
				"notes": tfutil.DSString("Free-form notes."),
				"permissions": dsschema.MapAttribute{
					MarkdownDescription: "Permission map stored for the group.",
					ElementType:         types.StringType,
					Computed:            true,
				},
			},
		},
		IDOf: func(m *GroupResourceModel) types.Int64 { return m.ID },
		Lookups: []tfutil.LookupField[GroupResourceModel]{{
			Attr:  "name",
			Get:   func(m *GroupResourceModel) types.String { return m.Name },
			Match: func(r tfutil.ListRow) string { return r.Name },
		}},
		FromAPI: func(ctx context.Context, api *peopleapi.Group, m *GroupResourceModel) error {
			m.ID = types.Int64Value(api.Id)
			m.Name = types.StringValue(api.Name)
			m.Notes = tfutil.StateStringPtr(api.Notes)
			// Unlike the resource, the data source always exposes the
			// full stored permission map.
			if len(api.Permissions) == 0 {
				m.Permissions = types.MapNull(types.StringType)
				return nil
			}
			permMap, diags := types.MapValueFrom(ctx, types.StringType, api.Permissions)
			if diags.HasError() {
				return fmt.Errorf("building permissions map: %v", diags)
			}
			m.Permissions = permMap
			return nil
		},
	})
}

var _ datasource.DataSource = &UserMeDataSource{}

// NewUserMeDataSource returns a new snipeit_user_me data source.
func NewUserMeDataSource() datasource.DataSource {
	return &UserMeDataSource{}
}

// UserMeDataSource returns the user that owns the configured API token.
type UserMeDataSource struct {
	svc *peopleapi.Service
}

func (d *UserMeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_me"
}

func (d *UserMeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := userDataSourceAttributes()
	attrs["id"] = tfutil.DSInt64("Numeric id of the user.")
	attrs["username"] = tfutil.DSString("Login name of the user.")
	attrs["email"] = tfutil.DSString("Email address of the user.")

	resp.Schema = dsschema.Schema{
		MarkdownDescription: "Returns the user that owns the API token used by the provider (`GET /users/me`).",
		Attributes:          attrs,
	}
}

func (d *UserMeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		d.svc = peopleapi.New(c)
	}
}

func (d *UserMeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data UserDataSourceModel

	api, err := d.svc.Me(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read current user", err.Error())
		return
	}
	if err := userFromAPI(ctx, api, &data); err != nil {
		resp.Diagnostics.AddError("Unable to map current user", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
