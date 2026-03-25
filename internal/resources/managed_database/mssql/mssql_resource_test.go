package mssql

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
)

// newMockAuthServer returns a test auth server that always issues a mock JWT token.
func newMockAuthServer(t *testing.T) *httptest.Server {
	t.Helper()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{ // nolint:gosec
			"access_token": "mock-jwt",
			"expires_in":   3600,
			"token_type":   "Bearer",
		})
	}))
	t.Cleanup(s.Close)
	return s
}

func TestResource_Create(t *testing.T) {
	tests := []struct {
		name           string
		request        client.CreateMSSQLInstanceRequest
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful creation",
			request: client.CreateMSSQLInstanceRequest{
				Name:    "test-db",
				Size:    "1Gi",
				Version: client.DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockResponse: &client.MSSQLInstance{
				Name:    "test-db",
				Size:    "1Gi",
				Version: client.DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "successful creation with tags",
			request: client.CreateMSSQLInstanceRequest{
				Name:    "tagged-db",
				Size:    "500Mi",
				Version: client.DatabaseVersionMSSQL2019_15,
				VPC:     "prod-vpc",
				Tags:    []client.Tag{{Key: "env", Value: "prod"}, {Key: "team", Value: "platform"}},
			},
			mockResponse: &client.MSSQLInstance{
				Name:    "tagged-db",
				Size:    "500Mi",
				Version: client.DatabaseVersionMSSQL2019_15,
				VPC:     "prod-vpc",
				Tags:    []client.Tag{{Key: "env", Value: "prod"}, {Key: "team", Value: "platform"}},
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "API error - conflict",
			request: client.CreateMSSQLInstanceRequest{
				Name:    "existing-db",
				Size:    "1Gi",
				Version: client.DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockResponse:   map[string]string{"error": "already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "API error - bad request",
			request: client.CreateMSSQLInstanceRequest{
				Name:    "bad-db",
				Size:    "invalid",
				Version: "UNKNOWN_VERSION",
				VPC:     "test-vpc",
			},
			mockResponse:   map[string]string{"error": "invalid size format"},
			mockStatusCode: http.StatusBadRequest,
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

			r := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			instance, err := r.client.CreateMSSQLInstance(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, instance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
				assert.Equal(t, tt.request.Name, instance.Name)
				assert.Equal(t, tt.request.Size, instance.Size)
				assert.Equal(t, tt.request.Version, instance.Version)
				assert.Equal(t, tt.request.VPC, instance.VPC)
			}
		})
	}
}

func TestResource_Read(t *testing.T) {
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
			mockResponse: &client.MSSQLInstance{
				Name:    "test-db",
				Size:    "1Gi",
				Version: client.DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:         "successful read with tags",
			instanceName: "tagged-db",
			mockResponse: &client.MSSQLInstance{
				Name:    "tagged-db",
				Size:    "2Gi",
				Version: client.DatabaseVersionMSSQL2025_17,
				VPC:     "prod-vpc",
				Tags:    []client.Tag{{Key: "env", Value: "staging"}},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "not found",
			instanceName:   "nonexistent-db",
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

			r := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			instance, err := r.client.GetMSSQLInstance(context.Background(), tt.instanceName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, instance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
				expected := tt.mockResponse.(*client.MSSQLInstance) //nolint:forcetypeassert
				assert.Equal(t, expected.Name, instance.Name)
				assert.Equal(t, expected.Size, instance.Size)
				assert.Equal(t, expected.Version, instance.Version)
				assert.Equal(t, expected.VPC, instance.VPC)
			}
		})
	}
}

func TestResource_List(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list with multiple instances",
			mockResponse: &client.ListMSSQLInstancesResponse{
				Data: []client.MSSQLInstance{
					{Name: "db-1", Size: "1Gi", Version: client.DatabaseVersionMSSQL2022_16, VPC: "vpc-a"},
					{Name: "db-2", Size: "2Gi", Version: client.DatabaseVersionMSSQL2019_15, VPC: "vpc-b"},
					{Name: "db-3", Size: "500Mi", Version: client.DatabaseVersionMSSQL2017_14, VPC: "vpc-c"},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  3,
		},
		{
			name:           "successful list - empty",
			mockResponse:   &client.ListMSSQLInstancesResponse{Data: []client.MSSQLInstance{}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:           "API error - server error",
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

			r := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			resp, err := r.client.ListMSSQLInstances(context.Background())

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
			mockResponse: &client.MSSQLInstance{
				Name:    "test-db",
				Size:    "1Gi",
				Version: client.DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:         "successful read with tags",
			instanceName: "tagged-db",
			mockResponse: &client.MSSQLInstance{
				Name:    "tagged-db",
				Size:    "4Gi",
				Version: client.DatabaseVersionMSSQL2025_17,
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

			instance, err := ds.client.GetMSSQLInstance(context.Background(), tt.instanceName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, instance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
				expected := tt.mockResponse.(*client.MSSQLInstance) //nolint:forcetypeassert
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
			mockResponse: &client.ListMSSQLInstancesResponse{
				Data: []client.MSSQLInstance{
					{Name: "db-1", Size: "1Gi", Version: client.DatabaseVersionMSSQL2022_16, VPC: "vpc-a"},
					{Name: "db-2", Size: "2Gi", Version: client.DatabaseVersionMSSQL2019_15, VPC: "vpc-b"},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			mockResponse:   &client.ListMSSQLInstancesResponse{Data: []client.MSSQLInstance{}},
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

			resp, err := ds.client.ListMSSQLInstances(context.Background())

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

func TestResource_Update(t *testing.T) {
	tests := []struct {
		name           string
		instanceName   string
		request        client.UpdateMSSQLInstanceRequest
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:         "successful update - change size",
			instanceName: "test-db",
			request: client.UpdateMSSQLInstanceRequest{
				Name:    "test-db",
				Size:    "2Gi",
				Version: client.DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockResponse: &client.MSSQLInstance{
				Name:    "test-db",
				Size:    "2Gi",
				Version: client.DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:         "successful update - change version and tags",
			instanceName: "test-db",
			request: client.UpdateMSSQLInstanceRequest{
				Name:    "test-db",
				Size:    "1Gi",
				Version: client.DatabaseVersionMSSQL2025_17,
				VPC:     "test-vpc",
				Tags:    []client.Tag{{Key: "env", Value: "prod"}},
			},
			mockResponse: &client.MSSQLInstance{
				Name:    "test-db",
				Size:    "1Gi",
				Version: client.DatabaseVersionMSSQL2025_17,
				VPC:     "test-vpc",
				Tags:    []client.Tag{{Key: "env", Value: "prod"}},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:         "API error - not found",
			instanceName: "nonexistent-db",
			request: client.UpdateMSSQLInstanceRequest{
				Name:    "nonexistent-db",
				Size:    "1Gi",
				Version: client.DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:         "API error - bad request",
			instanceName: "test-db",
			request: client.UpdateMSSQLInstanceRequest{
				Name:    "test-db",
				Size:    "invalid-size",
				Version: client.DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockResponse:   map[string]string{"error": "invalid storage size"},
			mockStatusCode: http.StatusBadRequest,
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

			r := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			instance, err := r.client.UpdateMSSQLInstance(context.Background(), tt.instanceName, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, instance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
				expected := tt.mockResponse.(*client.MSSQLInstance) //nolint:forcetypeassert
				assert.Equal(t, expected.Name, instance.Name)
				assert.Equal(t, expected.Size, instance.Size)
				assert.Equal(t, expected.Version, instance.Version)
				assert.Equal(t, expected.VPC, instance.VPC)
				assert.Equal(t, expected.Tags, instance.Tags)
			}
		})
	}
}

func TestResource_Delete(t *testing.T) {
	tests := []struct {
		name           string
		instanceName   string
		mockStatusCode int
		expectError    bool
	}{
		{
			name:           "successful deletion",
			instanceName:   "test-db",
			mockStatusCode: http.StatusNoContent,
			expectError:    false,
		},
		{
			name:           "not found",
			instanceName:   "nonexistent-db",
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:           "server error",
			instanceName:   "test-db",
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := newMockAuthServer(t)

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.mockStatusCode)
			}))
			defer server.Close()

			r := &Resource{
				client: client.NewDspcClient(server.URL, "test-ns", "test-user", "test-pass", authServer.URL, "test-org", 30).Network,
			}

			err := r.client.DeleteMSSQLInstance(context.Background(), tt.instanceName)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
