package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVM(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateVMRequest
		mockResponses  []mockResponse
		expectedResult VM
		expectError    bool
	}{
		{
			name: "successful creation",
			request: CreateVMRequest{
				Name:  "test-vm",
				SKUID: "test-sku-id",
				VPCID: "test-vpc-id",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vms/",
					statusCode: http.StatusCreated,
				},
				{
					method:     http.MethodGet,
					path:       "/vms/{name}",
					statusCode: http.StatusOK,
					response: VM{
						URN:    "test-vm-urn",
						Name:   "test-vm",
						Status: "running",
					},
				},
			},
			expectedResult: VM{
				URN:    "test-vm-urn",
				Name:   "test-vm",
				Status: "running",
			},
			expectError: false,
		},
		{
			name: "create error",
			request: CreateVMRequest{
				Name:  "existing-vm",
				SKUID: "test-sku-id",
				VPCID: "test-vpc-id",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vms/",
					statusCode: http.StatusConflict,
					response:   map[string]string{"error": "VM name already exists"},
				},
			},
			expectError: true,
		},
		{
			name: "get error",
			request: CreateVMRequest{
				Name:  "test-vm",
				SKUID: "test-sku-id",
				VPCID: "test-vpc-id",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vms/",
					statusCode: http.StatusCreated,
				},
				{
					method:     http.MethodGet,
					path:       "/vms/{name}",
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

			server := newMockRouteServer("/api/vm", tt.mockResponses)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).VirtualMachines

			vm, err := client.CreateVM(context.Background(), tt.request)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, vm)
			}
		})
	}
}

func TestGetVM(t *testing.T) {
	tests := []struct {
		name           string
		vmName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful get",
			vmName: "test-vm",
			mockResponse: &VM{
				Name:   "test-vm",
				URN:    "test-vm-urn",
				Status: "running",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			vmName:         "nonexistent-vm",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			vmName:         "test-vm",
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

			client := newTestDspcClient(server.URL, authServer.URL).VirtualMachines

			vm, err := client.GetVM(context.Background(), tt.vmName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.vmName, vm.Name)
			}
		})
	}
}

func TestListVMs(t *testing.T) {
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
				{Name: "vm-1", URN: "vm-1-urn", Status: "running"},
				{Name: "vm-2", URN: "vm-2-urn", Status: "stopped"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			mockResponse:   []*VM{},
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

			client := newTestDspcClient(server.URL, authServer.URL).VirtualMachines

			vms, err := client.ListVMs(context.Background())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, vms, tt.expectedCount)
			}
		})
	}
}
