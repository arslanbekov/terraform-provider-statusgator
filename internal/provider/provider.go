package provider

import (
	"context"
	"os"
	"time"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ provider.Provider = &StatusGatorProvider{}

type StatusGatorProvider struct {
	version string
}

type StatusGatorProviderModel struct {
	APIToken types.String `tfsdk:"api_token"`
	BaseURL  types.String `tfsdk:"base_url"`
	Timeout  types.Int64  `tfsdk:"timeout"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &StatusGatorProvider{
			version: version,
		}
	}
}

func (p *StatusGatorProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "statusgator"
	resp.Version = p.version
}

func (p *StatusGatorProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with StatusGator API to manage status pages, monitors, and incidents.",
		Attributes: map[string]schema.Attribute{
			"api_token": schema.StringAttribute{
				Description: "StatusGator API token. Can also be set via STATUSGATOR_API_TOKEN environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
			"base_url": schema.StringAttribute{
				Description: "Custom API base URL. Defaults to https://statusgator.com/api/v3",
				Optional:    true,
			},
			"timeout": schema.Int64Attribute{
				Description: "Request timeout in seconds. Defaults to 30.",
				Optional:    true,
			},
		},
	}
}

func (p *StatusGatorProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring StatusGator client")

	var config StatusGatorProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check for unknown values
	if config.APIToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Unknown StatusGator API Token",
			"The provider cannot create the StatusGator API client as there is an unknown configuration value for the API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the STATUSGATOR_API_TOKEN environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Default values from environment
	apiToken := os.Getenv("STATUSGATOR_API_TOKEN")
	baseURL := os.Getenv("STATUSGATOR_BASE_URL")
	if baseURL == "" {
		baseURL = statusgator.DefaultBaseURL
	}
	timeout := 30 * time.Second

	// Override with config values
	if !config.APIToken.IsNull() {
		apiToken = config.APIToken.ValueString()
	}
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}
	if !config.Timeout.IsNull() {
		timeout = time.Duration(config.Timeout.ValueInt64()) * time.Second
	}

	// Validate
	if apiToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("api_token"),
			"Missing StatusGator API Token",
			"The provider cannot create the StatusGator API client as there is a missing or empty value for the API token. "+
				"Set the api_token value in the configuration or use the STATUSGATOR_API_TOKEN environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating StatusGator client", map[string]interface{}{
		"base_url": baseURL,
		"timeout":  timeout.String(),
	})

	// Create client with options
	client, err := statusgator.NewClient(
		apiToken,
		statusgator.WithBaseURL(baseURL),
		statusgator.WithTimeout(timeout),
		statusgator.WithUserAgent("terraform-provider-statusgator/"+p.version),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create StatusGator API Client",
			"An unexpected error occurred when creating the StatusGator API client: "+err.Error(),
		)
		return
	}

	// Verify connection
	if err := client.Ping(ctx); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Connect to StatusGator API",
			"Failed to verify API connection. Please check your API token and network connectivity: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "StatusGator client configured successfully")

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *StatusGatorProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMonitorGroupResource,
		NewWebsiteMonitorResource,
		NewPingMonitorResource,
		NewCustomMonitorResource,
		NewServiceMonitorResource,
		NewSubscriberResource,
		NewIncidentResource,
	}
}

func (p *StatusGatorProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewBoardDataSource,
		NewBoardsDataSource,
	}
}
