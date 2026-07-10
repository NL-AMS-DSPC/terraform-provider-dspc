package subnet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
)

func TestDataSource_Read(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name:    "successful list with multiple subnets",
			vpcName: "test-vpc",
			mockResponse: []*client.Subnet{
				{Name: "public-subnet", CIDR: "10.0.0.0/25", Type: "public", Status: "active"},
				{Name: "private-subnet", CIDR: "10.0.0.128/25", Type: "private", Status: "active"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "successful list with empty result",
			vpcName:        "empty-vpc",
			mockResponse:   []*client.Subnet{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:    "successful list with single subnet",
			vpcName: "test-vpc",
			mockResponse: []*client.Subnet{
				{Name: "single-subnet", CIDR: "10.0.0.0/25", Type: "public", Status: "active"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  1,
		},
		{
			name:           "API error",
			vpcName:        "test-vpc",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{ // nolint:gosec
					"access_token": "mock-jwt",
					"expires_in":   3600,
					"token_type":   "Bearer",
				})
			}))
			defer authServer.Close()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/api/network/v1/namespaces/test-ns/vpcs/%s/subnets", tt.vpcName)
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != expectedPath {
					t.Fatalf("Expected %s path, got %s", expectedPath, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			dataSource := &DataSource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			subnets, err := dataSource.client.ListSubnetsForVPC(context.Background(), tt.vpcName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(subnets) != tt.expectedCount {
					t.Errorf("Expected %d subnets, got %d", tt.expectedCount, len(subnets))
				}
			}
		})
	}
}

func TestDataSource_Metadata(t *testing.T) {
	dataSource := &DataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "dspc",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)

	expectedTypeName := "dspc_subnets"
	if resp.TypeName != expectedTypeName {
		t.Errorf("Expected type name '%s', got '%s'", expectedTypeName, resp.TypeName)
	}
}

func TestDataSource_Schema(t *testing.T) {
	dataSource := &DataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	dataSource.Schema(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("Data source schema has errors: %v", resp.Diagnostics)
	}

	if resp.Schema.Attributes == nil {
		t.Error("Data source schema attributes is nil")
	}

	attributes := resp.Schema.Attributes
	if _, ok := attributes["vpc_name"]; !ok {
		t.Error("Data source schema missing 'vpc_name' attribute")
	}
	if _, ok := attributes["subnets"]; !ok {
		t.Error("Data source schema missing 'subnets' attribute")
	}
}

func TestDataSource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData interface{}
		expectError  bool
	}{
		{
			name:         "valid client",
			providerData: client.NewDspcClient("http://localhost", "test-ns", "test-user", "test-pass", "http://auth.example.com", "test-org", 30),
			expectError:  false,
		},
		{
			name:         "nil provider data",
			providerData: nil,
			expectError:  false,
		},
		{
			name:         "invalid provider data type",
			providerData: "not-a-client",
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataSource := &DataSource{}

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

func TestNewDataSource(t *testing.T) {
	dataSource := NewDataSource()

	if dataSource == nil {
		t.Error("NewDataSource returned nil")
	}
}
