// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// SPDX-License-Identifier: MPL-2.0

// Package provider implements the Terraform provider for Snipe-IT.
package provider

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/timrabl/terraform-provider-snipeit/internal/client"
	"github.com/timrabl/terraform-provider-snipeit/internal/services/organization"
)

// Ensure SnipeITProvider satisfies various provider interfaces.
var _ provider.Provider = &SnipeITProvider{}
var _ provider.ProviderWithFunctions = &SnipeITProvider{}

// SnipeITProvider defines the provider implementation.
type SnipeITProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// SnipeITProviderModel describes the provider data model.
type SnipeITProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Token    types.String `tfsdk:"token"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (p *SnipeITProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "snipeit"
	resp.Version = p.version
}

func (p *SnipeITProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Interact with a [Snipe-IT](https://snipeitapp.com/) asset management instance " +
			"via its REST API.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "Base URL of the Snipe-IT instance, e.g. `https://snipeit.example.com`. " +
					"May also be provided via the `SNIPEIT_URL` environment variable.",
				Optional: true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "Personal API token (Bearer token) generated in Snipe-IT under " +
					"*Manage API Keys*. May also be provided via the `SNIPEIT_TOKEN` environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Skip TLS certificate verification. Only use against development " +
					"instances with self-signed certificates. May also be set via `SNIPEIT_INSECURE=true`.",
				Optional: true,
			},
		},
	}
}

func (p *SnipeITProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data SnipeITProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	url := os.Getenv("SNIPEIT_URL")
	if !data.URL.IsNull() {
		url = data.URL.ValueString()
	}
	token := os.Getenv("SNIPEIT_TOKEN")
	if !data.Token.IsNull() {
		token = data.Token.ValueString()
	}
	insecure := os.Getenv("SNIPEIT_INSECURE") == "true"
	if !data.Insecure.IsNull() {
		insecure = data.Insecure.ValueBool()
	}

	if url == "" {
		resp.Diagnostics.AddAttributeError(path.Root("url"), "Missing Snipe-IT URL",
			"Set the provider attribute \"url\" or the SNIPEIT_URL environment variable.")
	}
	if token == "" {
		resp.Diagnostics.AddAttributeError(path.Root("token"), "Missing Snipe-IT API token",
			"Set the provider attribute \"token\" or the SNIPEIT_TOKEN environment variable.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	c, err := client.New(client.Config{
		URL:       url,
		Token:     token,
		Insecure:  insecure,
		UserAgent: fmt.Sprintf("terraform-provider-snipeit/%s", p.version),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Snipe-IT API client", err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *SnipeITProvider) Resources(ctx context.Context) []func() resource.Resource {
	// Every resource lives in a domain package under internal/services.
	rs := organization.Resources()
	return rs
}

func (p *SnipeITProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	ds := organization.DataSources()
	return ds
}

func (p *SnipeITProvider) Functions(ctx context.Context) []func() function.Function {
	return nil
}

// New returns a provider factory for the given version string.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &SnipeITProvider{
			version: version,
		}
	}
}
