package blockstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/stretchr/testify/assert"
)

func TestBlockStorageDataSource_Read(t *testing.T) {
	type blockModel struct {
		Name string `json:"name"`
		Size string `json:"size"`
	}

	tests := []struct {
		name           string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successfully get all blocks",
			mockResponse: []blockModel{
				{
					Name: "test-block-1",
					Size: "500mb",
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error",
			mockResponse:   "Internal server error",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/v1/namespaces/ns/pvcs", r.URL.Path)

				// Check Authorization header
				assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

				// Check Content-Type header
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create data source with mock client
			dataSource := &BlockStorageDataSource{
				client: client.NewDspcClient(server.URL, "ns", "test-api-key", 30).
					BlockStorage,
			}

			req := datasource.ReadRequest{}
			resp := &datasource.ReadResponse{
				State: tfsdk.State{
					Schema: schema.Schema{
						Attributes: map[string]schema.Attribute{
							"blocks": schema.ListNestedAttribute{
								Description: "List of blocks.",
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{},
										"size": schema.StringAttribute{},
									},
								},
							},
						},
					},
				},
			}

			dataSource.Read(context.Background(), req, resp)

			if tt.expectError {
				assert.True(t, resp.Diagnostics.HasError())
			} else {
				fmt.Println("errors", resp.Diagnostics.Errors())
				assert.False(t, resp.Diagnostics.HasError())

				var blocks BlockStorageDataSourceModel
				resp.State.Get(context.Background(), &blocks)

				assert.NotNil(t, blocks)
				assert.Equal(t, "test-block-1", blocks.Blocks[0].Name.ValueString())
				assert.Equal(t, "500mb", blocks.Blocks[0].Size.ValueString())
			}
		})
	}
}

func TestBlockStorageDataSource_Metadata(t *testing.T) {
	dataSource := &BlockStorageDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "dspc",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)

	assert.Equal(t, "dspc_blocks", resp.TypeName)
}

func TestBlockStorageDataSource_Schema(t *testing.T) {
	dataSource := &BlockStorageDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema.Attributes)

	// Check that required attributes exist
	attributes := resp.Schema.Attributes
	assert.Contains(t, attributes, "blocks")
}

func TestBlockStorageDataSource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData any
		expectError  bool
	}{
		{
			name:         "valid client",
			providerData: client.NewDspcClient("test-api-key", "ns", "test-api-key", 30).BlockStorage,
			expectError:  false,
		},
		{
			name:         "nil provider data",
			providerData: nil,
			expectError:  false, // Should not error, just skip configuration
		},
		{
			name:         "invalid provider data type",
			providerData: "not-a-client",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataSource := &BlockStorageDataSource{}

			req := datasource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			dataSource.Configure(context.Background(), req, resp)

			if tt.expectError {
				assert.True(t, resp.Diagnostics.HasError())
			} else {
				assert.False(t, resp.Diagnostics.HasError())
			}
		})
	}
}

func TestNewBlockStorageDataSource(t *testing.T) {
	dataSource := NewBlockStorageDataSource()
	assert.NotNil(t, dataSource)
}
