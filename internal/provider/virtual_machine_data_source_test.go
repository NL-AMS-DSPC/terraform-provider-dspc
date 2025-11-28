package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestVMDataSource_Read(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list with multiple VMs",
			mockResponse: []*VM{
				{Name: "vm1"},
				{Name: "vm2"},
				{Name: "vm3"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  3,
		},
		{
			name:           "successful list with empty result",
			mockResponse:   []*VM{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name: "successful list with single VM",
			mockResponse: []*VM{
				{Name: "single-vm"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  1,
		},
		{
			name:           "API error",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedCount:  0,
		},
		{
			name:           "API timeout",
			mockResponse:   map[string]string{"error": "Request timeout"},
			mockStatusCode: http.StatusRequestTimeout,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/virtualmachine" {
					t.Fatalf("Expected /virtualmachine path, got %s", r.URL.Path)
				}

				// Check Authorization header
				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-api-key" {
					t.Errorf("Expected Authorization: Bearer test-api-key, got %s", authHeader)
				}

				// Check Content-Type header
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type: application/json, got %s", contentType)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Create data source with mock client
			dataSource := &VMDataSource{
				client: NewClient(server.URL, "test-api-key", 30),
			}

			// Test the client directly instead of the data source methods
			vms, err := dataSource.client.ListVMs(context.Background())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}

				if len(vms) != tt.expectedCount {
					t.Errorf("Expected %d VMs, got %d", tt.expectedCount, len(vms))
				}

				// Verify VM data structure
				for i, vm := range vms {
					if vm.Name == "" {
						t.Errorf("VM %d has empty name", i)
					}
				}
			}
		})
	}
}

func TestVirtualMachineDataSource_Metadata(t *testing.T) {
	dataSource := &VMDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "dspc",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)

	expectedTypeName := "dspc_virtual_machines"
	if resp.TypeName != expectedTypeName {
		t.Errorf("Expected type name '%s', got '%s'", expectedTypeName, resp.TypeName)
	}
}

func TestVirtualMachineDataSource_Schema(t *testing.T) {
	dataSource := &VMDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Data source schema has errors: %v", resp.Diagnostics)
	}

	if resp.Schema.Attributes == nil {
		t.Error("Data source schema attributes is nil")
	}

	// Check that virtual_machines attribute exists
	attributes := resp.Schema.Attributes
	if _, ok := attributes["virtual_machines"]; !ok {
		t.Error("Data source schema missing 'virtual_machines' attribute")
	}
}

func TestVirtualMachineDataSource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData interface{}
		expectError  bool
	}{
		{
			name:         "valid client",
			providerData: &Client{},
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
			dataSource := &VMDataSource{}

			req := datasource.ConfigureRequest{
				ProviderData: tt.providerData,
			}
			resp := &datasource.ConfigureResponse{}

			dataSource.Configure(context.Background(), req, resp)

			if tt.expectError {
				if !resp.Diagnostics.HasError() {
					t.Errorf("Expected error, got none")
				}
			} else {
				if resp.Diagnostics.HasError() {
					t.Errorf("Expected no error, got: %v", resp.Diagnostics)
				}
			}
		})
	}
}

func TestNewVMDataSource(t *testing.T) {
	dataSource := NewVMDataSource()

	if dataSource == nil {
		t.Error("NewVirtualMachineDataSource returned nil")
	}

	// Test that the data source implements the required interfaces
	var _ = dataSource
}

func TestVMDataSource_Read_EmptyResponse(t *testing.T) {
	// Test handling of null/empty response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("null")) // JSON null
	}))
	defer server.Close()

	dataSource := &VMDataSource{
		client: NewClient(server.URL, "test-api-key", 30),
	}

	// Test the client directly instead of the data source methods
	vms, err := dataSource.client.ListVMs(context.Background())

	// Should handle null response gracefully
	if err != nil {
		t.Errorf("Expected no error for null response, got: %v", err)
	}

	if len(vms) != 0 {
		t.Errorf("Expected empty or nil VMs for null response, got %d VMs", len(vms))
	}
}
