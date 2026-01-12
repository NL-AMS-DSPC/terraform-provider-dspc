package tests

import (
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/suite"
)

type BlockStorageResourceSuite struct {
	MockProvider
}

func TestBlockStorageProvisioning(t *testing.T) {
	suite.Run(t, new(BlockStorageResourceSuite))
}

func (s *BlockStorageResourceSuite) TestBlockStorageResource_Delete() {
	s.Handlers = map[string]func() MockResponse{
		// Block exists (create/read)
		"GET /v1/namespaces/test-ns/pvcs/test-block": func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusOK,
				ResponseBody: struct {
					Name string
					Size string
				}{
					Name: "test-block",
					Size: "5Gi",
				},
			}
		},
		// On delete, simulate either success or "not found" as needed.
		"DELETE /v1/namespaces/test-ns/pvcs/test-block": func() MockResponse {
			return MockResponse{
				ResponseCode: http.StatusNoContent,
				ResponseBody: nil,
			}
		},
	}

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create resource
			{
				Config: getProvider(s.Server.URL, `
resource "dspc_block_storage" "test" {
  name = "test-block"
}
`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("dspc_block_storage.test", "name", "test-block"),
					resource.TestCheckResourceAttr("dspc_block_storage.test", "size", "5Gi"),
				),
			},
			// Step 2: Delete resource (empty config => resource should be destroyed)
			{
				Config: getProvider(s.Server.URL, ""),                                       // No resource block, triggers destroy
				Check:  resource.TestCheckNoResourceAttr("dspc_block_storage.test", "name"), // Ensure it's gone from state
				// Optionally: check with CheckDestroy helper
			},
		},
	})
}

//func TestBlockStorageResource_Schema(t *testing.T) {
//	ctx := context.Background()
//	r := NewBlockStorageResource()
//
//	schemaReq := tfresource.SchemaRequest{}
//	schemaResp := &tfresource.SchemaResponse{}
//
//	r.Schema(ctx, schemaReq, schemaResp)
//
//	assert.Contains(t, schemaResp.Schema.Attributes, "name")
//	assert.Contains(t, schemaResp.Schema.Attributes, "size")
//}
//
//func TestBlockStorageResource_Metadata(t *testing.T) {
//	ctx := context.Background()
//	r := NewBlockStorageResource()
//
//	req := tfresource.MetadataRequest{
//		ProviderTypeName: "dspc",
//	}
//	resp := &tfresource.MetadataResponse{}
//
//	r.Metadata(ctx, req, resp)
//
//	assert.Equal(t, "dspc_block", resp.TypeName)
//}
//
//func TestBlockStorageResource_Configure(t *testing.T) {
//	ctx := context.Background()
//	r := &BlockStorageResource{}
//	mockClient := &mockBlockStorageClient{}
//
//	req := tfresource.ConfigureRequest{
//		ProviderData: mockClient,
//	}
//	resp := &tfresource.ConfigureResponse{}
//
//	r.Configure(ctx, req, resp)
//
//	assert.False(t, resp.Diagnostics.HasError())
//	assert.Equal(t, mockClient, r.client)
//}
//
//func TestBlockStorageResource_Create(t *testing.T) {
//	ctx := context.Background()
//	mockClient := &mockBlockStorageClient{}
//	r := &BlockStorageResource{client: mockClient}
//
//	mockClient.On("CreateBlock", ctx, client.CreateBlockRequest{
//		Name: "test-block",
//	}).Return(&client.CreateBlockResponse{
//		Created: "test-block-created",
//	}, nil)
//
//	resp, err := r.client.CreateBlock(ctx, client.CreateBlockRequest{
//		Name: "test-block",
//	})
//
//	assert.NoError(t, err)
//	assert.Equal(t, "test-block-created", resp.Created)
//	mockClient.AssertExpectations(t)
//}
//
//func TestBlockStorageResource_Read(t *testing.T) {
//	ctx := context.Background()
//	mockClient := &mockBlockStorageClient{}
//	r := &BlockStorageResource{client: mockClient}
//
//	blockName := "test-block"
//
//	mockClient.On("GetBlock", ctx, blockName).Return(&client.Block{
//		Name: blockName,
//		Size: "10Gi",
//	}, nil)
//
//	state := tfsdk.State{
//		Schema: schema.Schema{
//			Attributes: map[string]schema.Attribute{
//				"name": schema.StringAttribute{},
//				"size": schema.StringAttribute{},
//			},
//		},
//	}
//	diagnostics := state.Set(ctx, &BlockResourceModel{
//		Name: types.StringValue(blockName),
//	})
//
//	req := tfresource.ReadRequest{
//		State: state,
//	}
//
//	resp := &tfresource.ReadResponse{
//		Diagnostics: diagnostics,
//		State: tfsdk.State{
//			Schema: schema.Schema{
//				Attributes: map[string]schema.Attribute{
//					"name": schema.StringAttribute{},
//					"size": schema.StringAttribute{},
//				},
//			},
//		},
//	}
//
//	r.Read(ctx, req, resp)
//
//	assert.False(t, resp.Diagnostics.HasError())
//	mockClient.AssertExpectations(t)
//}
//
//type mockBlockStorageClient struct {
//	mock.Mock
//}
//
//func (m *mockBlockStorageClient) UpdateBlock(ctx context.Context, blockName string, req client.UpdateBlockRequest) (*client.UpdateBlockResponse, error) {
//	args := m.Called(ctx, blockName, req)
//	if args.Get(0) == nil {
//		return nil, args.Error(1)
//	}
//	return args.Get(0).(*client.UpdateBlockResponse), args.Error(1)
//}
//
//func (m *mockBlockStorageClient) CreateBlock(ctx context.Context, req client.CreateBlockRequest) (*client.CreateBlockResponse, error) {
//	args := m.Called(ctx, req)
//	if args.Get(0) == nil {
//		return nil, args.Error(1)
//	}
//	return args.Get(0).(*client.CreateBlockResponse), args.Error(1)
//}
//
//func (m *mockBlockStorageClient) GetBlock(ctx context.Context, blockName string) (*client.Block, error) {
//	args := m.Called(ctx, blockName)
//	if args.Get(0) == nil {
//		return nil, args.Error(1)
//	}
//	return args.Get(0).(*client.Block), args.Error(1)
//}
//
//func (m *mockBlockStorageClient) DeleteBlock(ctx context.Context, blockName string) error {
//	args := m.Called(ctx, blockName)
//	return args.Error(0)
//}
