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
				Name:      "test-sg",
				Namespace: "test-ns",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "conflict error",
			sgName:         "existing-sg",
			mockResponse:   map[string]string{"error": "already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Network

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
				Name:      "test-sg",
				Namespace: "test-ns",
				IngressRules: []SecurityRule{
					{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 80}}},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			sgName:         "nonexistent",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Network

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
			name: "successful list",
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

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Network

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
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			sgName:         "test-sg",
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			sgName:         "nonexistent",
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, map[string]string{"message": "deleted"})
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Network

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
		request        AddRulesRequest
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful add egress rule",
			sgName: "test-sg",
			request: AddRulesRequest{
				Rules: []AddRuleEntry{
					{
						Direction: "egress",
						Ports:     []AddRulePortEntry{{Protocol: "TCP", Port: 5432}},
						Peers:     []AddRulePeerEntry{{IPBlock: &AddRuleIPBlockEntry{CIDR: "10.0.0.0/24"}}},
					},
				},
			},
			mockResponse: &SecurityGroup{
				Name: "test-sg",
				EgressRules: []SecurityRule{
					{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 5432}}},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:   "API error",
			sgName: "test-sg",
			request: AddRulesRequest{
				Rules: []AddRuleEntry{
					{Direction: "ingress", Ports: []AddRulePortEntry{{Protocol: "TCP", Port: 80}}},
				},
			},
			mockResponse:   map[string]string{"error": "bad request"},
			mockStatusCode: http.StatusBadRequest,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).Network

			sg, err := client.AddSecurityRules(context.Background(), tt.sgName, tt.request)
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
	authServer := createMockAuthServer()
	defer authServer.Close()

	mockResp := &ListRulesResponse{
		Ingress: []SecurityRule{
			{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 80}}},
		},
		Egress: []SecurityRule{
			{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 5432}}},
		},
	}

	server := newMockServer(http.StatusOK, mockResp)
	defer server.Close()

	client := newTestAscClient(server.URL, authServer.URL).Network

	resp, err := client.ListSecurityRules(context.Background(), "test-sg")
	assert.NoError(t, err)
	assert.Len(t, resp.Ingress, 1)
	assert.Len(t, resp.Egress, 1)
	assert.Equal(t, "TCP", resp.Ingress[0].Ports[0].Protocol)
}

func TestNetworkClient_AddSecurityRules_VerifiesRequestBody(t *testing.T) {
	authServer := createMockAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected POST request, got %s", r.Method)
		}

		var req AddRulesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		assert.Len(t, req.Rules, 1)
		assert.Equal(t, "egress", req.Rules[0].Direction)
		assert.Equal(t, "TCP", req.Rules[0].Ports[0].Protocol)
		assert.Equal(t, 5432, req.Rules[0].Ports[0].Port)
		assert.Equal(t, "10.0.0.0/24", req.Rules[0].Peers[0].IPBlock.CIDR)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(&SecurityGroup{
			Name: "test-sg",
			EgressRules: []SecurityRule{
				{Index: 0, Ports: []SecurityPort{{Protocol: "TCP", Port: 5432}}},
			},
		})
	}))
	defer server.Close()

	client := newTestAscClient(server.URL, authServer.URL).Network

	addReq := AddRulesRequest{
		Rules: []AddRuleEntry{
			{
				Direction: "egress",
				Ports:     []AddRulePortEntry{{Protocol: "TCP", Port: 5432}},
				Peers:     []AddRulePeerEntry{{IPBlock: &AddRuleIPBlockEntry{CIDR: "10.0.0.0/24"}}},
			},
		},
	}

	sg, err := client.AddSecurityRules(context.Background(), "test-sg", addReq)
	assert.NoError(t, err)
	assert.Equal(t, "test-sg", sg.Name)
}
