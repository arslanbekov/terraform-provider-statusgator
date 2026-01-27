package provider

import (
	"context"
	"fmt"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &PingMonitorResource{}
	_ resource.ResourceWithImportState = &PingMonitorResource{}
)

type PingMonitorResource struct {
	client *statusgator.Client
}

type PingMonitorResourceModel struct {
	ID            types.String `tfsdk:"id"`
	BoardID       types.String `tfsdk:"board_id"`
	Name          types.String `tfsdk:"name"`
	Host          types.String `tfsdk:"host"`
	CheckInterval types.Int64  `tfsdk:"check_interval"`
	Regions       types.List   `tfsdk:"regions"`
	GroupID       types.String `tfsdk:"group_id"`
	Status        types.String `tfsdk:"status"`
	Paused        types.Bool   `tfsdk:"paused"`
}

func NewPingMonitorResource() resource.Resource {
	return &PingMonitorResource{}
}

func (r *PingMonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ping_monitor"
}

func (r *PingMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusGator ping (ICMP) monitor.",
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
			"host": schema.StringAttribute{
				Description: "The host to ping (hostname or IP address).",
				Required:    true,
			},
			"check_interval": schema.Int64Attribute{
				Description: "Check interval in minutes. Must be between 1 and 1440.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1440),
				},
			},
			"regions": schema.ListAttribute{
				Description: "List of monitoring regions.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"group_id": schema.StringAttribute{
				Description: "The ID of the monitor group this monitor belongs to.",
				Optional:    true,
			},
			"status": schema.StringAttribute{
				Description: "Current status of the monitor (up, down, warn, unknown).",
				Computed:    true,
			},
			"paused": schema.BoolAttribute{
				Description: "Whether the monitor is paused.",
				Computed:    true,
			},
		},
	}
}

func (r *PingMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PingMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PingMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &statusgator.PingMonitorRequest{
		Name:     data.Name.ValueString(),
		Address:  data.Host.ValueString(),
		Interval: int(data.CheckInterval.ValueInt64()),
		GroupID:  data.GroupID.ValueString(),
	}

	// Convert regions
	if !data.Regions.IsNull() {
		var regions []string
		resp.Diagnostics.Append(data.Regions.ElementsAs(ctx, &regions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Regions = regions
	}

	tflog.Debug(ctx, "Creating ping monitor", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"name":     data.Name.ValueString(),
		"host":     data.Host.ValueString(),
	})

	monitor, err := r.client.PingMonitors.Create(ctx, data.BoardID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating StatusGator Ping Monitor",
			fmt.Sprintf("Could not create ping monitor: %s", err.Error()),
		)
		return
	}

	r.mapMonitorToModel(monitor, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PingMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PingMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading ping monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	monitors, _, err := r.client.Monitors.List(ctx, data.BoardID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading StatusGator Ping Monitor",
			fmt.Sprintf("Could not list monitors: %s", err.Error()),
		)
		return
	}

	var found bool
	for _, m := range monitors {
		if m.ID == data.ID.ValueString() && m.MonitorType == statusgator.MonitorTypePing {
			found = true
			data.Name = types.StringValue(m.DisplayName)
			data.Status = types.StringValue(string(m.FilteredStatus))
			data.Paused = types.BoolValue(m.IsPaused())
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

func (r *PingMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PingMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := &statusgator.PingMonitorRequest{
		Name:     data.Name.ValueString(),
		Address:  data.Host.ValueString(),
		Interval: int(data.CheckInterval.ValueInt64()),
		GroupID:  data.GroupID.ValueString(),
	}

	// Convert regions
	if !data.Regions.IsNull() {
		var regions []string
		resp.Diagnostics.Append(data.Regions.ElementsAs(ctx, &regions, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.Regions = regions
	}

	tflog.Debug(ctx, "Updating ping monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	monitor, err := r.client.PingMonitors.Update(ctx, data.BoardID.ValueString(), data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating StatusGator Ping Monitor",
			fmt.Sprintf("Could not update ping monitor ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	r.mapMonitorToModel(monitor, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PingMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data PingMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting ping monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	err := r.client.Monitors.Delete(ctx, data.BoardID.ValueString(), data.ID.ValueString())
	if err != nil {
		if statusgator.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting StatusGator Ping Monitor",
			fmt.Sprintf("Could not delete ping monitor ID %s: %s", data.ID.ValueString(), err.Error()),
		)
	}
}

func (r *PingMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importBoardChildState(ctx, req, resp)
}

func (r *PingMonitorResource) mapMonitorToModel(monitor *statusgator.PingMonitor, data *PingMonitorResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(monitor.ID)
	data.Name = types.StringValue(monitor.DisplayName)
	data.Host = types.StringValue(monitor.Address)
	data.CheckInterval = types.Int64Value(int64(monitor.Interval))
	data.Status = types.StringValue(string(monitor.FilteredStatus))
	data.Paused = types.BoolValue(monitor.IsPaused())

	if monitor.Group != nil {
		data.GroupID = types.StringValue(monitor.Group.ID)
	} else {
		data.GroupID = types.StringNull()
	}

	// Convert regions
	if len(monitor.Regions) > 0 {
		regionElements := make([]attr.Value, len(monitor.Regions))
		for i, region := range monitor.Regions {
			regionElements[i] = types.StringValue(region)
		}
		regionList, d := types.ListValue(types.StringType, regionElements)
		diags.Append(d...)
		data.Regions = regionList
	} else {
		data.Regions = types.ListNull(types.StringType)
	}
}
