package blockstorage

import (
	"context"
	"fmt"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &BlockStorageAttachmentDataSource{}
	_ datasource.DataSourceWithConfigure = &BlockStorageAttachmentDataSource{}
)

// BlockStorageAttachmentDataClient defines an interface for interacting with block storage attachment data operations.
// GetAttachment retrieves a specific block storage attachment from the data source.
type BlockStorageAttachmentDataClient interface {
	GetAttachment(ctx context.Context, blockName, vmName string) (*client.BlockStorageAttachment, error)
}

// BlockStorageAttachmentDataSource defines the data source implementation.
type BlockStorageAttachmentDataSource struct {
	client BlockStorageAttachmentDataClient
}

// BlockStorageAttachmentDataSourceModel describes the data source data model.
type BlockStorageAttachmentDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	BlockStorageName types.String `tfsdk:"block_storage_name"`
	VMName           types.String `tfsdk:"vm_name"`
}

// NewBlockStorageAttachmentDataSource creates a new BlockStorageAttachmentDataSource.
func NewBlockStorageAttachmentDataSource() datasource.DataSource {
	return &BlockStorageAttachmentDataSource{}
}

// Metadata updates the provided metadata with the data source type name.
func (d *BlockStorageAttachmentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block_storage_attachment"
}

// Schema updates the data source schema with the attributes for the data source.
func (d *BlockStorageAttachmentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves a specific block storage attachment in the DSPC platform.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier for the block storage attachment.",
				Computed:    true,
			},
			"block_storage_name": schema.StringAttribute{
				Description: "The name of the block storage.",
				Required:    true,
			},
			"vm_name": schema.StringAttribute{
				Description: "The name of the virtual machine.",
				Required:    true,
			},
		},
	}
}

// Configure creates a new API client and stores it in the response data for the data source to use.
func (d *BlockStorageAttachmentDataSource) Configure(
	_ context.Context,
	req datasource.ConfigureRequest,
	resp *datasource.ConfigureResponse,
) {
	if req.ProviderData == nil {
		return
	}

	dataClient, ok := req.ProviderData.(*client.DspcClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.DspcClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	if dataClient.BlockStorage == nil {
		resp.Diagnostics.AddError("Unexpected datasource configuration error",
			"Expected blockstorage service to be ready. Please report this issue to the provider developers.",
		)
		return
	}

	d.client = dataClient.BlockStorage
}

// Read reads the data from the API and stores it in the state.
func (d *BlockStorageAttachmentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config BlockStorageAttachmentDataSourceModel

	// Read configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get the specific attachment from the API
	attachment, err := d.client.GetAttachment(ctx, config.BlockStorageName.ValueString(), config.VMName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading block storage attachment",
			fmt.Sprintf("Could not read block storage attachment: %s", err.Error()),
		)
		return
	}

	// Set state with retrieved data
	var state BlockStorageAttachmentDataSourceModel
	state.ID = types.StringValue(createStateId(attachment.BlockName, attachment.VMName))
	state.BlockStorageName = types.StringValue(attachment.BlockName)
	state.VMName = types.StringValue(attachment.VMName)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
