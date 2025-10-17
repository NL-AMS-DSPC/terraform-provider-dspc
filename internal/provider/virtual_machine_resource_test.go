package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
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
			mockResponse: CreateVMResponse{
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
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create resource with mock client
			vmResource := &VirtualMachineResource{
				client: NewClient(server.URL, "test-api-key", 30),
			}

			// Test the client directly instead of the resource methods
			vm, err := vmResource.client.CreateVM(context.Background(), tt.vmName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if vm.Name != tt.vmName {
					t.Errorf("Expected VM name %s, got %s", tt.vmName, vm.Name)
				}
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
			mockResponse: DeleteVMResponse{
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
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create resource with mock client
			vmResource := &VirtualMachineResource{
				client: NewClient(server.URL, "test-api-key", 30),
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
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:     "successful import",
			importID: "test-vm",
			mockResponse: []*VM{
				{Name: "test-vm"},
				{Name: "other-vm"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:     "import non-existent VM",
			importID: "nonexistent-vm",
			mockResponse: []*VM{
				{Name: "other-vm"},
			},
			mockStatusCode: http.StatusOK,
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
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/virtualmachine" {
					t.Errorf("Expected /virtualmachine path, got %s", r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create resource with mock client
			vmResource := &VirtualMachineResource{
				client: NewClient(server.URL, "test-api-key", 30),
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
	vmResource := &VirtualMachineResource{}

	req := resource.UpdateRequest{}
	resp := &resource.UpdateResponse{}

	vmResource.Update(context.Background(), req, resp)

	// Update should always return an error
	if !resp.Diagnostics.HasError() {
		t.Error("Expected error from Update, got none")
	}
}
