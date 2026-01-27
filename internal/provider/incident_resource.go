package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &IncidentResource{}
	_ resource.ResourceWithImportState = &IncidentResource{}
)

type IncidentResource struct {
	client *statusgator.Client
}

type IncidentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	BoardID      types.String `tfsdk:"board_id"`
	Title        types.String `tfsdk:"title"`
	Message      types.String `tfsdk:"message"`
	Severity     types.String `tfsdk:"severity"`
	Phase        types.String `tfsdk:"phase"`
	MonitorIDs   types.List   `tfsdk:"monitor_ids"`
	ScheduledFor types.String `tfsdk:"scheduled_for"`
	ScheduledEnd types.String `tfsdk:"scheduled_end"`
}

func NewIncidentResource() resource.Resource {
	return &IncidentResource{}
}

func (r *IncidentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incident"
}

func (r *IncidentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a StatusGator incident or maintenance window.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the incident.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"board_id": schema.StringAttribute{
				Description: "The ID of the board this incident belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				Description: "The title of the incident.",
				Required:    true,
			},
			"message": schema.StringAttribute{
				Description: "The message/description of the incident.",
				Required:    true,
			},
			"severity": schema.StringAttribute{
				Description: "Severity of the incident: minor, major, or maintenance.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("minor", "major", "maintenance"),
				},
			},
			"phase": schema.StringAttribute{
				Description: "Phase of the incident: investigating, identified, monitoring, resolved, scheduled, in_progress, verifying, or completed.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("investigating", "identified", "monitoring", "resolved", "scheduled", "in_progress", "verifying", "completed"),
				},
			},
			"monitor_ids": schema.ListAttribute{
				Description: "List of affected monitor IDs.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"scheduled_for": schema.StringAttribute{
				Description: "Start time for scheduled maintenance (RFC3339 format).",
				Optional:    true,
			},
			"scheduled_end": schema.StringAttribute{
				Description: "End time for scheduled maintenance (RFC3339 format).",
				Optional:    true,
			},
		},
	}
}

func (r *IncidentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IncidentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data IncidentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := &statusgator.IncidentRequest{
		Title:    data.Title.ValueString(),
		Message:  data.Message.ValueString(),
		Severity: statusgator.IncidentSeverity(data.Severity.ValueString()),
		Phase:    statusgator.IncidentPhase(data.Phase.ValueString()),
	}

	// Convert monitor IDs
	if !data.MonitorIDs.IsNull() {
		var monitorIDs []string
		resp.Diagnostics.Append(data.MonitorIDs.ElementsAs(ctx, &monitorIDs, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.MonitorIDs = monitorIDs
	}

	// Parse scheduled times
	if !data.ScheduledFor.IsNull() && data.ScheduledFor.ValueString() != "" {
		t, err := time.Parse(time.RFC3339, data.ScheduledFor.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid scheduled_for Format",
				fmt.Sprintf("Could not parse scheduled_for as RFC3339: %s", err.Error()),
			)
			return
		}
		createReq.ScheduledFor = &t
	}

	if !data.ScheduledEnd.IsNull() && data.ScheduledEnd.ValueString() != "" {
		t, err := time.Parse(time.RFC3339, data.ScheduledEnd.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid scheduled_end Format",
				fmt.Sprintf("Could not parse scheduled_end as RFC3339: %s", err.Error()),
			)
			return
		}
		createReq.ScheduledEnd = &t
	}

	tflog.Debug(ctx, "Creating incident", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"title":    data.Title.ValueString(),
		"severity": data.Severity.ValueString(),
	})

	incident, err := r.client.Incidents.Create(ctx, data.BoardID.ValueString(), createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating StatusGator Incident",
			fmt.Sprintf("Could not create incident: %s", err.Error()),
		)
		return
	}

	r.mapIncidentToModel(ctx, incident, &data, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IncidentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IncidentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading incident", map[string]interface{}{
		"board_id":    data.BoardID.ValueString(),
		"incident_id": data.ID.ValueString(),
	})

	incidents, err := r.client.Incidents.ListAll(ctx, data.BoardID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading StatusGator Incident",
			fmt.Sprintf("Could not list incidents: %s", err.Error()),
		)
		return
	}

	var found bool
	for _, inc := range incidents {
		if inc.ID == data.ID.ValueString() {
			found = true
			r.mapIncidentToModel(ctx, &inc, &data, &resp.Diagnostics)
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IncidentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data IncidentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Incidents are updated by adding updates, not by modifying the original
	updateReq := &statusgator.IncidentUpdateRequest{
		Message: data.Message.ValueString(),
		Phase:   statusgator.IncidentPhase(data.Phase.ValueString()),
	}

	tflog.Debug(ctx, "Updating incident", map[string]interface{}{
		"board_id":    data.BoardID.ValueString(),
		"incident_id": data.ID.ValueString(),
	})

	_, err := r.client.Incidents.AddUpdate(ctx, data.BoardID.ValueString(), data.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating StatusGator Incident",
			fmt.Sprintf("Could not update incident ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *IncidentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data IncidentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Deleting incident (resolving)", map[string]interface{}{
		"board_id":    data.BoardID.ValueString(),
		"incident_id": data.ID.ValueString(),
	})

	// Incidents cannot be deleted, so we resolve them instead
	resolveReq := &statusgator.IncidentUpdateRequest{
		Message: "Incident resolved via Terraform",
		Phase:   statusgator.IncidentPhaseResolved,
	}

	_, err := r.client.Incidents.AddUpdate(ctx, data.BoardID.ValueString(), data.ID.ValueString(), resolveReq)
	if err != nil {
		if statusgator.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError(
			"Error Resolving StatusGator Incident",
			fmt.Sprintf("Could not resolve incident ID %s: %s", data.ID.ValueString(), err.Error()),
		)
	}
}

func (r *IncidentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importBoardChildState(ctx, req, resp)
}

func (r *IncidentResource) mapIncidentToModel(ctx context.Context, incident *statusgator.Incident, data *IncidentResourceModel, diags *diag.Diagnostics) {
	data.ID = types.StringValue(incident.ID)
	data.Title = types.StringValue(incident.Title)
	data.Severity = types.StringValue(string(incident.Severity))
	data.Phase = types.StringValue(string(incident.Phase))

	// Get the latest update message if available
	if len(incident.Updates) > 0 {
		data.Message = types.StringValue(incident.Updates[0].Message)
	}

	// Convert monitor IDs
	if len(incident.MonitorIDs) > 0 {
		monitorElements := make([]attr.Value, len(incident.MonitorIDs))
		for i, id := range incident.MonitorIDs {
			monitorElements[i] = types.StringValue(id)
		}
		monitorList, d := types.ListValue(types.StringType, monitorElements)
		diags.Append(d...)
		data.MonitorIDs = monitorList
	} else {
		data.MonitorIDs = types.ListNull(types.StringType)
	}

	// Convert scheduled times
	if incident.ScheduledFor != nil {
		data.ScheduledFor = types.StringValue(incident.ScheduledFor.Format(time.RFC3339))
	} else {
		data.ScheduledFor = types.StringNull()
	}
	if incident.ScheduledEnd != nil {
		data.ScheduledEnd = types.StringValue(incident.ScheduledEnd.Format(time.RFC3339))
	} else {
		data.ScheduledEnd = types.StringNull()
	}
}
