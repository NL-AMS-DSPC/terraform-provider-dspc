package blockstorage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestBlockStorageAttachmentDataSource_Read(t *testing.T) {
	tests := []struct {
		name              string
		mockResponse      any
		mockStatusCode    int
		expectError       bool
		expectedBlockName string
		expectedVMName    string
	}{
		{
			name: "successfully get a block storage attachment",
			mockResponse: []map[string]any{
				{
					"name":         "pvc-test-1",
					"namespace":    "test-namespace",
					"attachedToVM": "vm-test-1",
				},
			},
			mockStatusCode:    http.StatusOK,
			expectError:       false,
			expectedBlockName: "pvc-test-1",
			expectedVMName:    "vm-test-1",
		},
		{
			name:           "attachment not found",
			mockResponse:   "VM vm-test-1 not found: some error",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name: "PVC not attached to specified VM",
			mockResponse: []map[string]any{
				{
					"name":         "pvc-test-2",
					"namespace":    "test-namespace",
					"attachedToVM": "vm-test-2",
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "/virtualmachines/vm-test-1/pvcs", r.URL.Path)

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
			dataSource := &BlockStorageAttachmentDataSource{
				client: client.NewDspcClient(server.URL, "test-api-key", 30).
					BlockStorage,
			}

			// Test the client directly instead of the data source methods
			attachment, err := dataSource.client.GetAttachment(context.Background(), "pvc-test-1", "vm-test-1")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBlockName, attachment.BlockName)
				assert.Equal(t, tt.expectedVMName, attachment.VMName)
			}
		})
	}
}

func TestBlockStorageAttachmentDataSource_Metadata(t *testing.T) {
	dataSource := &BlockStorageAttachmentDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "dspc",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)

	assert.Equal(t, "dspc_block_storage_attachment", resp.TypeName)
}

func TestBlockStorageAttachmentDataSource_Schema(t *testing.T) {
	dataSource := &BlockStorageAttachmentDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(context.Background(), req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.NotNil(t, resp.Schema.Attributes)

	// Check that required attributes exist
	attributes := resp.Schema.Attributes
	assert.Contains(t, attributes, "id")
	assert.Contains(t, attributes, "block_storage_name")
	assert.Contains(t, attributes, "vm_name")
}

func TestBlockStorageAttachmentDataSource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData any
		expectError  bool
	}{
		{
			name:         "valid client",
			providerData: client.NewDspcClient("test-api-key", "test-api-key", 30).BlockStorage,
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
			dataSource := &BlockStorageAttachmentDataSource{}

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

func TestNewBlockStorageAttachmentDataSource(t *testing.T) {
	dataSource := NewBlockStorageAttachmentDataSource()
	assert.NotNil(t, dataSource)
}

type apiMock struct {
}

func (api *apiMock) Create(ctx context.Context, path string, body interface{}, out interface{}) error {
	return nil
}

func (api *apiMock) Get(ctx context.Context, path string, out interface{}) error {
	return nil
}

func (api *apiMock) Delete(ctx context.Context, path string) error {
	return nil
}
