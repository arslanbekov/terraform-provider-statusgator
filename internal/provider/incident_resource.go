package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	ID          types.String `tfsdk:"id"`
	BoardID     types.String `tfsdk:"board_id"`
	Name        types.String `tfsdk:"name"`
	Details     types.String `tfsdk:"details"`
	Severity    types.String `tfsdk:"severity"`
	Phase       types.String `tfsdk:"phase"`
	WillStartAt types.String `tfsdk:"will_start_at"`
	WillEndAt   types.String `tfsdk:"will_end_at"`
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
			"name": schema.StringAttribute{
				Description: "The name of the incident.",
				Required:    true,
			},
			"details": schema.StringAttribute{
				Description: "The details/description of the incident.",
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
			"will_start_at": schema.StringAttribute{
				Description: "Start time for scheduled maintenance (RFC3339 format).",
				Optional:    true,
			},
			"will_end_at": schema.StringAttribute{
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
		Name:     data.Name.ValueString(),
		Details:  data.Details.ValueString(),
		Severity: statusgator.IncidentSeverity(data.Severity.ValueString()),
		Phase:    statusgator.IncidentPhase(data.Phase.ValueString()),
	}

	// Parse scheduled times
	if !data.WillStartAt.IsNull() && data.WillStartAt.ValueString() != "" {
		t, err := time.Parse(time.RFC3339, data.WillStartAt.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid will_start_at Format",
				fmt.Sprintf("Could not parse will_start_at as RFC3339: %s", err.Error()),
			)
			return
		}
		createReq.WillStartAt = &t
	}

	if !data.WillEndAt.IsNull() && data.WillEndAt.ValueString() != "" {
		t, err := time.Parse(time.RFC3339, data.WillEndAt.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid will_end_at Format",
				fmt.Sprintf("Could not parse will_end_at as RFC3339: %s", err.Error()),
			)
			return
		}
		createReq.WillEndAt = &t
	}

	tflog.Debug(ctx, "Creating incident", map[string]interface{}{
		"board_id": data.BoardID.ValueString(),
		"name":     data.Name.ValueString(),
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
		Details: data.Details.ValueString(),
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
		Details: "Incident resolved via Terraform",
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
	data.Name = types.StringValue(incident.Name)
	data.Details = types.StringValue(incident.Details)
	data.Severity = types.StringValue(string(incident.Severity))
	data.Phase = types.StringValue(string(incident.Phase))

	// Convert scheduled times
	if incident.WillStartAt != nil {
		data.WillStartAt = types.StringValue(incident.WillStartAt.Format(time.RFC3339))
	} else {
		data.WillStartAt = types.StringNull()
	}
	if incident.WillEndAt != nil {
		data.WillEndAt = types.StringValue(incident.WillEndAt.Format(time.RFC3339))
	} else {
		data.WillEndAt = types.StringNull()
	}
}
