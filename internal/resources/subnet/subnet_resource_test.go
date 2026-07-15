package subnet

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestResource_Create(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
		vpcID          string
		subnetName     string
		cidr           string
		subnetType     string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:       "successful creation",
			vpcName:    "test-vpc",
			vpcID:      "test-vpc-id",
			subnetName: "test-subnet",
			cidr:       "10.0.0.0/25",
			subnetType: "public",
			mockResponse: &client.Subnet{
				Name:   "test-subnet",
				CIDR:   "10.0.0.0/25",
				Type:   "public",
				VPCID:  "test-vpc",
				Status: "pending",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "API error - conflict",
			vpcName:        "test-vpc",
			vpcID:          "test-vpc-id",
			subnetName:     "existing-subnet",
			cidr:           "10.0.0.0/25",
			subnetType:     "public",
			mockResponse:   map[string]string{"error": "Subnet name already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name:           "API error - VPC not found",
			vpcName:        "nonexistent-vpc",
			vpcID:          "nonexistent-vpc-id",
			subnetName:     "test-subnet",
			cidr:           "10.0.0.0/25",
			subnetType:     "public",
			mockResponse:   map[string]string{"error": "VPC not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			subnetResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			subnet, err := subnetResource.client.CreateSubnet(
				context.Background(),
				tt.vpcName,
				tt.vpcID,
				tt.subnetName,
				tt.cidr,
				tt.subnetType,
				nil,
			)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.subnetName, subnet.Name)
				assert.Equal(t, tt.cidr, subnet.CIDR)
				assert.Equal(t, tt.subnetType, subnet.Type)
			}
		})
	}
}

func TestResource_Read_FindSubnet(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
		subnetName     string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectFound    bool
	}{
		{
			name:       "found in list",
			vpcName:    "test-vpc",
			subnetName: "test-subnet",
			mockResponse: []*client.Subnet{
				{Name: "other-subnet", CIDR: "10.0.0.128/25", Type: "private", VPCID: "test-vpc", Status: "active"},
				{Name: "test-subnet", CIDR: "10.0.0.0/25", Type: "public", VPCID: "test-vpc", Status: "active"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectFound:    true,
		},
		{
			name:       "not found in list",
			vpcName:    "test-vpc",
			subnetName: "missing-subnet",
			mockResponse: []*client.Subnet{
				{Name: "other-subnet", CIDR: "10.0.0.0/25", Type: "public", VPCID: "test-vpc", Status: "active"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectFound:    false,
		},
		{
			name:           "empty list",
			vpcName:        "test-vpc",
			subnetName:     "test-subnet",
			mockResponse:   []*client.Subnet{},
			mockStatusCode: http.StatusOK,
			expectError:    true,
			expectFound:    false,
		},
		{
			name:           "API error",
			vpcName:        "test-vpc",
			subnetName:     "test-subnet",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectFound:    false,
		},
	}

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			subnetResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			subnet, err := subnetResource.findSubnet(context.Background(), tt.vpcName, tt.subnetName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.expectFound {
				assert.NotNil(t, subnet)
				assert.Equal(t, tt.subnetName, subnet.Name)
			}
		})
	}
}

func TestResource_Delete(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
		subnetName     string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			vpcName:        "test-vpc",
			subnetName:     "test-subnet",
			mockResponse:   map[string]string{"deleted": "test-subnet"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error - not found",
			vpcName:        "test-vpc",
			subnetName:     "nonexistent-subnet",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			subnetResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			err := subnetResource.client.DeleteSubnet(context.Background(), tt.vpcName, tt.subnetName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResource_ImportState_CompositeID(t *testing.T) {
	tests := []struct {
		name        string
		importID    string
		expectParts int
	}{
		{
			name:        "valid composite ID",
			importID:    "my-vpc:my-subnet",
			expectParts: 2,
		},
		{
			name:        "missing colon",
			importID:    "my-vpc-my-subnet",
			expectParts: 1,
		},
		{
			name:        "empty string",
			importID:    "",
			expectParts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := splitImportID(tt.importID)
			assert.Len(t, parts, tt.expectParts)
		})
	}
}

func TestResource_Update(t *testing.T) {
	subnetResource := &Resource{}

	req := resource.UpdateRequest{}
	resp := &resource.UpdateResponse{}

	subnetResource.Update(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("Expected error from Update, got none")
	}
}

func TestCreateSubnetStateID(t *testing.T) {
	id := createSubnetStateID("my-vpc", "my-subnet")
	assert.Equal(t, "my-vpc:my-subnet", id)
}

func TestIsNotFoundError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "not found error",
			err:      fmt.Errorf("subnet %q not found in VPC %q", "test", "vpc"),
			expected: true,
		},
		{
			name:     "API 404 error",
			err:      fmt.Errorf("API error 404: not found"),
			expected: true,
		},
		{
			name:     "other error",
			err:      fmt.Errorf("connection refused"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNotFoundError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
