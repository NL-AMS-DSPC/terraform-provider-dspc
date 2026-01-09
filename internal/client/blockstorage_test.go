package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockStorageService_CreateAttachment(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		vmName         string
		blockName      string
	}{
		{
			name:           "successfully create a block storage attachment",
			mockStatusCode: http.StatusOK,
			mockResponse: BlockStorageAttachment{
				BlockName: "pvc-test-1",
				VMName:    "vm-test-1",
			},
			blockName:   "pvc-test-1",
			vmName:      "vm-test-1",
			expectError: false,
		},
		{
			name:           "vm not found",
			mockResponse:   map[string]string{"status": "VM not found"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "PVC already attached to another VM",
			mockResponse:   map[string]string{"status": "PVC already attached to another VM"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:           "PVC not found",
			mockResponse:   map[string]string{"status": "PVC not found"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			client := NewDspcClient(server.URL, "test-api-key", 30).BlockStorage

			attachment, err := client.CreateAttachment(t.Context(), "pvc-test-1", "vm-test-1")
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, attachment.BlockName, tt.blockName)
				assert.Equal(t, attachment.VMName, tt.vmName)
			}
		})
	}
}
