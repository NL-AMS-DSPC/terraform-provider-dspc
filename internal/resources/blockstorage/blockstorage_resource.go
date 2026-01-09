package resources

import (
	"context"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
// var (
//
//	_ resource.Resource                = &VMResource{}
//	_ resource.ResourceWithConfigure   = &VMResource{}
//	_ resource.ResourceWithImportState = &VMResource{}
//
// )
// TODO: name BlockStorage

type blockStorageClient interface {
	CreateBlock()
}

type BlockResource struct {
	client *provider.Client
}

func NewBlockResource() resource.Resource { return &BlockResource{} }

func (r *BlockResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_block"
}

func (r *BlockResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Block in the DSPC platform",
		// TODO: check if correct
		Attributes: map[string]schema.Attribute{
			"id":   schema.StringAttribute{},
			"size": schema.Int64Attribute{},
		},
	}
}

func (r *BlockResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *BlockResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *BlockResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	//TODO implement me
	panic("implement me")
}

func (r *BlockResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	//TODO implement me
	panic("implement me")
}

type BlockModel struct {
	ID   types.String `tfsdk:"id"`
	Size types.Int64  `tfsdk:"size"`
}

// name
// size
// storageClass
// accessMode ->
// volumeMode
