// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mattn/go-mastodon"
)

// Ensure MastodonProvider satisfies various provider interfaces.
var _ provider.Provider = &MastodonProvider{}
var _ provider.ProviderWithFunctions = &MastodonProvider{}

// MastodonProvider defines the provider implementation.
type MastodonProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// MastodonProviderModel describes the provider data model.
type MastodonProviderModel struct {
	Host         types.String `tfsdk:"host"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
	Email        types.String `tfsdk:"email"`
	Password     types.String `tfsdk:"password"`
	AccessToken  types.String `tfsdk:"access_token"`
}

func (p *MastodonProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "mastodon"
	resp.Version = p.version
}

func (p *MastodonProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Mastodon host to connect to.",
				Optional:            true,
			},
			"client_id": schema.StringAttribute{
				MarkdownDescription: "Client ID for Mastodon App.",
				Optional:            true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "Client Secret for Mastodon App.",
				Optional:            true,
			},
			"email": schema.StringAttribute{
				MarkdownDescription: "Username to connect to the server as.",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password to use for connecting to the server.",
				Optional:            true,
				Sensitive:           true,
			},
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Password to use for connecting to the server.",
				Optional:            true,
				Sensitive:           true,
				DeprecationMessage:  "Use email and password instead.",
			},
		},
	}
}

func (p *MastodonProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data MastodonProviderModel
	tflog.Debug(ctx, "mastodon_provider configure")
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if data.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Mastodon API Host",
			"The provider cannot create the Mastodon API client as there is an unknown configuration value for the Mastodon API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MASTODON_HOST environment variable.",
		)
	}
	host := os.Getenv("MASTODON_HOST")
	if !data.Host.IsNull() {
		host = data.Host.ValueString()
	}
	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("user-access-token"),
			"Missing Mastodon Credentials",
			"The provider cannot create the Mastodon API client as the Host is not set.",
		)
	}

	if data.ClientID.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client-id"),
			"Unknown Mastodon Client ID",
			"The provider cannot create the Mastodon API client as there is an unknown configuration value for the Mastodon Client ID. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MASTODON_CLIENT_ID environment variable.",
		)
	}
	client_id := os.Getenv("MASTODON_CLIENT_ID")
	if !data.ClientID.IsNull() {
		client_id = data.ClientID.ValueString()
	}
	if client_id == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("user-access-token"),
			"Missing Mastodon Credentials",
			"The provider cannot create the Mastodon API client as the Client ID is not set.",
		)
	}

	if data.ClientSecret.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("client-secret"),
			"Unknown Mastodon Client Secret",
			"The provider cannot create the Mastodon API client as there is an unknown configuration value for the Mastodon Client Secret. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MASTODON_CLIENT_SECRET environment variable.",
		)
	}
	client_secret := os.Getenv("MASTODON_CLIENT_SECRET")
	if !data.ClientSecret.IsNull() {
		client_secret = data.ClientSecret.ValueString()
	}
	if client_secret == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("user-access-token"),
			"Missing Mastodon Credentials",
			"The provider cannot create the Mastodon API client as the Client Secret is not set.",
		)
	}

	if data.Email.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("user-email"),
			"Unknown Mastodon User Email",
			"The provider cannot create the Mastodon API client as there is an unknown configuration value for the Mastodon User Email. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MASTODON_USER_EMAIL environment variable.",
		)
	}
	user_email := os.Getenv("MASTODON_USER_EMAIL")
	if !data.Email.IsNull() {
		user_email = data.Email.ValueString()
	}

	if data.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("user-password"),
			"Unknown Mastodon User Password",
			"The provider cannot create the Mastodon API client as there is an unknown configuration value for the Mastodon User Password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MASTODON_USER_PASSWORD environment variable.",
		)
	}
	user_password := os.Getenv("MASTODON_USER_PASSWORD")
	if !data.Password.IsNull() {
		user_password = data.Password.ValueString()
	}

	if data.AccessToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("user-access-token"),
			"Unknown Mastodon User Password",
			"The provider cannot create the Mastodon API client as there is an unknown configuration value for the Mastodon User Password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the MASTODON_USER_PASSWORD environment variable.",
		)
	}
	access_token := os.Getenv("MASTODON_ACCESS_TOKEN")
	if !data.AccessToken.IsNull() {
		access_token = data.AccessToken.ValueString()
	}

	if access_token == "" && (user_email == "" || user_password == "") {
		resp.Diagnostics.AddAttributeError(
			path.Root("user-access-token"),
			"Missing Mastodon Credentials",
			"The provider cannot create the Mastodon API client as neither the Access Token or the Username and Password fields are set.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var config mastodon.Config
	if access_token != "" {
		tflog.Debug(ctx, "mastodon_provider configure with access token")
		config = mastodon.Config{
			Server:       host,
			ClientID:     client_id,
			ClientSecret: client_secret,
			AccessToken:  access_token,
		}
	} else {
		tflog.Debug(ctx, "mastodon_provider configure without access token")
		config = mastodon.Config{
			Server:       host,
			ClientID:     client_id,
			ClientSecret: client_secret,
		}
	}

	c := mastodon.NewClient(&config)
	user, err := c.GetAccountCurrentUser(context.Background())
	if err != nil {
		tflog.Error(ctx, "GetAccountCurrentUser Error: "+err.Error())
		resp.Diagnostics.AddError(
			"Mastodon GetAccountCurrentUser Failed, API is not usable.",
			err.Error(),
		)
	}

	tflog.Debug(ctx, "mastodon_provider current user: "+user.Acct)

	if access_token != "" {
		ctx = tflog.SetField(ctx, "mastodon_access_token", access_token)       //ANNO We can log the access token to help with debugging.
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "mastodon_access_token") //ANNO We can also make sure to filter out the value from the logs.
	} else if user_email != "" && user_password != "" {
		ctx = tflog.SetField(ctx, "mastodon_user_email", user_email)
		ctx = tflog.SetField(ctx, "mastodon_user_password", user_password)
		ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "mastodon_user_password")
	} else {
		resp.Diagnostics.AddAttributeError( //ANNO We can provide more than one error on the same flow.
			path.Root("user-access-token"),
			"Missing Mastodon Credentials",
			"The provider cannot create the Mastodon API client as neither the Access Token or the Username and Password fields are set.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Example client configuration for data sources and resources
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *MastodonProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPostResource,
	}
}

func (p *MastodonProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAccountDataSource,
	}
}

func (p *MastodonProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		NewIdentityFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MastodonProvider{
			version: version,
		}
	}
}
