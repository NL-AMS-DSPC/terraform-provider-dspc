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
		vpcID          string
		vpcName        string
		tags           []Tag
		subnets        []CreateSubnetRequest
		mockResponse   interface{}
		mockStatusCode int
		expectedResult CreateVPCResponse
		expectError    bool
	}{
		{
			name:    "successful creation",
			vpcID:   "",
			vpcName: "test-vpc",
			mockResponse: &CreateVPCResponse{
				ID:   "test-vpc-id",
				URN:  "test-vpc-id",
				Name: "test-vpc",
			},
			mockStatusCode: http.StatusCreated,
			expectedResult: CreateVPCResponse{
				ID:   "test-vpc-id",
				URN:  "test-vpc-id",
				Name: "test-vpc",
			},
			expectError: false,
		},
		{
			name:           "conflict error",
			vpcID:          "",
			vpcName:        "existing-vpc",
			mockResponse:   map[string]string{"error": "VPC name already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name:           "validation error",
			vpcID:          "",
			vpcName:        "",
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

			vpc, err := client.CreateVPC(context.Background(), tt.vpcID, tt.vpcName, tt.tags, tt.subnets)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, vpc)
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
		vpcID          string
		subnetName     string
		cidr           string
		subnetType     string
		tags           []Tag
		mockResponse   interface{}
		mockStatusCode int
		expectedResult CreateSubnetResponse
		expectError    bool
	}{
		{
			name:       "successful creation",
			vpcName:    "test-vpc",
			vpcID:      "test-vpc-id",
			subnetName: "test-subnet",
			cidr:       "10.0.0.0/25",
			subnetType: "public",
			tags: []Tag{
				{
					Key:   "k1",
					Value: "v1",
				},
			},
			mockResponse: &CreateSubnetResponse{
				ID:      "test-id",
				URN:     "test-urn",
				Created: "created",
			},
			mockStatusCode: http.StatusCreated,
			expectedResult: CreateSubnetResponse{
				ID:      "test-id",
				URN:     "test-urn",
				Created: "created",
			},
			expectError: false,
		},
		{
			name:           "conflict error",
			vpcName:        "test-vpc",
			vpcID:          "test-vpc-id",
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
			vpcID:          "nonexistent-vpc-id",
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

			response, err := client.CreateSubnet(context.Background(), tt.vpcName, tt.vpcID, tt.subnetName, tt.cidr, tt.subnetType, tt.tags)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, response)
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
				{Name: "test-vpc-public", CIDR: "10.0.0.0/25", Type: "public", VPCID: "test-vpc"},
				{Name: "test-vpc-private", CIDR: "10.0.0.128/25", Type: "private", VPCID: "test-vpc"},
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
		assert.Equal(t, []Tag{{
			Key:   "k1",
			Value: "v1",
		}}, req.Tags)
		assert.Equal(t, []CreateSubnetRequest{{
			Name:  "s1",
			CIDR:  "cidr",
			VPCID: "vpc1",
			Type:  "private",
			Tags: []Tag{{
				Key:   "sk1",
				Value: "sv1",
			}},
		}}, req.Subnets)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(&VPC{
			Name:   req.Name,
			Status: "pending",
		})
	}))
	defer server.Close()

	tags := []Tag{{
		Key:   "k1",
		Value: "v1",
	}}
	subnets := []CreateSubnetRequest{{
		Name:  "s1",
		CIDR:  "cidr",
		VPCID: "vpc1",
		Type:  "private",
		Tags: []Tag{{
			Key:   "sk1",
			Value: "sv1",
		}},
	}}
	client := newTestDspcClient(server.URL, authServer.URL).Network
	vpc, err := client.CreateVPC(context.Background(), "my-vpc-id", "my-vpc", tags, subnets)

	assert.NoError(t, err)
	assert.Equal(t, "my-vpc", vpc.Name)
}
