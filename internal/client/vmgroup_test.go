package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateVMGroup(t *testing.T) {
	tests := map[string]struct {
		request        CreateVMGroupRequest
		mockResponses  []mockResponse
		expectedResult VMGroup
		expectError    bool
	}{
		"successful creation": {
			request: CreateVMGroupRequest{
				Name:  "test-vm-group",
				SKUID: "test-sku-id",
				VPCID: "test-vpc-id",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vm-groups/",
					statusCode: http.StatusCreated,
				},
				{
					method:     http.MethodGet,
					path:       "/vm-groups/{name}",
					statusCode: http.StatusOK,
					response: VMGroup{
						URN:    "test-vm-group-urn",
						Name:   "test-vm-group",
						Status: "running",
					},
				},
			},
			expectedResult: VMGroup{
				URN:    "test-vm-group-urn",
				Name:   "test-vm-group",
				Status: "running",
			},
			expectError: false,
		},
		"create error": {
			request: CreateVMGroupRequest{
				Name:  "existing-vm-group",
				SKUID: "test-sku-id",
				VPCID: "test-vpc-id",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vm-groups/",
					statusCode: http.StatusConflict,
					response:   map[string]string{"error": "VM group name already exists"},
				},
			},
			expectError: true,
		},
		"get error": {
			request: CreateVMGroupRequest{
				Name:  "test-vm-group",
				SKUID: "test-sku-id",
				VPCID: "test-vpc-id",
			},
			mockResponses: []mockResponse{
				{
					method:     http.MethodPost,
					path:       "/vm-groups/",
					statusCode: http.StatusCreated,
				},
				{
					method:     http.MethodGet,
					path:       "/vm-groups/{name}",
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

			client := newTestAscClient(server.URL, authServer.URL).VMGroups

			vmGroup, err := client.CreateVMGroup(context.Background(), tt.request)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, vmGroup)
			}
		})
	}
}

func TestGetVMGroup(t *testing.T) {
	tests := map[string]struct {
		vmGroupName    string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		"successful get": {
			vmGroupName: "test-vm-group",
			mockResponse: &VMGroup{
				Name:   "test-vm-group",
				URN:    "test-vm-group-urn",
				Status: "running",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		"not found": {
			vmGroupName:    "nonexistent-vm-group",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		"server error": {
			vmGroupName:    "test-vm-group",
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

			client := newTestAscClient(server.URL, authServer.URL).VMGroups

			vmGroup, err := client.GetVMGroup(context.Background(), tt.vmGroupName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.vmGroupName, vmGroup.Name)
			}
		})
	}
}

func TestListVMGroups(t *testing.T) {
	tests := map[string]struct {
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		"successful list with multiple VM groups": {
			mockResponse: []*VMGroup{
				{Name: "vm-group-1", URN: "vm-group-1-urn", Status: "running"},
				{Name: "vm-group-2", URN: "vm-group-2-urn", Status: "stopped"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		"empty list": {
			mockResponse:   []*VMGroup{},
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

			client := newTestAscClient(server.URL, authServer.URL).VMGroups

			vmGroups, err := client.ListVMGroups(context.Background())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, vmGroups, tt.expectedCount)
			}
		})
	}
}

func TestDeleteVMGroup(t *testing.T) {
	tests := map[string]struct {
		vmGroupName    string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		"successful deletion": {
			vmGroupName:    "test-vm-group",
			mockResponse:   map[string]string{"deleted": "test-vm-group"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		"not found": {
			vmGroupName:    "nonexistent-vm-group",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
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

			client := newTestAscClient(server.URL, authServer.URL).VMGroups

			err := client.DeleteVMGroup(context.Background(), tt.vmGroupName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
