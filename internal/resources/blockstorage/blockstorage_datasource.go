package resources

import (
	"context"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

type blockDataClient interface {
	ListBlocks(ctx context.Context) ([]*client.Block, error)
}

type BlockDataSource struct {
	client blockDataClient
}

type BlockDataSourceModel struct {
	Blocks []BlockModel `tfsdk:"blocks""`
}

func NewBlockDataSource() *BlockDataSourceModel {
	return &BlockDataSourceModel{}
}

func (d *BlockDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_virtual_machines"
}
