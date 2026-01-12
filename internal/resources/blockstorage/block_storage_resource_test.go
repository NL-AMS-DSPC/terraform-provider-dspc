package blockstorage

import (
	"context"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	tfresource "github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBlockStorageResource_Schema(t *testing.T) {
	ctx := context.Background()
	r := NewBlockStorageResource()

	schemaReq := tfresource.SchemaRequest{}
	schemaResp := &tfresource.SchemaResponse{}

	r.Schema(ctx, schemaReq, schemaResp)

	assert.Contains(t, schemaResp.Schema.Attributes, "name")
	assert.Contains(t, schemaResp.Schema.Attributes, "size")
}

func TestBlockStorageResource_Metadata(t *testing.T) {
	ctx := context.Background()
	r := NewBlockStorageResource()

	req := tfresource.MetadataRequest{
		ProviderTypeName: "dspc",
	}
	resp := &tfresource.MetadataResponse{}

	r.Metadata(ctx, req, resp)

	assert.Equal(t, "dspc_block", resp.TypeName)
}

func TestBlockStorageResource_Configure(t *testing.T) {
	ctx := context.Background()
	r := &BlockStorageResource{}
	mockClient := &mockBlockStorageClient{}

	req := tfresource.ConfigureRequest{
		ProviderData: mockClient,
	}
	resp := &tfresource.ConfigureResponse{}

	r.Configure(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.Equal(t, mockClient, r.client)
}

func TestBlockStorageResource_Create(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockBlockStorageClient{}
	r := &BlockStorageResource{client: mockClient}

	mockClient.On("CreateBlock", ctx, client.CreateBlockRequest{
		Name: "test-block",
	}).Return(&client.CreateBlockResponse{
		Created: "test-block-created",
	}, nil)

	resp, err := r.client.CreateBlock(ctx, client.CreateBlockRequest{
		Name: "test-block",
	})

	assert.NoError(t, err)
	assert.Equal(t, "test-block-created", resp.Created)
	mockClient.AssertExpectations(t)
}

func TestBlockStorageResource_Read(t *testing.T) {
	ctx := context.Background()
	mockClient := &mockBlockStorageClient{}
	r := &BlockStorageResource{client: mockClient}

	blockName := "test-block"

	mockClient.On("GetBlock", ctx, blockName).Return(&client.Block{
		Name: blockName,
		Size: "10Gi",
	}, nil)

	state := tfsdk.State{
		Schema: schema.Schema{
			Attributes: map[string]schema.Attribute{
				"name": schema.StringAttribute{},
				"size": schema.StringAttribute{},
			},
		},
	}
	diagnostics := state.Set(ctx, &BlockResourceModel{
		Name: types.StringValue(blockName),
	})

	req := tfresource.ReadRequest{
		State: state,
	}

	resp := &tfresource.ReadResponse{
		Diagnostics: diagnostics,
		State: tfsdk.State{
			Schema: schema.Schema{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{},
					"size": schema.StringAttribute{},
				},
			},
		},
	}

	r.Read(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	mockClient.AssertExpectations(t)
}

type mockBlockStorageClient struct {
	mock.Mock
}

func (m *mockBlockStorageClient) UpdateBlock(ctx context.Context, blockName string, req client.UpdateBlockRequest) (*client.UpdateBlockResponse, error) {
	args := m.Called(ctx, blockName, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.UpdateBlockResponse), args.Error(1)
}

func (m *mockBlockStorageClient) CreateBlock(ctx context.Context, req client.CreateBlockRequest) (*client.CreateBlockResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.CreateBlockResponse), args.Error(1)
}

func (m *mockBlockStorageClient) GetBlock(ctx context.Context, blockName string) (*client.Block, error) {
	args := m.Called(ctx, blockName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*client.Block), args.Error(1)
}

func (m *mockBlockStorageClient) DeleteBlock(ctx context.Context, blockName string) error {
	args := m.Called(ctx, blockName)
	return args.Error(0)
}
