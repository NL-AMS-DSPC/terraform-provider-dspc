package postgresql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestDataSource_Read(t *testing.T) {
	tests := []struct {
		name           string
		instanceName   string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:         "successful read",
			instanceName: "test-db",
			mockResponse: &client.PostgreSQLInstance{
				Name:    "test-db",
				Size:    "1Gi",
				Version: client.DatabaseVersionPostgres17,
				VPC:     "test-vpc",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:         "successful read with tags",
			instanceName: "tagged-db",
			mockResponse: &client.PostgreSQLInstance{
				Name:    "tagged-db",
				Size:    "4Gi",
				Version: client.DatabaseVersionPostgres18,
				VPC:     "staging-vpc",
				Tags:    []client.Tag{{Key: "owner", Value: "team-a"}},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			instanceName:   "missing-db",
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			instanceName:   "any-db",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := newMockAuthServer(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			ds := &DataSource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			instance, err := ds.client.GetPostgreSQLInstance(context.Background(), tt.instanceName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, instance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
				expected := tt.mockResponse.(*client.PostgreSQLInstance) //nolint:forcetypeassert
				assert.Equal(t, expected.Name, instance.Name)
				assert.Equal(t, expected.Size, instance.Size)
				assert.Equal(t, expected.Version, instance.Version)
				assert.Equal(t, expected.VPC, instance.VPC)
				assert.Equal(t, expected.Tags, instance.Tags)
			}
		})
	}
}

func TestDataSource_List(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list",
			mockResponse: &client.ListPostgreSQLInstancesResponse{
				Data: []client.PostgreSQLInstance{
					{Name: "db-1", Size: "1Gi", Version: client.DatabaseVersionPostgres17, VPC: "vpc-a"},
					{Name: "db-2", Size: "2Gi", Version: client.DatabaseVersionPostgres16, VPC: "vpc-b"},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			mockResponse:   &client.ListPostgreSQLInstancesResponse{Data: []client.PostgreSQLInstance{}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:           "API error",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := newMockAuthServer(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatusCode)
				_ = json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			ds := &DataSource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			resp, err := ds.client.ListPostgreSQLInstances(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.Data, tt.expectedCount)
			}
		})
	}
}
