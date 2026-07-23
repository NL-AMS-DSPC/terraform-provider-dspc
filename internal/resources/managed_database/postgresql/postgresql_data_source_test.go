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
				SkuSize: "gp-2",
				Version: client.DatabaseVersionPostgres17,
				VPCID:   "test-vpc",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:         "successful read with tags",
			instanceName: "tagged-db",
			mockResponse: &client.PostgreSQLInstance{
				Name:    "tagged-db",
				SkuSize: "c-8",
				Version: client.DatabaseVersionPostgres18,
				VPCID:   "staging-vpc",
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
				client: client.NewAscClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).ManagedDB,
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
				assert.Equal(t, expected.SkuSize, instance.SkuSize)
				assert.Equal(t, expected.Version, instance.Version)
				assert.Equal(t, expected.VPCID, instance.VPCID)
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
					{Name: "db-1", SkuSize: "gp-2", Version: client.DatabaseVersionPostgres17, VPCID: "11111111-1111-1111-1111-111111111111"},
					{Name: "db-2", SkuSize: "gp-4", Version: client.DatabaseVersionPostgres16, VPCID: "11111111-1111-1111-1111-111111111112"},
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
				client: client.NewAscClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).ManagedDB,
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
