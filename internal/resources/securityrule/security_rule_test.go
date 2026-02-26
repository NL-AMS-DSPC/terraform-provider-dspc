package securityrule

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestResource_Create(t *testing.T) {
	tests := []struct {
		name           string
		sgName         string
		direction      string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:      "successful add ingress rule",
			sgName:    "test-sg",
			direction: "ingress",
			mockResponse: &client.SecurityGroup{
				Name: "test-sg",
				IngressRules: []client.SecurityRule{
					{Index: 0, Ports: []client.SecurityPort{{Protocol: "TCP", Port: 80}}},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error - security group not found",
			sgName:         "nonexistent-sg",
			direction:      "ingress",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-jwt-token",
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

			ruleResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			rules := []client.AddRuleRequest{
				{
					Direction: tt.direction,
					Rule: client.SecurityRule{
						Ports: []client.SecurityPort{{Protocol: "TCP", Port: 80}},
					},
				},
			}

			sg, err := ruleResource.client.AddSecurityRules(context.Background(), tt.sgName, rules)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.sgName, sg.Name)
			}
		})
	}
}

func TestResource_Delete(t *testing.T) {
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
			name:           "API error - not found",
			sgName:         "nonexistent-sg",
			direction:      "ingress",
			index:          0,
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-jwt-token",
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

			ruleResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			err := ruleResource.client.DeleteSecurityRule(context.Background(), tt.sgName, tt.direction, tt.index)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResource_Metadata(t *testing.T) {
	r := NewResource()
	req := resource.MetadataRequest{ProviderTypeName: "dspc"}
	resp := &resource.MetadataResponse{}
	r.Metadata(context.Background(), req, resp)
	assert.Equal(t, "dspc_security_rule", resp.TypeName)
}

func TestResource_Schema(t *testing.T) {
	r := NewResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), req, resp)

	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "security_group_name")
	assert.Contains(t, resp.Schema.Attributes, "direction")
	assert.Contains(t, resp.Schema.Attributes, "index")
	assert.Contains(t, resp.Schema.Blocks, "peers")
	assert.Contains(t, resp.Schema.Blocks, "ports")
}

func TestDataSource_Read(t *testing.T) {
	tests := []struct {
		name           string
		sgName         string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:   "successful list rules",
			sgName: "test-sg",
			mockResponse: &client.ListRulesResponse{
				Ingress: []client.SecurityRule{
					{Index: 0, Ports: []client.SecurityPort{{Protocol: "TCP", Port: 80}}},
				},
				Egress: []client.SecurityRule{
					{Index: 0, Ports: []client.SecurityPort{{Protocol: "TCP", Port: 443}}},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "API error - not found",
			sgName:         "nonexistent-sg",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
	}

	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "mock-jwt-token",
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

			ds := &DataSource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			resp, err := ds.client.ListSecurityRules(context.Background(), tt.sgName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestDataSource_Metadata(t *testing.T) {
	d := NewDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "dspc"}
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "dspc_security_rules", resp.TypeName)
}

func TestDataSource_Schema(t *testing.T) {
	d := NewDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), req, resp)

	assert.Contains(t, resp.Schema.Attributes, "security_group_name")
	assert.Contains(t, resp.Schema.Attributes, "ingress_rules")
	assert.Contains(t, resp.Schema.Attributes, "egress_rules")
}

func TestParseRuleStateID(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		wantSG        string
		wantDirection string
		wantIndex     int
		wantErr       bool
	}{
		{
			name:          "valid ID",
			id:            "my-sg:ingress:0",
			wantSG:        "my-sg",
			wantDirection: "ingress",
			wantIndex:     0,
			wantErr:       false,
		},
		{
			name:          "valid egress ID",
			id:            "test-sg:egress:2",
			wantSG:        "test-sg",
			wantDirection: "egress",
			wantIndex:     2,
			wantErr:       false,
		},
		{
			name:    "invalid format - missing parts",
			id:      "my-sg:ingress",
			wantErr: true,
		},
		{
			name:    "invalid index",
			id:      "my-sg:ingress:abc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg, direction, index, err := parseRuleStateID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSG, sg)
				assert.Equal(t, tt.wantDirection, direction)
				assert.Equal(t, tt.wantIndex, index)
			}
		})
	}
}
