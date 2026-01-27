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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &MonitorGroupResource{}
	_ resource.ResourceWithImportState = &MonitorGroupResource{}
)

type MonitorGroupResource struct {
	client *statusgator.Client
}

type MonitorGroupResourceModel struct {
	ID        types.String `tfsdk:"id"`
	BoardID   types.String `tfsdk:"board_id"`
	Name      types.String `tfsdk:"name"`
	Position  types.Int64  `tfsdk:"position"`
	Collapsed types.Bool   `tfsdk:"collapsed"`
}

func NewMonitorGroupResource() resource.Resource {
	return &MonitorGroupResource{}
}

func (r *MonitorGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_group"
}

func (r *MonitorGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusGator monitor group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the monitor group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"board_id": schema.StringAttribute{
				Description: "The ID of the board this group belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the monitor group.",
				Required:    true,
			},
			"position": schema.Int64Attribute{
				Description: "The display position of the group. Defaults to 0.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
			},
			"collapsed": schema.BoolAttribute{
				Description: "Whether the group is collapsed by default. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *MonitorGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MonitorGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MonitorGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	collapsed := data.Collapsed.ValueBool()
	createReq := &statusgator.MonitorGroupRequest{
		Name:      data.Name.ValueString(),
		Position:  int(data.Position.ValueInt64()),
		Collapsed: &collapsed,
	}

	tflog.Debug(ctx, "Creating monitor group", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"name":     data.Name.ValueString(),
	})

	group, err := r.client.MonitorGroups.Create(ctx, data.BoardID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating StatusGator Monitor Group",
			fmt.Sprintf("Could not create monitor group: %s", err.Error()),
		)
		return
	}

	data.ID = types.StringValue(group.ID)
	data.Name = types.StringValue(group.Name)
	data.Position = types.Int64Value(int64(group.Position))
	data.Collapsed = types.BoolValue(group.Collapsed)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MonitorGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading monitor group", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"group_id": data.ID.ValueString(),
	})

	group, err := r.client.MonitorGroups.Get(ctx, data.BoardID.ValueString(), data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading StatusGator Monitor Group",
			fmt.Sprintf("Could not read monitor group ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	data.ID = types.StringValue(group.ID)
	data.Name = types.StringValue(group.Name)
	data.Position = types.Int64Value(int64(group.Position))
	data.Collapsed = types.BoolValue(group.Collapsed)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MonitorGroupResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	collapsed := data.Collapsed.ValueBool()
	updateReq := &statusgator.MonitorGroupRequest{
		Name:      data.Name.ValueString(),
		Position:  int(data.Position.ValueInt64()),
		Collapsed: &collapsed,
	}

	tflog.Debug(ctx, "Updating monitor group", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"group_id": data.ID.ValueString(),
	})

	group, err := r.client.MonitorGroups.Update(ctx, data.BoardID.ValueString(), data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating StatusGator Monitor Group",
			fmt.Sprintf("Could not update monitor group ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	data.Name = types.StringValue(group.Name)
	data.Position = types.Int64Value(int64(group.Position))
	data.Collapsed = types.BoolValue(group.Collapsed)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *MonitorGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data MonitorGroupResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting monitor group", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"group_id": data.ID.ValueString(),
	})

	err := r.client.MonitorGroups.Delete(ctx, data.BoardID.ValueString(), data.ID.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting StatusGator Monitor Group",
			fmt.Sprintf("Could not delete monitor group ID %s: %s", data.ID.ValueString(), err.Error()),
		)
	}
}

func (r *MonitorGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: board_id/group_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID format: board_id/group_id, got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("board_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
