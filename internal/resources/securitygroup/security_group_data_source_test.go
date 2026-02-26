package securitygroup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/stretchr/testify/assert"
)

func TestDataSource_Read(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list with multiple security groups",
			mockResponse: []*client.SecurityGroup{
				{Name: "sg-1"},
				{Name: "sg-2"},
				{Name: "sg-3"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  3,
		},
		{
			name:           "successful list with empty result",
			mockResponse:   []*client.SecurityGroup{},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name: "successful list with single security group",
			mockResponse: []*client.SecurityGroup{
				{Name: "single-sg"},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  1,
		},
		{
			name:           "API error",
			mockResponse:   map[string]string{"error": "Internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Fatalf("Expected GET request, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			ds := &DataSource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			sgs, err := ds.client.ListSecurityGroups(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, sgs, tt.expectedCount)
			}
		})
	}
}

func TestDataSource_Metadata(t *testing.T) {
	d := NewDataSource()
	req := datasource.MetadataRequest{ProviderTypeName: "dspc"}
	resp := &datasource.MetadataResponse{}
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "dspc_security_groups", resp.TypeName)
}

func TestDataSource_Schema(t *testing.T) {
	d := NewDataSource()
	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}
	d.Schema(context.Background(), req, resp)

	assert.Contains(t, resp.Schema.Attributes, "security_groups")
}
