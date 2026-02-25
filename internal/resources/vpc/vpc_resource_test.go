package vpc

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
	vpcPath = "/v1/namespaces/test-ns/vpcs"
)

func TestVPCResource_Create(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
		cidr           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:    "successful creation",
			vpcName: "test-vpc",
			cidr:    "10.0.0.0/24",
			mockResponse: &client.VPC{
				Name:   "test-vpc",
				CIDR:   "10.0.0.0/24",
				Status: "pending",
				Subnets: []client.Subnet{
					{Name: "test-vpc-public", CIDR: "10.0.0.0/25", Type: "public", VPCRef: "test-vpc"},
					{Name: "test-vpc-private", CIDR: "10.0.0.128/25", Type: "private", VPCRef: "test-vpc"},
				},
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "API error - conflict",
			vpcName:        "existing-vpc",
			cidr:           "10.0.0.0/24",
			mockResponse:   map[string]string{"error": "VPC name already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			vpcResource := &VPCResource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-api-key", 30).Network,
			}

			vpc, err := vpcResource.client.CreateVPC(context.Background(), tt.vpcName, tt.cidr)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.vpcName, vpc.Name)
				assert.Equal(t, tt.cidr, vpc.CIDR)
			}
		})
	}
}

func TestVPCResource_Delete(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			vpcName:        "test-vpc",
			mockResponse:   map[string]string{"deleted": "test-vpc"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error - not found",
			vpcName:        "nonexistent-vpc",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			vpcResource := &VPCResource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-api-key", 30).Network,
			}

			err := vpcResource.client.DeleteVPC(context.Background(), tt.vpcName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVPCResource_ImportState(t *testing.T) {
	tests := []struct {
		name           string
		importID       string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:     "successful import",
			importID: "test-vpc",
			mockResponse: &client.VPC{
				Name:   "test-vpc",
				CIDR:   "10.0.0.0/24",
				Status: "active",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "import non-existent VPC",
			importID:       "nonexistent-vpc",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "API error during import",
			importID:       "test-vpc",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != vpcPath+"/"+tt.importID {
					t.Fatalf("Expected %s path, got %s", vpcPath+"/"+tt.importID, r.URL.Path)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			vpcResource := &VPCResource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-api-key", 30).Network,
			}

			vpc, err := vpcResource.client.GetVPC(context.Background(), tt.importID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.importID, vpc.Name)
			}
		})
	}
}

func TestVPCResource_Update(t *testing.T) {
	vpcResource := &VPCResource{}

	req := resource.UpdateRequest{}
	resp := &resource.UpdateResponse{}

	vpcResource.Update(context.Background(), req, resp)

	// Update should always return an error
	if !resp.Diagnostics.HasError() {
		t.Error("Expected error from Update, got none")
	}
}
