package provider

import (
	"context"
	"fmt"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var (
	_ resource.Resource                = &WebsiteMonitorResource{}
	_ resource.ResourceWithImportState = &WebsiteMonitorResource{}
)

type WebsiteMonitorResource struct {
	client *statusgator.Client
}

type WebsiteMonitorResourceModel struct {
	ID              types.String `tfsdk:"id"`
	BoardID         types.String `tfsdk:"board_id"`
	Name            types.String `tfsdk:"name"`
	URL             types.String `tfsdk:"url"`
	CheckInterval   types.Int64  `tfsdk:"check_interval"`
	HTTPMethod      types.String `tfsdk:"http_method"`
	ExpectedStatus  types.Int64  `tfsdk:"expected_status"`
	ContentMatch    types.String `tfsdk:"content_match"`
	Timeout         types.Int64  `tfsdk:"timeout"`
	FollowRedirects types.Bool   `tfsdk:"follow_redirects"`
	Headers         types.Map    `tfsdk:"headers"`
	Regions         types.List   `tfsdk:"regions"`
	GroupID         types.String `tfsdk:"group_id"`
	Status          types.String `tfsdk:"status"`
	Paused          types.Bool   `tfsdk:"paused"`
}

func NewWebsiteMonitorResource() resource.Resource {
	return &WebsiteMonitorResource{}
}

func (r *WebsiteMonitorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_website_monitor"
}

func (r *WebsiteMonitorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusGator website HTTP monitor.",
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
			"url": schema.StringAttribute{
				Description: "The URL to monitor. Must be a valid HTTP or HTTPS URL.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						urlRegex(),
						"must be a valid HTTP or HTTPS URL",
					),
				},
			},
			"check_interval": schema.Int64Attribute{
				Description: "Check interval in minutes. Must be between 1 and 1440.",
				Required:    true,
				Validators: []validator.Int64{
					int64validator.Between(1, 1440),
				},
			},
			"http_method": schema.StringAttribute{
				Description: "HTTP method to use. Defaults to GET.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("GET"),
				Validators: []validator.String{
					stringvalidator.OneOf("GET", "POST", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS"),
				},
			},
			"expected_status": schema.Int64Attribute{
				Description: "Expected HTTP status code. Defaults to 200. Must be between 100 and 599.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(200),
				Validators: []validator.Int64{
					int64validator.Between(100, 599),
				},
			},
			"content_match": schema.StringAttribute{
				Description: "String to match in response body.",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Request timeout in seconds. Defaults to 30. Must be between 1 and 120.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(30),
				Validators: []validator.Int64{
					int64validator.Between(1, 120),
				},
			},
			"follow_redirects": schema.BoolAttribute{
				Description: "Whether to follow HTTP redirects. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"headers": schema.MapAttribute{
				Description: "Custom HTTP headers to send. Note: Headers may contain sensitive data like authentication tokens.",
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
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

func (r *WebsiteMonitorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *WebsiteMonitorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WebsiteMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	followRedirects := data.FollowRedirects.ValueBool()
	createReq := &statusgator.WebsiteMonitorRequest{
		Name:            data.Name.ValueString(),
		URL:             data.URL.ValueString(),
		CheckInterval:   int(data.CheckInterval.ValueInt64()),
		HTTPMethod:      data.HTTPMethod.ValueString(),
		ExpectedStatus:  int(data.ExpectedStatus.ValueInt64()),
		ContentMatch:    data.ContentMatch.ValueString(),
		Timeout:         int(data.Timeout.ValueInt64()),
		FollowRedirects: &followRedirects,
		GroupID:         data.GroupID.ValueString(),
	}

	// Convert headers
	if !data.Headers.IsNull() {
		headers := make(map[string]string)
		resp.Diagnostics.Append(data.Headers.ElementsAs(ctx, &headers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.Headers = headers
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

	tflog.Debug(ctx, "Creating website monitor", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"name":     data.Name.ValueString(),
	})

	monitor, err := r.client.WebsiteMonitors.Create(ctx, data.BoardID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating StatusGator Website Monitor",
			fmt.Sprintf("Could not create website monitor: %s", err.Error()),
		)
		return
	}

	r.mapMonitorToModel(ctx, monitor, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebsiteMonitorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WebsiteMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading website monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	// Note: The StatusGator API doesn't provide a direct Get endpoint for monitors.
	// We use List and filter to verify the monitor exists and refresh basic fields.
	// Full configuration (URL, headers, etc.) is preserved from state.
	monitors, _, err := r.client.Monitors.List(ctx, data.BoardID.ValueString(), nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading StatusGator Website Monitor",
			fmt.Sprintf("Could not list monitors: %s", err.Error()),
		)
		return
	}

	var found bool
	for _, m := range monitors {
		if m.ID == data.ID.ValueString() && m.Type == statusgator.MonitorTypeWebsite {
			found = true
			// Only update fields available in the generic Monitor response
			data.Name = types.StringValue(m.Name)
			data.Status = types.StringValue(string(m.Status))
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

func (r *WebsiteMonitorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WebsiteMonitorResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	followRedirects := data.FollowRedirects.ValueBool()
	updateReq := &statusgator.WebsiteMonitorRequest{
		Name:            data.Name.ValueString(),
		URL:             data.URL.ValueString(),
		CheckInterval:   int(data.CheckInterval.ValueInt64()),
		HTTPMethod:      data.HTTPMethod.ValueString(),
		ExpectedStatus:  int(data.ExpectedStatus.ValueInt64()),
		ContentMatch:    data.ContentMatch.ValueString(),
		Timeout:         int(data.Timeout.ValueInt64()),
		FollowRedirects: &followRedirects,
		GroupID:         data.GroupID.ValueString(),
	}

	// Convert headers
	if !data.Headers.IsNull() {
		headers := make(map[string]string)
		resp.Diagnostics.Append(data.Headers.ElementsAs(ctx, &headers, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.Headers = headers
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

	tflog.Debug(ctx, "Updating website monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	monitor, err := r.client.WebsiteMonitors.Update(ctx, data.BoardID.ValueString(), data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating StatusGator Website Monitor",
			fmt.Sprintf("Could not update website monitor ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	r.mapMonitorToModel(ctx, monitor, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebsiteMonitorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WebsiteMonitorResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting website monitor", map[string]interface{}{
		"board_id":   data.BoardID.ValueString(),
		"monitor_id": data.ID.ValueString(),
	})

	err := r.client.Monitors.Delete(ctx, data.BoardID.ValueString(), data.ID.ValueString())
	if err != nil {
		if statusgator.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Deleting StatusGator Website Monitor",
			fmt.Sprintf("Could not delete website monitor ID %s: %s", data.ID.ValueString(), err.Error()),
		)
	}
}

func (r *WebsiteMonitorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importBoardChildState(ctx, req, resp)
}

func (r *WebsiteMonitorResource) mapMonitorToModel(ctx context.Context, monitor *statusgator.WebsiteMonitor, data *WebsiteMonitorResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(monitor.ID)
	data.Name = types.StringValue(monitor.Name)
	data.URL = types.StringValue(monitor.URL)
	data.CheckInterval = types.Int64Value(int64(monitor.CheckInterval))
	data.HTTPMethod = types.StringValue(monitor.HTTPMethod)
	data.ExpectedStatus = types.Int64Value(int64(monitor.ExpectedStatus))
	data.Timeout = types.Int64Value(int64(monitor.Timeout))
	data.FollowRedirects = types.BoolValue(monitor.FollowRedirects)
	data.Status = types.StringValue(string(monitor.Status))
	data.Paused = types.BoolValue(monitor.IsPaused())

	if monitor.ContentMatch != "" {
		data.ContentMatch = types.StringValue(monitor.ContentMatch)
	} else {
		data.ContentMatch = types.StringNull()
	}

	if monitor.Group != nil {
		data.GroupID = types.StringValue(monitor.Group.ID)
	} else {
		data.GroupID = types.StringNull()
	}

	// Convert headers
	if len(monitor.Headers) > 0 {
		headerElements := make(map[string]attr.Value)
		for k, v := range monitor.Headers {
			headerElements[k] = types.StringValue(v)
		}
		headerMap, d := types.MapValue(types.StringType, headerElements)
		diags.Append(d...)
		data.Headers = headerMap
	} else {
		data.Headers = types.MapNull(types.StringType)
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
