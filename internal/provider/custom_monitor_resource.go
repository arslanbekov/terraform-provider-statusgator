package provider

import (
	"context"
	"fmt"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &CustomMonitorResource{}
	_ resource.ResourceWithImportState = &CustomMonitorResource{}
)

type CustomMonitorResource struct {
	client *statusgator.Client
}

type CustomMonitorResourceModel struct {
	ID          types.String `tfsdk:"id"`
	BoardID     types.String `tfsdk:"board_id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	GroupID     types.String `tfsdk:"group_id"`
}

func NewCustomMonitorResource() resource.Resource {
	return &CustomMonitorResource{}
}

func (r *CustomMonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_custom_monitor"
}

func (r *CustomMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusGator custom monitor with manually-controlled status.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the monitor.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"board_id": schema.StringAttribute{
				Description: "The ID of the board this monitor belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the monitor.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the monitor.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Status of the monitor: up, warn, or down.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("up", "warn", "down"),
				},
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the monitor group this monitor belongs to.",
				Optional:    true,
			},
		},
	}
}

func (r *CustomMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*statusgator.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *statusgator.Client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *CustomMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CustomMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &statusgator.CustomMonitorRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Status:      statusgator.MonitorStatus(data.Status.ValueString()),
		GroupID:     data.GroupID.ValueString(),
	}

	tflog.Debug(ctx, "Creating custom monitor", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"name":     data.Name.ValueString(),
		"status":   data.Status.ValueString(),
	})

	monitor, err := r.client.CustomMonitors.Create(ctx, data.BoardID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating StatusGator Custom Monitor",
			fmt.Sprintf("Could not create custom monitor: %s", err.Error()),
		)
		return
	}

	data.ID = types.StringValue(monitor.ID)
	data.Name = types.StringValue(monitor.DisplayName)
	if monitor.Description != nil {
		data.Description = types.StringValue(*monitor.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.Status = types.StringValue(string(monitor.FilteredStatus))
	if monitor.Group != nil {
		data.GroupID = types.StringValue(monitor.Group.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CustomMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading custom monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	monitors, _, err := r.client.Monitors.List(ctx, data.BoardID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading StatusGator Custom Monitor",
			fmt.Sprintf("Could not list monitors: %s", err.Error()),
		)
		return
	}

	var found bool
	for _, m := range monitors {
		if m.ID == data.ID.ValueString() && m.MonitorType == statusgator.MonitorTypeCustom {
			found = true
			data.Name = types.StringValue(m.DisplayName)
			data.Status = types.StringValue(string(m.FilteredStatus))
			if m.Group != nil {
				data.GroupID = types.StringValue(m.Group.ID)
			} else {
				data.GroupID = types.StringNull()
			}
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CustomMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &statusgator.CustomMonitorRequest{
		Name:        data.Name.ValueString(),
		Description: data.Description.ValueString(),
		Status:      statusgator.MonitorStatus(data.Status.ValueString()),
		GroupID:     data.GroupID.ValueString(),
	}

	tflog.Debug(ctx, "Updating custom monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	monitor, err := r.client.CustomMonitors.Update(ctx, data.BoardID.ValueString(), data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating StatusGator Custom Monitor",
			fmt.Sprintf("Could not update custom monitor ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	data.Name = types.StringValue(monitor.DisplayName)
	if monitor.Description != nil {
		data.Description = types.StringValue(*monitor.Description)
	} else {
		data.Description = types.StringNull()
	}
	data.Status = types.StringValue(string(monitor.FilteredStatus))
	if monitor.Group != nil {
		data.GroupID = types.StringValue(monitor.Group.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CustomMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CustomMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting custom monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	err := r.client.Monitors.Delete(ctx, data.BoardID.ValueString(), data.ID.ValueString())
	if err != nil {
		if statusgator.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting StatusGator Custom Monitor",
			fmt.Sprintf("Could not delete custom monitor ID %s: %s", data.ID.ValueString(), err.Error()),
		)
	}
}

func (r *CustomMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importBoardChildState(ctx, req, resp)
}
