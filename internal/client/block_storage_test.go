package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// createMockAuthServer creates a mock Keycloak authentication server for testing
func createMockAuthServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{ // nolint:gosec
			"access_token": "mock-jwt",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
}

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
			mockResponse: CreateBlockAttachmentResponse{
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
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			// Create mock server
			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).BlockStorage

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

func TestBlockStorageService_GetAttachment(t *testing.T) {
	tests := []struct {
		name              string
		mockResponse      interface{}
		mockStatusCode    int
		expectedBlockName string
		expectedVMName    string
		expectError       bool
		expectedError     string
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
			expectedBlockName: "pvc-test-1",
			expectedVMName:    "vm-test-1",
		},
		{
			name:           "vm not found",
			mockResponse:   "VM vm-test-1 not found: some error",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedError:  "API error 500: \"VM vm-test-1 not found: some error\"\n",
		},
		{
			name: "PVC not found",
			mockResponse: []map[string]any{
				{
					"name":         "pvc-test-2",
					"namespace":    "test-namespace",
					"attachedToVM": "vm-test-2",
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectedError:  "attachment not found for block pvc-test-1 on VM vm-test-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()
			client := newTestDspcClient(server.URL, authServer.URL).BlockStorage
			attachment, err := client.GetAttachment(t.Context(), "pvc-test-1", "vm-test-1")
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, attachment.BlockName, tt.expectedBlockName)
				assert.Equal(t, attachment.VMName, tt.expectedVMName)
			}
		})
	}
}

func TestBlockStorageService_DeleteAttachment(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedError  string
	}{
		{
			name:           "successfully delete a block storage attachment",
			mockStatusCode: http.StatusOK,
			mockResponse:   map[string]string{},
			expectError:    false,
		},
		{
			name:           "vm not found",
			mockResponse:   "VM vm-test-1 not found: some error",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedError:  "API error 500: \"VM vm-test-1 not found: some error\"\n",
		},
		{
			name:           "PVC not attached to VM",
			mockResponse:   "PVC pvc-test-1 is not attached to VM vm-test-1",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedError:  "API error 500: \"PVC pvc-test-1 is not attached to VM vm-test-1\"\n",
		},
		{
			name:           "could not detach PVC from VM",
			mockResponse:   "could not detach PVC from VM: internal error",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedError:  "API error 500: \"could not detach PVC from VM: internal error\"\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).BlockStorage

			err := client.DeleteAttachment(t.Context(), "pvc-test-1", "vm-test-1")
			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
