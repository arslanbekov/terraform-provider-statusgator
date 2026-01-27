package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &SubscriberResource{}
	_ resource.ResourceWithImportState = &SubscriberResource{}
)

type SubscriberResource struct {
	client *statusgator.Client
}

type SubscriberResourceModel struct {
	ID               types.String `tfsdk:"id"`
	BoardID          types.String `tfsdk:"board_id"`
	Email            types.String `tfsdk:"email"`
	SkipConfirmation types.Bool   `tfsdk:"skip_confirmation"`
	Confirmed        types.Bool   `tfsdk:"confirmed"`
}

func NewSubscriberResource() resource.Resource {
	return &SubscriberResource{}
}

func (r *SubscriberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscriber"
}

func (r *SubscriberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusGator status page subscriber.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the subscriber.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"board_id": schema.StringAttribute{
				Description: "The ID of the board this subscriber belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"email": schema.StringAttribute{
				Description: "The email address of the subscriber.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"skip_confirmation": schema.BoolAttribute{
				Description: "Whether to skip the confirmation email. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"confirmed": schema.BoolAttribute{
				Description: "Whether the subscriber has confirmed their email.",
				Computed:    true,
			},
		},
	}
}

func (r *SubscriberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*statusgator.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *statusgator.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *SubscriberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SubscriberResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &statusgator.SubscriberRequest{
		Email:            data.Email.ValueString(),
		SkipConfirmation: data.SkipConfirmation.ValueBool(),
	}

	tflog.Debug(ctx, "Creating subscriber", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"email":    data.Email.ValueString(),
	})

	subscriber, err := r.client.Subscribers.Add(ctx, data.BoardID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating StatusGator Subscriber",
			fmt.Sprintf("Could not create subscriber: %s", err.Error()),
		)
		return
	}

	data.ID = types.StringValue(subscriber.ID)
	data.Email = types.StringValue(subscriber.Email)
	data.Confirmed = types.BoolValue(subscriber.Confirmed)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SubscriberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading subscriber", map[string]interface{}{
		"board_id":      data.BoardID.ValueString(),
		"subscriber_id": data.ID.ValueString(),
	})

	subscribers, err := r.client.Subscribers.ListAll(ctx, data.BoardID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading StatusGator Subscriber",
			fmt.Sprintf("Could not list subscribers: %s", err.Error()),
		)
		return
	}

	var found bool
	for _, s := range subscribers {
		if s.ID == data.ID.ValueString() {
			found = true
			data.Email = types.StringValue(s.Email)
			data.Confirmed = types.BoolValue(s.Confirmed)
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Subscribers cannot be updated - email and board_id both force replacement
	var data SubscriberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SubscriberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SubscriberResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting subscriber", map[string]interface{}{
		"board_id":      data.BoardID.ValueString(),
		"subscriber_id": data.ID.ValueString(),
	})

	err := r.client.Subscribers.DeleteByID(ctx, data.BoardID.ValueString(), data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting StatusGator Subscriber",
			fmt.Sprintf("Could not delete subscriber ID %s: %s", data.ID.ValueString(), err.Error()),
		)
	}
}

func (r *SubscriberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: board_id/subscriber_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("board_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
