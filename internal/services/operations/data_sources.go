package operations

import (
	"context"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	operationsapi "github.com/timrabl/terraform-provider-snipeit/internal/api/operations"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ datasource.DataSource = &ActivityReportsDataSource{}
	_ datasource.DataSource = &HardwareAuditDataSource{}
)

// NewActivityReportsDataSource returns a new snipeit_activity_reports data source.
func NewActivityReportsDataSource() datasource.DataSource {
	return &ActivityReportsDataSource{}
}

// ActivityReportsDataSource reads the activity report (audit log).
type ActivityReportsDataSource struct {
	svc *operationsapi.Service
}

// ActivityReportRowModel is one activity log row.
type ActivityReportRowModel struct {
	ID         types.Int64  `tfsdk:"id"`
	ActionType types.String `tfsdk:"action_type"`
	ItemType   types.String `tfsdk:"item_type"`
	ItemID     types.Int64  `tfsdk:"item_id"`
	ItemName   types.String `tfsdk:"item_name"`
	TargetType types.String `tfsdk:"target_type"`
	TargetID   types.Int64  `tfsdk:"target_id"`
	TargetName types.String `tfsdk:"target_name"`
	AdminName  types.String `tfsdk:"admin_name"`
	Note       types.String `tfsdk:"note"`
	ActionDate types.String `tfsdk:"action_date"`
}

// ActivityReportsDataSourceModel describes the data source data model.
type ActivityReportsDataSourceModel struct {
	Limit      types.Int64              `tfsdk:"limit"`
	Offset     types.Int64              `tfsdk:"offset"`
	ActionType types.String             `tfsdk:"action_type"`
	ItemType   types.String             `tfsdk:"item_type"`
	ItemID     types.Int64              `tfsdk:"item_id"`
	Total      types.Int64              `tfsdk:"total"`
	Rows       []ActivityReportRowModel `tfsdk:"rows"`
}

func (d *ActivityReportsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_activity_reports"
}

func (d *ActivityReportsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Reads rows from the Snipe-IT activity report (audit log), newest first.",
		Attributes: map[string]schema.Attribute{
			"limit": schema.Int64Attribute{
				MarkdownDescription: "Maximum number of rows to return (default 50).",
				Optional:            true,
			},
			"offset": schema.Int64Attribute{
				MarkdownDescription: "Offset into the result set.",
				Optional:            true,
			},
			"action_type": schema.StringAttribute{
				MarkdownDescription: "Filter by action type, e.g. `checkout`, `checkin from`, `update`, `create`.",
				Optional:            true,
			},
			"item_type": schema.StringAttribute{
				MarkdownDescription: "Filter by item type, e.g. `asset`, `user`, `license`, `accessory`.",
				Optional:            true,
			},
			"item_id": schema.Int64Attribute{
				MarkdownDescription: "Filter by item id (requires `item_type`).",
				Optional:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of matching rows on the server.",
				Computed:            true,
			},
			"rows": schema.ListNestedAttribute{
				MarkdownDescription: "Activity log rows.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":          schema.Int64Attribute{MarkdownDescription: "Row id.", Computed: true},
						"action_type": schema.StringAttribute{MarkdownDescription: "Action performed.", Computed: true},
						"item_type":   schema.StringAttribute{MarkdownDescription: "Type of the item acted on.", Computed: true},
						"item_id":     schema.Int64Attribute{MarkdownDescription: "Id of the item acted on.", Computed: true},
						"item_name":   schema.StringAttribute{MarkdownDescription: "Name of the item acted on.", Computed: true},
						"target_type": schema.StringAttribute{MarkdownDescription: "Type of the checkout target.", Computed: true},
						"target_id":   schema.Int64Attribute{MarkdownDescription: "Id of the checkout target.", Computed: true},
						"target_name": schema.StringAttribute{MarkdownDescription: "Name of the checkout target.", Computed: true},
						"admin_name":  schema.StringAttribute{MarkdownDescription: "Name of the acting user.", Computed: true},
						"note":        schema.StringAttribute{MarkdownDescription: "Note attached to the action.", Computed: true},
						"action_date": schema.StringAttribute{MarkdownDescription: "Timestamp of the action.", Computed: true},
					},
				},
			},
		},
	}
}

func (d *ActivityReportsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		d.svc = operationsapi.New(c)
	}
}

func (d *ActivityReportsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ActivityReportsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	query := url.Values{}
	if !data.Limit.IsNull() {
		query.Set("limit", strconv.FormatInt(data.Limit.ValueInt64(), 10))
	}
	if !data.Offset.IsNull() {
		query.Set("offset", strconv.FormatInt(data.Offset.ValueInt64(), 10))
	}
	if !data.ActionType.IsNull() {
		query.Set("action_type", data.ActionType.ValueString())
	}
	if !data.ItemType.IsNull() {
		query.Set("item_type", data.ItemType.ValueString())
	}
	if !data.ItemID.IsNull() {
		query.Set("item_id", strconv.FormatInt(data.ItemID.ValueInt64(), 10))
	}

	list, err := d.svc.ActivityReport(ctx, query)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read activity report", err.Error())
		return
	}

	data.Total = types.Int64Value(list.Total)
	data.Rows = make([]ActivityReportRowModel, 0, len(list.Rows))
	for _, row := range list.Rows {
		out := ActivityReportRowModel{
			ID:         types.Int64Value(row.Id),
			ActionType: types.StringValue(row.ActionType),
			Note:       tfutil.StateStringPtr(row.Note),
		}
		if row.Item != nil {
			out.ItemType = types.StringValue(row.Item.Type)
			out.ItemID = types.Int64Value(row.Item.Id)
			out.ItemName = tfutil.StateString(row.Item.Name)
		}
		if row.Target != nil {
			out.TargetType = types.StringValue(row.Target.Type)
			out.TargetID = types.Int64Value(row.Target.Id)
			out.TargetName = tfutil.StateString(row.Target.Name)
		}
		if row.Admin != nil {
			out.AdminName = tfutil.StateString(row.Admin.Name)
		}
		if row.ActionDate != nil {
			out.ActionDate = types.StringValue(row.ActionDate.DateTime)
		}
		data.Rows = append(data.Rows, out)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// NewHardwareAuditDueDataSource returns the snipeit_hardware_audit_due data
// source, listing assets due for audit.
func NewHardwareAuditDueDataSource() datasource.DataSource {
	return &HardwareAuditDataSource{suffix: "_hardware_audit_due", overdue: false,
		desc: "Lists assets that are due (or nearly due) for audit."}
}

// NewHardwareAuditOverdueDataSource returns the snipeit_hardware_audit_overdue
// data source, listing assets overdue for audit.
func NewHardwareAuditOverdueDataSource() datasource.DataSource {
	return &HardwareAuditDataSource{suffix: "_hardware_audit_overdue", overdue: true,
		desc: "Lists assets that are overdue for audit."}
}

// HardwareAuditDataSource lists assets due or overdue for audit.
type HardwareAuditDataSource struct {
	svc     *operationsapi.Service
	suffix  string
	overdue bool
	desc    string
}

// HardwareAuditAssetModel is one asset row in an audit list.
type HardwareAuditAssetModel struct {
	ID       types.Int64  `tfsdk:"id"`
	AssetTag types.String `tfsdk:"asset_tag"`
	Name     types.String `tfsdk:"name"`
}

// HardwareAuditDataSourceModel describes the data source data model.
type HardwareAuditDataSourceModel struct {
	Total  types.Int64               `tfsdk:"total"`
	Assets []HardwareAuditAssetModel `tfsdk:"assets"`
}

func (d *HardwareAuditDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + d.suffix
}

func (d *HardwareAuditDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: d.desc,
		Attributes: map[string]schema.Attribute{
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of matching assets.",
				Computed:            true,
			},
			"assets": schema.ListNestedAttribute{
				MarkdownDescription: "Matching assets.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":        schema.Int64Attribute{MarkdownDescription: "Asset id.", Computed: true},
						"asset_tag": schema.StringAttribute{MarkdownDescription: "Asset tag.", Computed: true},
						"name":      schema.StringAttribute{MarkdownDescription: "Asset name.", Computed: true},
					},
				},
			},
		},
	}
}

func (d *HardwareAuditDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		d.svc = operationsapi.New(c)
	}
}

func (d *HardwareAuditDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var list *operationsapi.AuditList
	var err error
	if d.overdue {
		list, err = d.svc.AuditOverdue(ctx)
	} else {
		list, err = d.svc.AuditDue(ctx)
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read hardware audit list", err.Error())
		return
	}

	data := HardwareAuditDataSourceModel{
		Total:  types.Int64Value(list.Total),
		Assets: make([]HardwareAuditAssetModel, 0, len(list.Rows)),
	}
	for _, row := range list.Rows {
		data.Assets = append(data.Assets, HardwareAuditAssetModel{
			ID:       types.Int64Value(row.Id),
			AssetTag: types.StringValue(row.AssetTag),
			Name:     tfutil.StateStringPtr(row.Name),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
