// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mattn/go-mastodon"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &AccountDataSource{}

func NewAccountDataSource() datasource.DataSource {
	return &AccountDataSource{}
}

// AccountDataSource defines the data source implementation.
type AccountDataSource struct {
	client *mastodon.Client
}

// AccountDataSourceModel describes the data source data model.
type AccountDataSourceModel struct {
	Username    types.String `tfsdk:"username"`
	Id          types.String `tfsdk:"id"`
	DisplayName types.String `tfsdk:"display_name"`
	Note        types.String `tfsdk:"note"`
	Locked      types.Bool   `tfsdk:"locked"`
	Bot         types.Bool   `tfsdk:"bot"`
}

func (d *AccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account"
}

func (d *AccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Account data source",

		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				MarkdownDescription: "Account configurable attribute",
				Optional:            false,
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Account identifier",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
			"display_name": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
			"note": schema.StringAttribute{
				MarkdownDescription: "",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
			"bot": schema.BoolAttribute{
				MarkdownDescription: "",
				Computed:            true,
				Optional:            false,
				Required:            false,
			},
		},
	}
}

func (d *AccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mastodon.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *mastodon.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *AccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccountDataSourceModel

	tflog.Debug(ctx, "mastodon_account data source read")

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	account, err := d.client.AccountLookup(ctx, data.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to lookup account",
			fmt.Sprintf("Failed to lookup account: %s", err),
		)
		return
	}

	data.Id = types.StringValue(string(account.ID))
	data.DisplayName = types.StringValue(account.DisplayName)
	data.Note = types.StringValue(account.Note)
	data.Locked = types.BoolValue(account.Locked)
	data.Bot = types.BoolValue(account.Bot)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read the mastodon_account data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
