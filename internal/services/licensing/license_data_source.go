package licensing

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	licensingapi "github.com/timrabl/terraform-provider-snipeit/internal/api/licensing"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var _ datasource.DataSource = &LicenseDataSource{}

// NewLicenseDataSource returns a new snipeit_license data source.
func NewLicenseDataSource() datasource.DataSource {
	return &LicenseDataSource{}
}

// LicenseDataSource looks up a single license by id or exact name.
type LicenseDataSource struct {
	svc *licensingapi.Service
}

// LicenseDataSourceModel describes the data source data model.
type LicenseDataSourceModel struct {
	ID              types.Int64  `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Seats           types.Int64  `tfsdk:"seats"`
	FreeSeatsCount  types.Int64  `tfsdk:"free_seats_count"`
	CategoryID      types.Int64  `tfsdk:"category_id"`
	CompanyID       types.Int64  `tfsdk:"company_id"`
	ManufacturerID  types.Int64  `tfsdk:"manufacturer_id"`
	SupplierID      types.Int64  `tfsdk:"supplier_id"`
	OrderNumber     types.String `tfsdk:"order_number"`
	PurchaseOrder   types.String `tfsdk:"purchase_order"`
	PurchaseDate    types.String `tfsdk:"purchase_date"`
	ExpirationDate  types.String `tfsdk:"expiration_date"`
	TerminationDate types.String `tfsdk:"termination_date"`
	LicenseName     types.String `tfsdk:"license_name"`
	LicenseEmail    types.String `tfsdk:"license_email"`
	Serial          types.String `tfsdk:"serial"`
	Reassignable    types.Bool   `tfsdk:"reassignable"`
	Maintained      types.Bool   `tfsdk:"maintained"`
	Notes           types.String `tfsdk:"notes"`
}

func (d *LicenseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_license"
}

func (d *LicenseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Looks up a single license by `id` or exact `name`. Exactly one of these " +
			"must be set.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the license.",
				Optional:            true,
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Exact name of the license to look up.",
				Optional:            true,
				Computed:            true,
			},
			"seats":            schema.Int64Attribute{MarkdownDescription: "Number of seats.", Computed: true},
			"free_seats_count": schema.Int64Attribute{MarkdownDescription: "Number of unassigned seats.", Computed: true},
			"category_id":      schema.Int64Attribute{MarkdownDescription: "Id of the category.", Computed: true},
			"company_id":       schema.Int64Attribute{MarkdownDescription: "Id of the owning company.", Computed: true},
			"manufacturer_id":  schema.Int64Attribute{MarkdownDescription: "Id of the manufacturer.", Computed: true},
			"supplier_id":      schema.Int64Attribute{MarkdownDescription: "Id of the supplier.", Computed: true},
			"order_number":     schema.StringAttribute{MarkdownDescription: "Order number of the purchase.", Computed: true},
			"purchase_order":   schema.StringAttribute{MarkdownDescription: "Purchase order number.", Computed: true},
			"purchase_date":    schema.StringAttribute{MarkdownDescription: "Purchase date (`YYYY-MM-DD`).", Computed: true},
			"expiration_date":  schema.StringAttribute{MarkdownDescription: "Expiration date (`YYYY-MM-DD`).", Computed: true},
			"termination_date": schema.StringAttribute{MarkdownDescription: "Termination date (`YYYY-MM-DD`).", Computed: true},
			"license_name":     schema.StringAttribute{MarkdownDescription: "Name of the licensee.", Computed: true},
			"license_email":    schema.StringAttribute{MarkdownDescription: "Email of the licensee.", Computed: true},
			"serial":           schema.StringAttribute{MarkdownDescription: "Product key / serial.", Computed: true},
			"reassignable":     schema.BoolAttribute{MarkdownDescription: "Whether seats can be reassigned.", Computed: true},
			"maintained":       schema.BoolAttribute{MarkdownDescription: "Whether the license is maintained.", Computed: true},
			"notes":            schema.StringAttribute{MarkdownDescription: "Free-form notes.", Computed: true},
		},
	}
}

func (d *LicenseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		d.svc = licensingapi.New(c)
	}
}

func (d *LicenseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data LicenseDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var api *licensingapi.License
	switch {
	case !data.ID.IsNull():
		found, err := d.svc.GetLicense(ctx, data.ID.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Unable to look up license", err.Error())
			return
		}
		api = found
	case !data.Name.IsNull():
		name := data.Name.ValueString()
		list, err := d.svc.SearchLicenses(ctx, name)
		if err != nil {
			resp.Diagnostics.AddError("Unable to search licenses", err.Error())
			return
		}
		var matches []licensingapi.License
		for i := range list.Rows {
			if list.Rows[i].Name == name {
				matches = append(matches, list.Rows[i])
			}
		}
		switch len(matches) {
		case 0:
			resp.Diagnostics.AddError("License not found",
				fmt.Sprintf("No license with exact name %q.", name))
			return
		case 1:
			api = &matches[0]
		default:
			resp.Diagnostics.AddError("License name ambiguous",
				fmt.Sprintf("%d licenses share the exact name %q; look up by id instead.", len(matches), name))
			return
		}
	default:
		resp.Diagnostics.AddError("Missing lookup key", "Exactly one of id or name must be set.")
		return
	}

	data.ID = types.Int64Value(api.Id)
	data.Name = types.StringValue(api.Name)
	data.Seats = types.Int64Value(int64(api.Seats))
	data.FreeSeatsCount = types.Int64Value(int64(api.FreeSeatsCount))
	data.CategoryID = types.Int64Value(api.Category.IDOrZero())
	data.CompanyID = tfutil.StateRefID(api.Company)
	data.ManufacturerID = tfutil.StateRefID(api.Manufacturer)
	data.SupplierID = tfutil.StateRefID(api.Supplier)
	data.OrderNumber = tfutil.StateStringPtr(api.OrderNumber)
	data.PurchaseOrder = tfutil.StateStringPtr(api.PurchaseOrder)
	data.PurchaseDate = stateDate(api.PurchaseDate)
	data.ExpirationDate = stateDate(api.ExpirationDate)
	data.TerminationDate = stateDate(api.TerminationDate)
	data.LicenseName = tfutil.StateStringPtr(api.LicenseName)
	data.LicenseEmail = tfutil.StateStringPtr(api.LicenseEmail)
	data.Serial = tfutil.StateStringPtr(api.ProductKey)
	data.Reassignable = types.BoolValue(bool(api.Reassignable))
	data.Maintained = types.BoolValue(bool(api.Maintained))
	data.Notes = tfutil.StateStringPtr(api.Notes)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
