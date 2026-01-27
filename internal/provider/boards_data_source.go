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

var _ datasource.DataSource = &BoardsDataSource{}

type BoardsDataSource struct {
	client *statusgator.Client
}

type BoardsDataSourceModel struct {
	Boards []BoardModel `tfsdk:"boards"`
}

type BoardModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	PublicToken types.String `tfsdk:"public_token"`
}

func NewBoardsDataSource() datasource.DataSource {
	return &BoardsDataSource{}
}

func (d *BoardsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_boards"
}

func (d *BoardsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all StatusGator boards in the account.",
		Attributes: map[string]schema.Attribute{
			"boards": schema.ListNestedAttribute{
				Description: "List of boards.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the board.",
							Computed:    true,
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
				},
			},
		},
	}
}

func (d *BoardsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *BoardsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BoardsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Reading all boards")

	boards, err := d.client.Boards.ListAll(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to List StatusGator Boards",
			fmt.Sprintf("Could not list boards: %s", err.Error()),
		)
		return
	}

	data.Boards = make([]BoardModel, len(boards))
	for i, board := range boards {
		data.Boards[i] = BoardModel{
			ID:          types.StringValue(board.ID),
			Name:        types.StringValue(board.Name),
			PublicToken: types.StringValue(board.PublicToken),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
