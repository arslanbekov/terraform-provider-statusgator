package provider

import (
	"context"
	"fmt"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &ServiceMonitorResource{}
	_ resource.ResourceWithImportState = &ServiceMonitorResource{}
)

type ServiceMonitorResource struct {
	client *statusgator.Client
}

type ServiceMonitorResourceModel struct {
	ID          types.String `tfsdk:"id"`
	BoardID     types.String `tfsdk:"board_id"`
	ServiceID   types.String `tfsdk:"service_id"`
	Name        types.String `tfsdk:"name"`
	GroupID     types.String `tfsdk:"group_id"`
	ServiceName types.String `tfsdk:"service_name"`
	Status      types.String `tfsdk:"status"`
}

func NewServiceMonitorResource() resource.Resource {
	return &ServiceMonitorResource{}
}

func (r *ServiceMonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_monitor"
}

func (r *ServiceMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusGator service monitor that tracks an external status page.",
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
			"service_id": schema.StringAttribute{
				Description: "The ID of the external service to monitor.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Custom display name for the monitor. If not set, the service name will be used.",
				Optional:    true,
				Computed:    true,
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the monitor group this monitor belongs to.",
				Optional:    true,
			},
			"service_name": schema.StringAttribute{
				Description: "The name of the external service.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the monitor (up, down, warn, unknown).",
				Computed:    true,
			},
		},
	}
}

func (r *ServiceMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServiceMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ServiceMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &statusgator.ServiceMonitorRequest{
		ServiceID: data.ServiceID.ValueString(),
		Name:      data.Name.ValueString(),
		GroupID:   data.GroupID.ValueString(),
	}

	tflog.Debug(ctx, "Creating service monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"service_id": data.ServiceID.ValueString(),
	})

	monitor, err := r.client.ServiceMonitors.Create(ctx, data.BoardID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating StatusGator Service Monitor",
			fmt.Sprintf("Could not create service monitor: %s", err.Error()),
		)
		return
	}

	data.ID = types.StringValue(monitor.ID)
	data.Name = types.StringValue(monitor.Name)
	data.ServiceID = types.StringValue(monitor.GetServiceID())
	data.ServiceName = types.StringValue(monitor.GetServiceName())
	data.Status = types.StringValue(string(monitor.Status))
	if monitor.Group != nil {
		data.GroupID = types.StringValue(monitor.Group.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ServiceMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading service monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	monitors, _, err := r.client.Monitors.List(ctx, data.BoardID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading StatusGator Service Monitor",
			fmt.Sprintf("Could not list monitors: %s", err.Error()),
		)
		return
	}

	var found bool
	for _, m := range monitors {
		if m.ID == data.ID.ValueString() && m.Type == statusgator.MonitorTypeService {
			found = true
			data.Name = types.StringValue(m.Name)
			data.Status = types.StringValue(string(m.Status))
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

func (r *ServiceMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ServiceMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &statusgator.ServiceMonitorRequest{
		Name:    data.Name.ValueString(),
		GroupID: data.GroupID.ValueString(),
	}

	tflog.Debug(ctx, "Updating service monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	monitor, err := r.client.ServiceMonitors.Update(ctx, data.BoardID.ValueString(), data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating StatusGator Service Monitor",
			fmt.Sprintf("Could not update service monitor ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	data.Name = types.StringValue(monitor.Name)
	data.ServiceID = types.StringValue(monitor.GetServiceID())
	data.ServiceName = types.StringValue(monitor.GetServiceName())
	data.Status = types.StringValue(string(monitor.Status))
	if monitor.Group != nil {
		data.GroupID = types.StringValue(monitor.Group.ID)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ServiceMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ServiceMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting service monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	err := r.client.Monitors.Delete(ctx, data.BoardID.ValueString(), data.ID.ValueString())
	if err != nil {
		if statusgator.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting StatusGator Service Monitor",
			fmt.Sprintf("Could not delete service monitor ID %s: %s", data.ID.ValueString(), err.Error()),
		)
	}
}

func (r *ServiceMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importBoardChildState(ctx, req, resp)
}
