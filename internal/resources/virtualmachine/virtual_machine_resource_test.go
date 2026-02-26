package virtualmachine

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

const (
	vmPath = "/v1/namespaces/test-ns/virtualmachines"
)

func TestVirtualMachineResource_Create(t *testing.T) {
	tests := []struct {
		name           string
		vmName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful creation",
			vmName: "test-vm",
			mockResponse: client.CreateVMResponse{
				Created: "test-vm",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error",
			vmName:         "test-vm",
			mockResponse:   map[string]string{"error": "VM already exists"},
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Keycloak auth server
			authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"access_token": "mock-jwt-token",
					"expires_in":   3600,
					"token_type":   "Bearer",
				})
			}))
			defer authServer.Close()

			// Create mock server that handles both POST (create) and GET (fetch)
			requestCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				requestCount++
				switch r.Method {
				case http.MethodPost:
					w.WriteHeader(tt.mockStatusCode)
					_ = json.NewEncoder(w).Encode(tt.mockResponse)
				case http.MethodGet:
					// Return a full VM response for the follow-up GET
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(&client.VM{
						Name: tt.vmName,
						SKU:  client.SKU{ID: "gp-2", Name: "General Purpose 2"},
					})
				}
			}))
			defer server.Close()

			// Create resource with mock client
			vmResource := &VMResource{
				client: client.NewDspcClient(
					server.URL,
					"test-ns",
					"test-client-id",
					"test-client-secret",
					authServer.URL,
					"test-realm",
					30,
				).VirtualMachines,
			}

			// Test the client directly instead of the resource methods
			vm, err := vmResource.client.CreateVM(context.Background(), tt.vmName, "gp-2")

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.vmName, vm.Name)
			}
		})
	}
}

func TestVirtualMachineResource_Delete(t *testing.T) {
	tests := []struct {
		name           string
		vmName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful deletion",
			vmName: "test-vm",
			mockResponse: client.DeleteVMResponse{
				Deleted: "test-vm",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error",
			vmName:         "nonexistent-vm",
			mockResponse:   map[string]string{"error": "VM not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock Keycloak auth server
			authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"access_token": "mock-jwt-token",
					"expires_in":   3600,
					"token_type":   "Bearer",
				})
			}))
			defer authServer.Close()

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create resource with mock client
			vmResource := &VMResource{
				client: client.NewDspcClient(
					server.URL,
					"test-ns",
					"test-client-id",
					"test-client-secret",
					authServer.URL,
					"test-realm",
					30,
				).VirtualMachines,
			}

			// Test the client directly instead of the resource methods
			err := vmResource.client.DeleteVM(context.Background(), tt.vmName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestVirtualMachineResource_ImportState(t *testing.T) {
	tests := []struct {
		name           string
		importID       string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:     "successful import",
			importID: "test-vm",
			mockResponse: &client.VM{
				Name: "test-vm",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "import non-existent VM",
			importID:       "nonexistent-vm",
			mockResponse:   "vm not found",
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "API error during import",
			importID:       "test-vm",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{
					"access_token": "mock-jwt-token",
					"expires_in":   3600,
					"token_type":   "Bearer",
				})
			}))
			defer authServer.Close()

			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET request, got %s", r.Method)
				}
				expectedPath := client.DefaultServiceConfig().VM.PathPrefix + vmPath + "/" + tt.importID
				if r.URL.Path != expectedPath {
					t.Fatalf("Expected %s path, got %s", expectedPath, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create resource with mock client
			vmResource := &VMResource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-client-id", "test-client-secret", authServer.URL, "test-realm", 30).VirtualMachines,
			}

			// Test the client directly instead of the resource methods
			vm, err := vmResource.client.GetVM(context.Background(), tt.importID)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if vm.Name != tt.importID {
					t.Errorf("Expected VM name %s, got %s", tt.importID, vm.Name)
				}
			}
		})
	}
}

func TestVirtualMachineResource_Update(t *testing.T) {
	vmResource := &VMResource{}

	req := resource.UpdateRequest{}
	resp := &resource.UpdateResponse{}

	vmResource.Update(context.Background(), req, resp)

	// Update should always return an error
	if !resp.Diagnostics.HasError() {
		t.Error("Expected error from Update, got none")
	}
}
