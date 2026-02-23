package vpc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func TestVPCDataSource_Read(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list with multiple VPCs",
			mockResponse: []*client.VPC{
				{Name: "vpc-1", CIDR: "10.0.0.0/24", Status: "active"},
				{Name: "vpc-2", CIDR: "10.1.0.0/24", Status: "active"},
				{Name: "vpc-3", CIDR: "10.2.0.0/24", Status: "pending"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  3,
		},
		{
			name:           "successful list with empty result",
			mockResponse:   []*client.VPC{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name: "successful list with single VPC",
			mockResponse: []*client.VPC{
				{Name: "single-vpc", CIDR: "10.0.0.0/24", Status: "active"},
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
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/v1/namespaces/test-ns/vpcs" {
					t.Fatalf("Expected /v1/namespaces/test-ns/vpcs path, got %s", r.URL.Path)
				}

				authHeader := r.Header.Get("Authorization")
				if authHeader != "Bearer test-api-key" {
					t.Errorf("Expected Authorization: Bearer test-api-key, got %s", authHeader)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			dataSource := &VPCDataSource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-api-key", 30).Network,
			}

			vpcs, err := dataSource.client.ListVPCs(context.Background())

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				if len(vpcs) != tt.expectedCount {
					t.Errorf("Expected %d VPCs, got %d", tt.expectedCount, len(vpcs))
				}
				for i, v := range vpcs {
					if v.Name == "" {
						t.Errorf("VPC %d has empty name", i)
					}
				}
			}
		})
	}
}

func TestVPCDataSource_Metadata(t *testing.T) {
	dataSource := &VPCDataSource{}

	req := datasource.MetadataRequest{
		ProviderTypeName: "dspc",
	}
	resp := &datasource.MetadataResponse{}

	dataSource.Metadata(context.Background(), req, resp)

	expectedTypeName := "dspc_vpcs"
	if resp.TypeName != expectedTypeName {
		t.Errorf("Expected type name '%s', got '%s'", expectedTypeName, resp.TypeName)
	}
}

func TestVPCDataSource_Schema(t *testing.T) {
	dataSource := &VPCDataSource{}

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
	if _, ok := attributes["vpcs"]; !ok {
		t.Error("Data source schema missing 'vpcs' attribute")
	}
}

func TestVPCDataSource_Configure(t *testing.T) {
	tests := []struct {
		name         string
		providerData interface{}
		expectError  bool
	}{
		{
			name:         "valid client",
			providerData: client.NewDspcClient("http://localhost", "test-ns", "test-key", 30),
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
			dataSource := &VPCDataSource{}

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

func TestNewVPCDataSource(t *testing.T) {
	dataSource := NewVPCDataSource()

	if dataSource == nil {
		t.Error("NewVPCDataSource returned nil")
	}
}

func TestVPCDataSource_Read_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("null"))
	}))
	defer server.Close()

	dataSource := &VPCDataSource{
		client: client.NewDspcClient(server.URL, "test-ns", "test-api-key", 30).Network,
	}

	vpcs, err := dataSource.client.ListVPCs(context.Background())

	if err != nil {
		t.Errorf("Expected no error for null response, got: %v", err)
	}

	if len(vpcs) != 0 {
		t.Errorf("Expected empty or nil VPCs for null response, got %d VPCs", len(vpcs))
	}
}

type vpcClientMock struct {
}

func (m *vpcClientMock) ListVPCs(_ context.Context) ([]*client.VPC, error) {
	return []*client.VPC{}, nil
}
