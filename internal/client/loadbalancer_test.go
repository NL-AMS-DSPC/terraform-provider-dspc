package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateLoadBalancer(t *testing.T) {
	tests := map[string]struct {
		request        CreateLoadBalancerRequest
		mockResponse   any
		mockStatusCode int
		expectedResult *CreateLoadBalancerResponse
		expectError    bool
	}{
		"successful creation": {
			request: CreateLoadBalancerRequest{
				Name:     "test-lb",
				Scheme:   "internet-facing",
				SKUID:    "test-sku-id",
				VPCID:    "test-vpc-id",
				SubnetID: "test-subnet-id",
				Listeners: []ListenerRequest{
					{Protocol: "TCP", Port: 80},
				},
				Backends: []BackendRequest{
					{IP: "10.0.0.1", Port: 8080},
				},
				Algorithm: "round-robin",
			},
			mockResponse: CreateLoadBalancerResponse{
				ID:   "test-lb-id",
				Name: "test-lb",
			},
			mockStatusCode: http.StatusCreated,
			expectedResult: &CreateLoadBalancerResponse{
				ID:   "test-lb-id",
				Name: "test-lb",
			},
			expectError: false,
		},
		"create error": {
			request: CreateLoadBalancerRequest{
				Name:     "existing-lb",
				Scheme:   "internal",
				SKUID:    "test-sku-id",
				VPCID:    "test-vpc-id",
				SubnetID: "test-subnet-id",
			},
			mockResponse:   map[string]string{"error": "load balancer name already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		"server error": {
			request: CreateLoadBalancerRequest{
				Name:     "test-lb",
				Scheme:   "internal",
				SKUID:    "test-sku-id",
				VPCID:    "test-vpc-id",
				SubnetID: "test-subnet-id",
			},
			mockResponse:   map[string]string{"error": "internal error"},
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

			client := newTestAscClient(server.URL, authServer.URL).LoadBalancers

			lb, err := client.CreateLoadBalancer(context.Background(), tt.request)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, lb)
			}
		})
	}
}

func TestDeleteLoadBalancer(t *testing.T) {
	tests := map[string]struct {
		loadBalancerName string
		mockResponse     any
		mockStatusCode   int
		expectError      bool
	}{
		"successful deletion": {
			loadBalancerName: "test-lb",
			mockResponse:     map[string]string{"deleted": "test-lb"},
			mockStatusCode:   http.StatusOK,
			expectError:      false,
		},
		"not found": {
			loadBalancerName: "nonexistent-lb",
			mockResponse:     map[string]string{"error": "not found"},
			mockStatusCode:   http.StatusNotFound,
			expectError:      true,
		},
		"server error": {
			loadBalancerName: "test-lb",
			mockResponse:     map[string]string{"error": "internal error"},
			mockStatusCode:   http.StatusInternalServerError,
			expectError:      true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			// Create mock auth server
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).LoadBalancers

			err := client.DeleteLoadBalancer(context.Background(), tt.loadBalancerName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
