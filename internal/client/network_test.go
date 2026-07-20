package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVPC(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateVPCRequest
		mockResponses  []mockResponse
		expectedResult VPC
		expectError    bool
	}{
		{
			name: "successful creation",
			request: CreateVPCRequest{
				Name: "test-vpc",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vpcs",
					statusCode: http.StatusCreated,
				},
				{
					method:     http.MethodGet,
					path:       "/vpcs/{name}",
					statusCode: http.StatusOK,
					response: VPC{
						ID:   "test-vpc-id",
						URN:  "test-vpc-urn",
						Name: "test-vpc",
					},
				},
			},
			expectedResult: VPC{
				ID:   "test-vpc-id",
				URN:  "test-vpc-urn",
				Name: "test-vpc",
			},
			expectError: false,
		},
		{
			name: "create error",
			request: CreateVPCRequest{
				Name: "existing-vpc",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vpcs",
					statusCode: http.StatusConflict,
					response:   map[string]string{"error": "VPC name already exists"},
				},
			},
			expectError: true,
		},
		{
			name: "get error",
			request: CreateVPCRequest{
				Name: "test-vpc",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vpcs",
					statusCode: http.StatusCreated,
				},
				{
					method:     http.MethodGet,
					path:       "/vpcs/{name}",
					statusCode: http.StatusBadGateway,
					response:   map[string]string{"error": "internal error"},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockRouteServer("/api/network", tt.mockResponses)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			vpc, err := client.CreateVPC(context.Background(), tt.request)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, vpc)
			}
		})
	}
}

func TestGetVPC(t *testing.T) {
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

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
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

func TestListVPCs(t *testing.T) {
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

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
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

func TestDeleteVPC(t *testing.T) {
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

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
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

func TestCreateSubnet(t *testing.T) {
	tests := []struct {
		name           string
		vpcName        string
		request        CreateSubnetRequest
		mockResponses  []mockResponse
		expectedResult Subnet
		expectError    bool
	}{
		{
			name:    "successful creation",
			vpcName: "test-vpc",
			request: CreateSubnetRequest{
				VPCID: "test-vpc-id",
				Name:  "test-subnet",
				CIDR:  "10.0.0.0/25",
				Type:  "public",
				Tags: []Tag{
					{
						Key:   "k1",
						Value: "v1",
					},
				},
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vpcs/{name}/subnets",
					statusCode: http.StatusCreated,
				},
				{
					method:     http.MethodGet,
					path:       "/vpcs/{name}/subnets",
					statusCode: http.StatusOK,
					response: []Subnet{
						{
							ID:   "test-id",
							URN:  "test-urn",
							Name: "test-subnet",
						},
					},
				},
			},
			expectedResult: Subnet{
				ID:   "test-id",
				URN:  "test-urn",
				Name: "test-subnet",
			},
			expectError: false,
		},
		{
			name:    "create error",
			vpcName: "test-vpc",
			request: CreateSubnetRequest{
				VPCID: "test-vpc-id",
				Name:  "existing-subnet",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vpcs/{name}/subnets",
					statusCode: http.StatusConflict,
					response:   map[string]string{"error": "subnet name already exists in this VPC"},
				},
			},
			expectError: true,
		},
		{
			name:    "list error",
			vpcName: "test-vpc",
			request: CreateSubnetRequest{
				VPCID: "test-vpc-id",
				Name:  "test-subnet",
				CIDR:  "10.0.0.0/25",
				Type:  "public",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vpcs/{name}/subnets",
					statusCode: http.StatusCreated,
				},
				{
					method:     http.MethodGet,
					path:       "/vpcs/{name}/subnets",
					statusCode: http.StatusBadGateway,
					response:   map[string]string{"error": "internal error"},
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockRouteServer("/api/network", tt.mockResponses)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			response, err := client.CreateSubnet(context.Background(), tt.vpcName, tt.request)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, response)
			}
		})
	}
}

func TestListSubnetsForVPC(t *testing.T) {
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

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
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

func TestDeleteSubnet(t *testing.T) {
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

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
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
