// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

package people

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	peopleapi "github.com/timrabl/terraform-provider-snipeit/internal/api/people"
	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/tfutil"
)

var (
	_ resource.Resource                = &GroupResource{}
	_ resource.ResourceWithImportState = &GroupResource{}
)

// NewGroupResource returns a new snipeit_group resource.
func NewGroupResource() resource.Resource {
	return &GroupResource{}
}

// GroupResource manages a Snipe-IT permission group.
type GroupResource struct {
	svc *peopleapi.Service
}

// GroupResourceModel describes the resource data model.
type GroupResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Notes       types.String `tfsdk:"notes"`
	Permissions types.Map    `tfsdk:"permissions"`
}

func (r *GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

func (r *GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a permission group in Snipe-IT.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				MarkdownDescription: "Numeric id of the group.",
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the group. Must be unique.",
				Required:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Free-form notes.",
				Optional:            true,
			},
			"permissions": schema.MapAttribute{
				MarkdownDescription: "Permission map of the group, e.g. `{ \"assets.view\" = \"1\" }`. " +
					"Values are `\"1\"` (allow), `\"0\"` (inherit/deny) or `\"-1\"` (explicit deny). " +
					"Snipe-IT stores exactly the submitted map and an update replaces it entirely, " +
					"so this attribute round-trips without drift. After `terraform import` the " +
					"attribute is null; the next apply with a configured map takes it over.",
				ElementType: types.StringType,
				Optional:    true,
			},
		},
	}
}

func (r *GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if c := tfutil.ClientFromProviderData(req.ProviderData, &resp.Diagnostics); c != nil {
		r.svc = peopleapi.New(c)
	}
}

func (m *GroupResourceModel) toBody(ctx context.Context) (map[string]any, error) {
	body := map[string]any{"name": m.Name.ValueString()}
	tfutil.BodyString(body, "notes", m.Notes)
	if !m.Permissions.IsNull() && !m.Permissions.IsUnknown() {
		perms := map[string]string{}
		if diags := m.Permissions.ElementsAs(ctx, &perms, false); diags.HasError() {
			return nil, fmt.Errorf("decoding permissions: %v", diags)
		}
		body["permissions"] = perms
	}
	return body, nil
}

func (m *GroupResourceModel) fromAPI(ctx context.Context, api *peopleapi.Group) error {
	m.ID = types.Int64Value(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Notes = tfutil.StateStringPtrKeep(api.Notes, m.Notes)
	// When permissions are not configured, Snipe-IT generates a full default
	// map ("0" for every known permission). Reflecting that map into state
	// would drift against a null config, so an unconfigured map stays null.
	// Consequence for import: permissions start null and are taken over on
	// the next apply with a configured map.
	if m.Permissions.IsNull() || m.Permissions.IsUnknown() || len(api.Permissions) == 0 {
		m.Permissions = types.MapNull(types.StringType)
		return nil
	}
	// api.Permissions values are client.FlexString (tolerant of the string
	// form on <= 8.0 and the numeric form on 8.4+); normalize to plain strings.
	// Snipe-IT 8.4+ echoes back the FULL permission map for a partial config,
	// so reflect only the keys the user actually configured — otherwise the
	// extra server-side keys read as "inconsistent result after apply".
	configured := m.Permissions.Elements()
	perms := make(map[string]string, len(configured))
	for k, v := range api.Permissions {
		if _, ok := configured[k]; ok {
			perms[k] = string(v)
		}
	}
	permMap, diags := types.MapValueFrom(ctx, types.StringType, perms)
	if diags.HasError() {
		return fmt.Errorf("building permissions map: %v", diags)
	}
	m.Permissions = permMap
	return nil
}

func (r *GroupResource) read(ctx context.Context, id int64, data *GroupResourceModel) error {
	api, err := r.svc.GetGroup(ctx, id)
	if err != nil {
		return err
	}
	return data.fromAPI(ctx, api)
}

func (r *GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := data.toBody(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build group request", err.Error())
		return
	}
	id, err := r.svc.CreateGroup(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create group", err.Error())
		return
	}
	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read group after create", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.read(ctx, data.ID.ValueInt64(), &data); err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Unable to read group", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body, err := data.toBody(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to build group request", err.Error())
		return
	}

	// Snipe-IT v8.0.x bug: group updates that include a permissions map
	// return HTTP 500 "Server Error" even though the update is applied.
	// On error, re-read and verify the server matches the desired state
	// before surfacing anything.
	id := data.ID.ValueInt64()
	updateErr := r.svc.UpdateGroup(ctx, id, body)
	if updateErr != nil && !r.updateApplied(ctx, id, &data, body) {
		resp.Diagnostics.AddError("Unable to update group", updateErr.Error())
		return
	}

	if err := r.read(ctx, id, &data); err != nil {
		resp.Diagnostics.AddError("Unable to read group after update", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// updateApplied reports whether the server state matches the intended update
// body despite an error response (see the v8.0.x bug note above).
func (r *GroupResource) updateApplied(ctx context.Context, id int64, data *GroupResourceModel, body map[string]any) bool {
	api, err := r.svc.GetGroup(ctx, id)
	if err != nil {
		return false
	}
	if api.Name != data.Name.ValueString() {
		return false
	}
	wantPerms, ok := body["permissions"].(map[string]string)
	if !ok {
		return true
	}
	// api.Permissions values are FlexString (tolerant across versions);
	// compare as plain strings.
	gotPerms := make(map[string]string, len(api.Permissions))
	for k, v := range api.Permissions {
		gotPerms[k] = string(v)
	}
	return maps.Equal(gotPerms, wantPerms)
}

func (r *GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data GroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.svc.DeleteGroup(ctx, data.ID.ValueInt64())
	if err != nil && !errors.Is(err, client.ErrNotFound) {
		resp.Diagnostics.AddError("Unable to delete group", err.Error())
	}
}

func (r *GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tfutil.ImportNumericID(ctx, req, resp)
}
