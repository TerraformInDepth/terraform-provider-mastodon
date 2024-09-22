package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mattn/go-mastodon"
	"github.com/microcosm-cc/bluemonday"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &PostResource{}
var _ resource.ResourceWithImportState = &PostResource{}

func NewPostResource() resource.Resource {
	return &PostResource{}
}

// PostResource defines the resource implementation.
type PostResource struct {
	client *mastodon.Client
}

// PostResourceModel describes the resource data model.
type PostResourceModel struct {
	Id                types.String `tfsdk:"id"`
	CreatedAt         types.String `tfsdk:"created_at"`
	Account           types.String `tfsdk:"account"`
	Content           types.String `tfsdk:"content"`
	Visibility        types.String `tfsdk:"visibility"`
	Sensitive         types.Bool   `tfsdk:"sensitive"`
	PreserveOnDestroy types.Bool   `tfsdk:"preserve_on_destroy"`
}

func (r *PostResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_post"
}

func (r *PostResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "This resource is used to manage posts on a Mastodon instance.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				Required:            false,
				Optional:            false,
				MarkdownDescription: "Unique identifier of the post.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Timestamp of when the post was created.",
				Computed:            true,
				Required:            false,
				Optional:            false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account": schema.StringAttribute{
				MarkdownDescription: "Account that created the post",
				Computed:            true,
				Required:            false,
				Optional:            false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The content of the post.",
				Required:            true,
			},
			"visibility": schema.StringAttribute{
				MarkdownDescription: "The post visibility: can be `public`, `unlisted`, `private`, or `direct`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("public"),
			},
			"sensitive": schema.BoolAttribute{
				MarkdownDescription: "Whether the post contains sensitive content.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"preserve_on_destroy": schema.BoolAttribute{
				MarkdownDescription: "When destroyed, preserve the post on the server.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
		},
	}
}

func (r *PostResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*mastodon.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *mastodon.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PostResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PostResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	toot := mastodon.Toot{
		Status:     data.Content.ValueString(),
		Visibility: data.Visibility.ValueString(),
		Sensitive:  data.Sensitive.ValueBool(),
	}

	post, err := r.client.PostStatus(context.Background(), &toot)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create post, got error: %s", err))
		return
	}

	p := bluemonday.NewPolicy()

	// Update the model with the created post data
	data.Id = types.StringValue(string(post.ID))
	data.CreatedAt = types.StringValue(post.CreatedAt.String())
	data.Account = types.StringValue(string(post.Account.ID))
	data.Content = types.StringValue(p.Sanitize(post.Content))
	data.Visibility = types.StringValue(post.Visibility)
	data.Sensitive = types.BoolValue(post.Sensitive)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PostResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PostResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	post, err := r.client.GetStatus(context.Background(), mastodon.ID(data.Id.ValueString()))

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read post, got error: %s", err))
		return
	}

	p := bluemonday.NewPolicy()

	data.Id = types.StringValue(string(post.ID))
	data.CreatedAt = types.StringValue(post.CreatedAt.String())
	data.Account = types.StringValue(string(post.Account.ID))
	data.Content = types.StringValue(p.Sanitize(post.Content))
	data.Visibility = types.StringValue(post.Visibility)
	data.Sensitive = types.BoolValue(post.Sensitive)

	// During imports the `preserve_on_destroy` attribute may not be set.
	if data.PreserveOnDestroy.IsNull() {
		data.PreserveOnDestroy = types.BoolValue(false)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PostResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PostResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	toot := mastodon.Toot{
		Status:     data.Content.ValueString(),
		Visibility: data.Visibility.ValueString(),
		Sensitive:  data.Sensitive.ValueBool(),
	}

	post, err := r.client.UpdateStatus(context.Background(), &toot, mastodon.ID(data.Id.ValueString()))

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read post, got error: %s", err))
		return
	}

	p := bluemonday.NewPolicy()

	data.Id = types.StringValue(string(post.ID))
	data.CreatedAt = types.StringValue(post.CreatedAt.String())
	data.Account = types.StringValue(string(post.Account.ID))
	data.Content = types.StringValue(p.Sanitize(post.Content))
	data.Visibility = types.StringValue(post.Visibility)
	data.Sensitive = types.BoolValue(post.Sensitive)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PostResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PostResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.PreserveOnDestroy.ValueBool() {
		tflog.Debug(ctx, "preserve_on_destroy is enabled: preserving post on server.")
		return
	}

	err := r.client.DeleteStatus(context.Background(), mastodon.ID(data.Id.ValueString()))

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete post, got error: %s", err))
		return
	}

}

func (r *PostResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
