package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNetworkClient_CreateSecurityGroup(t *testing.T) {
	tests := []struct {
		name           string
		sgName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful creation",
			sgName: "test-sg",
			mockResponse: &SecurityGroup{
				Name: "test-sg",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "conflict error",
			sgName:         "existing-sg",
			mockResponse:   map[string]string{"error": "security group name already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name:           "validation error",
			sgName:         "",
			mockResponse:   map[string]string{"error": "validation error"},
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			sg, err := client.CreateSecurityGroup(context.Background(), tt.sgName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.sgName, sg.Name)
			}
		})
	}
}

func TestNetworkClient_GetSecurityGroup(t *testing.T) {
	tests := []struct {
		name           string
		sgName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful get",
			sgName: "test-sg",
			mockResponse: &SecurityGroup{
				Name: "test-sg",
				IngressRules: []SecurityRule{
					{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 80}}},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			sgName:         "nonexistent-sg",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			sgName:         "test-sg",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			sg, err := client.GetSecurityGroup(context.Background(), tt.sgName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.sgName, sg.Name)
			}
		})
	}
}

func TestNetworkClient_ListSecurityGroups(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list with multiple security groups",
			mockResponse: []*SecurityGroup{
				{Name: "sg-1"},
				{Name: "sg-2"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			mockResponse:   []*SecurityGroup{},
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
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			sgs, err := client.ListSecurityGroups(context.Background())
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, sgs, tt.expectedCount)
			}
		})
	}
}

func TestNetworkClient_DeleteSecurityGroup(t *testing.T) {
	tests := []struct {
		name           string
		sgName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			sgName:         "test-sg",
			mockResponse:   map[string]string{"message": "security group deleted"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			sgName:         "nonexistent-sg",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			err := client.DeleteSecurityGroup(context.Background(), tt.sgName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNetworkClient_AddSecurityRules(t *testing.T) {
	tests := []struct {
		name           string
		sgName         string
		rules          []AddRuleRequest
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful add ingress rule",
			sgName: "test-sg",
			rules: []AddRuleRequest{
				{
					Direction: "ingress",
					Rule: SecurityRule{
						Ports: []SecurityPort{{Protocol: "TCP", Port: 80}},
						Peers: []SecurityPeer{{IPBlock: &IPBlock{CIDR: "10.0.0.0/24"}}},
					},
				},
			},
			mockResponse: &SecurityGroup{
				Name: "test-sg",
				IngressRules: []SecurityRule{
					{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 80}}},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "security group not found",
			sgName:         "nonexistent-sg",
			rules:          []AddRuleRequest{{Direction: "ingress", Rule: SecurityRule{Ports: []SecurityPort{{Protocol: "TCP", Port: 80}}}}},
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "validation error",
			sgName:         "test-sg",
			rules:          []AddRuleRequest{{Direction: "invalid", Rule: SecurityRule{}}},
			mockResponse:   map[string]string{"error": "validation error"},
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			sg, err := client.AddSecurityRules(context.Background(), tt.sgName, tt.rules)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.sgName, sg.Name)
			}
		})
	}
}

func TestNetworkClient_ListSecurityRules(t *testing.T) {
	tests := []struct {
		name           string
		sgName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful list",
			sgName: "test-sg",
			mockResponse: &ListRulesResponse{
				Ingress: []SecurityRule{
					{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 80}}},
				},
				Egress: []SecurityRule{
					{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 443}}},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			sgName:         "nonexistent-sg",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			resp, err := client.ListSecurityRules(context.Background(), tt.sgName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestNetworkClient_DeleteSecurityRule(t *testing.T) {
	tests := []struct {
		name           string
		sgName         string
		direction      string
		index          int
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			sgName:         "test-sg",
			direction:      "ingress",
			index:          0,
			mockResponse:   map[string]string{"message": "rule deleted"},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			sgName:         "nonexistent-sg",
			direction:      "ingress",
			index:          0,
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).Network

			err := client.DeleteSecurityRule(context.Background(), tt.sgName, tt.direction, tt.index)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNetworkClient_CreateSecurityGroup_VerifiesRequestBody(t *testing.T) {
	authServer := createMockAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected POST request, got %s", r.Method)
		}
		expectedPath := DefaultServiceConfig().Network.PathPrefix + "/v1/namespaces/test-ns/security-groups"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req CreateSecurityGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		assert.Equal(t, "my-sg", req.Name)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(&SecurityGroup{
			Name: req.Name,
		})
	}))
	defer server.Close()

	client := newTestDspcClient(server.URL, authServer.URL).Network
	sg, err := client.CreateSecurityGroup(context.Background(), "my-sg")

	assert.NoError(t, err)
	assert.Equal(t, "my-sg", sg.Name)
}

func TestNetworkClient_AddSecurityRules_VerifiesRequestBody(t *testing.T) {
	authServer := createMockAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected POST request, got %s", r.Method)
		}
		expectedPath := DefaultServiceConfig().Network.PathPrefix + "/v1/namespaces/test-ns/security-groups/my-sg/rules"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req AddRulesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		assert.Len(t, req.Rules, 1)
		assert.Equal(t, "egress", req.Rules[0].Direction)
		assert.Len(t, req.Rules[0].Rule.Ports, 1)
		assert.Equal(t, "TCP", req.Rules[0].Rule.Ports[0].Protocol)
		assert.Equal(t, 443, req.Rules[0].Rule.Ports[0].Port)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&SecurityGroup{
			Name: "my-sg",
			EgressRules: []SecurityRule{
				{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 443}}},
			},
		})
	}))
	defer server.Close()

	client := newTestDspcClient(server.URL, authServer.URL).Network
	rules := []AddRuleRequest{
		{
			Direction: "egress",
			Rule: SecurityRule{
				Ports: []SecurityPort{{Protocol: "TCP", Port: 443}},
			},
		},
	}
	sg, err := client.AddSecurityRules(context.Background(), "my-sg", rules)

	assert.NoError(t, err)
	assert.Equal(t, "my-sg", sg.Name)
	assert.Len(t, sg.EgressRules, 1)
}
