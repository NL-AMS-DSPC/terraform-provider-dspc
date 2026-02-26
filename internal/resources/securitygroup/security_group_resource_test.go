package securitygroup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
)

func TestResource_Create(t *testing.T) {
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
			mockResponse: &client.SecurityGroup{
				Name: "test-sg",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "API error - conflict",
			sgName:         "existing-sg",
			mockResponse:   map[string]string{"error": "security group name already exists"},
			mockStatusCode: http.StatusConflict,
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

			sgResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			sg, err := sgResource.client.CreateSecurityGroup(context.Background(), tt.sgName)

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

			sgResource := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			err := sgResource.client.DeleteSecurityGroup(context.Background(), tt.sgName)

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
	assert.Equal(t, "dspc_security_group", resp.TypeName)
}

func TestResource_Schema(t *testing.T) {
	r := NewResource()
	req := resource.SchemaRequest{}
	resp := &resource.SchemaResponse{}
	r.Schema(context.Background(), req, resp)

	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
}
