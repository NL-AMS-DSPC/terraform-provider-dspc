package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkClient_CreateVPC(t *testing.T) {
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
			mockResponse: &VPC{
				Name:   "test-vpc",
				CIDR:   "10.0.0.0/24",
				Status: "pending",
				Subnets: []Subnet{
					{Name: "test-vpc-public", CIDR: "10.0.0.0/25", Type: "public", VPCRef: "test-vpc"},
					{Name: "test-vpc-private", CIDR: "10.0.0.128/25", Type: "private", VPCRef: "test-vpc"},
				},
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "conflict error",
			vpcName:        "existing-vpc",
			cidr:           "10.0.0.0/24",
			mockResponse:   map[string]string{"error": "VPC name already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name:           "validation error",
			vpcName:        "",
			cidr:           "invalid",
			mockResponse:   map[string]string{"error": "validation error"},
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			vpc, err := client.CreateVPC(context.Background(), tt.vpcName, tt.cidr)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.vpcName, vpc.Name)
				assert.Equal(t, tt.cidr, vpc.CIDR)
				assert.Equal(t, "pending", vpc.Status)
				assert.Len(t, vpc.Subnets, 2)
			}
		})
	}
}

func TestNetworkClient_GetVPC(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:    "successful get",
			vpcName: "test-vpc",
			mockResponse: &VPC{
				Name:   "test-vpc",
				CIDR:   "10.0.0.0/24",
				Status: "active",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			vpcName:        "nonexistent-vpc",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			vpcName:        "test-vpc",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			vpc, err := client.GetVPC(context.Background(), tt.vpcName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.vpcName, vpc.Name)
			}
		})
	}
}

func TestNetworkClient_ListVPCs(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list with multiple VPCs",
			mockResponse: []*VPC{
				{Name: "vpc-1", CIDR: "10.0.0.0/24", Status: "active"},
				{Name: "vpc-2", CIDR: "10.1.0.0/24", Status: "active"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			mockResponse:   []*VPC{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:           "server error",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			vpcs, err := client.ListVPCs(context.Background())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, vpcs, tt.expectedCount)
			}
		})
	}
}

func TestNetworkClient_DeleteVPC(t *testing.T) {
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
			name:           "not found",
			vpcName:        "nonexistent-vpc",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			err := client.DeleteVPC(context.Background(), tt.vpcName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNetworkClient_CreateSubnet(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
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
			subnetName: "test-subnet",
			cidr:       "10.0.0.0/25",
			subnetType: "public",
			mockResponse: &Subnet{
				Name:   "test-subnet",
				CIDR:   "10.0.0.0/25",
				Type:   "public",
				VPCRef: "test-vpc",
				Status: "pending",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "conflict error",
			vpcName:        "test-vpc",
			subnetName:     "existing-subnet",
			cidr:           "10.0.0.0/25",
			subnetType:     "public",
			mockResponse:   map[string]string{"error": "subnet name already exists in this VPC"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name:           "VPC not found",
			vpcName:        "nonexistent-vpc",
			subnetName:     "test-subnet",
			cidr:           "10.0.0.0/25",
			subnetType:     "public",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			subnet, err := client.CreateSubnet(context.Background(), tt.vpcName, tt.subnetName, tt.cidr, tt.subnetType)
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

func TestNetworkClient_ListSubnetsForVPC(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name:    "successful list",
			vpcName: "test-vpc",
			mockResponse: []*Subnet{
				{Name: "test-vpc-public", CIDR: "10.0.0.0/25", Type: "public", VPCRef: "test-vpc"},
				{Name: "test-vpc-private", CIDR: "10.0.0.128/25", Type: "private", VPCRef: "test-vpc"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			vpcName:        "test-vpc",
			mockResponse:   []*Subnet{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:           "server error",
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
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			subnets, err := client.ListSubnetsForVPC(context.Background(), tt.vpcName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, subnets, tt.expectedCount)
			}
		})
	}
}

func TestNetworkClient_DeleteSubnet(t *testing.T) {
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
			name:           "not found",
			vpcName:        "test-vpc",
			subnetName:     "nonexistent-subnet",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			err := client.DeleteSubnet(context.Background(), tt.vpcName, tt.subnetName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNetworkClient_CreateVPC_VerifiesRequestBody(t *testing.T) {
	// Create mock auth server
	authServer := createMockAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected POST request, got %s", r.Method)
		}
		expectedPath := DefaultServiceConfig().Network.PathPrefix + "/v1/namespaces/test-ns/vpcs"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req CreateVPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		assert.Equal(t, "my-vpc", req.Name)
		assert.Equal(t, "10.0.0.0/24", req.CIDR)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(&VPC{
			Name:   req.Name,
			CIDR:   req.CIDR,
			Status: "pending",
		})
	}))
	defer server.Close()

	client := newTestDspcClient(server.URL, authServer.URL).Network
	vpc, err := client.CreateVPC(context.Background(), "my-vpc", "10.0.0.0/24")

	assert.NoError(t, err)
	assert.Equal(t, "my-vpc", vpc.Name)
}

func TestNetworkClient_CreateSubnet_VerifiesRequestBody(t *testing.T) {
	// Create mock auth server
	authServer := createMockAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected POST request, got %s", r.Method)
		}
		expectedPath := DefaultServiceConfig().Network.PathPrefix + "/v1/namespaces/test-ns/vpcs/my-vpc/subnets"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req CreateSubnetRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		assert.Equal(t, "my-subnet", req.Name)
		assert.Equal(t, "10.0.0.0/25", req.CIDR)
		assert.Equal(t, "public", req.Type)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(&Subnet{
			Name:   req.Name,
			CIDR:   req.CIDR,
			Type:   req.Type,
			VPCRef: "my-vpc",
			Status: "pending",
		})
	}))
	defer server.Close()

	client := newTestDspcClient(server.URL, authServer.URL).Network
	subnet, err := client.CreateSubnet(context.Background(), "my-vpc", "my-subnet", "10.0.0.0/25", "public")

	assert.NoError(t, err)
	assert.Equal(t, "my-subnet", subnet.Name)
	assert.Equal(t, "my-vpc", subnet.VPCRef)
}
