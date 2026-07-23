package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVM(t *testing.T) {
	tests := map[string]struct {
		request        CreateVMRequest
		mockResponses  []mockResponse
		expectedResult VM
		expectError    bool
	}{
		"successful creation": {
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
		"create error": {
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
		"get error": {
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

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockRouteServer("/api/vm", tt.mockResponses)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).VirtualMachines

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
	tests := map[string]struct {
		vmName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		"successful get": {
			vmName: "test-vm",
			mockResponse: &VM{
				Name:   "test-vm",
				URN:    "test-vm-urn",
				Status: "running",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		"not found": {
			vmName:         "nonexistent-vm",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		"server error": {
			vmName:         "test-vm",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).VirtualMachines

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
	tests := map[string]struct {
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		"successful list with multiple VMs": {
			mockResponse: []*VM{
				{Name: "vm-1", URN: "vm-1-urn", Status: "running"},
				{Name: "vm-2", URN: "vm-2-urn", Status: "stopped"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		"empty list": {
			mockResponse:   []*VM{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		"server error": {
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).VirtualMachines

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
