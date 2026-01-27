package provider

import (
	"context"
	"fmt"

	"github.com/arslanbekov/statusgator-go-client/statusgator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &BoardDataSource{}

type BoardDataSource struct {
	client *statusgator.Client
}

type BoardDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PublicToken types.String `tfsdk:"public_token"`
}

func NewBoardDataSource() datasource.DataSource {
	return &BoardDataSource{}
}

func (d *BoardDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_board"
}

func (d *BoardDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details of a specific StatusGator board by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the board.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the board.",
				Computed:    true,
			},
			"public_token": schema.StringAttribute{
				Description: "The public access token for the board's status page.",
				Computed:    true,
			},
		},
	}
}

func (d *BoardDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*statusgator.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *statusgator.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *BoardDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BoardDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading board", map[string]interface{}{
		"board_id": data.ID.ValueString(),
	})

	board, err := d.client.Boards.Get(ctx, data.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read StatusGator Board",
			fmt.Sprintf("Could not read board ID %s: %s", data.ID.ValueString(), err.Error()),
		)
		return
	}

	data.ID = types.StringValue(board.ID)
	data.Name = types.StringValue(board.Name)
	data.PublicToken = types.StringValue(board.PublicToken)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
