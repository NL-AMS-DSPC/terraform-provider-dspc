package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgreSQLClient_CreateInstance(t *testing.T) {
	tests := []struct {
		name           string
		request        CreatePostgreSQLInstanceRequest
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful creation",
			request: CreatePostgreSQLInstanceRequest{
				Name:    "test-db",
				SkuSize: "gp-2",
				Version: DatabaseVersionPostgres17,
				VPCID:   "11111111-1111-1111-1111-111111111111",
			},
			mockResponse: &PostgreSQLInstance{
				Name:    "test-db",
				SkuSize: "gp-2",
				Version: DatabaseVersionPostgres17,
				VPCID:   "11111111-1111-1111-1111-111111111111",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "creation with tags",
			request: CreatePostgreSQLInstanceRequest{
				Name:    "tagged-db",
				SkuSize: "500Mi",
				Version: DatabaseVersionPostgres16,
				VPCID:   "11111111-1111-1111-1111-111111111112",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockResponse: &PostgreSQLInstance{
				Name:    "tagged-db",
				SkuSize: "500Mi",
				Version: DatabaseVersionPostgres16,
				VPCID:   "11111111-1111-1111-1111-111111111112",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "conflict error",
			request: CreatePostgreSQLInstanceRequest{
				Name:    "existing-db",
				SkuSize: "gp-2",
				Version: DatabaseVersionPostgres17,
				VPCID:   "11111111-1111-1111-1111-111111111111",
			},
			mockResponse:   map[string]string{"error": "already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "bad request",
			request: CreatePostgreSQLInstanceRequest{
				Name:    "bad-db",
				SkuSize: "invalid",
				Version: "UNKNOWN",
				VPCID:   "11111111-1111-1111-1111-111111111111",
			},
			mockResponse:   map[string]string{"error": "invalid size format"},
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

			client := newTestAscClient(server.URL, authServer.URL).ManagedDB

			instance, err := client.CreatePostgreSQLInstance(context.Background(), tt.request)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, instance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
				assert.Equal(t, tt.request.Name, instance.Name)
				assert.Equal(t, tt.request.SkuSize, instance.SkuSize)
				assert.Equal(t, tt.request.Version, instance.Version)
				assert.Equal(t, tt.request.VPCID, instance.VPCID)
			}
		})
	}
}

func TestPostgreSQLClient_GetInstance(t *testing.T) {
	tests := []struct {
		name           string
		instanceName   string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:         "successful get",
			instanceName: "test-db",
			mockResponse: &PostgreSQLInstance{
				Name:    "test-db",
				SkuSize: "gp-2",
				Version: DatabaseVersionPostgres17,
				VPCID:   "11111111-1111-1111-1111-111111111111",
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
			instanceName:   "test-db",
			mockResponse:   map[string]string{"error": "internal server error"},
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).ManagedDB

			instance, err := client.GetPostgreSQLInstance(context.Background(), tt.instanceName)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, instance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
				assert.Equal(t, tt.instanceName, instance.Name)
			}
		})
	}
}

func TestPostgreSQLClient_ListInstances(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list",
			mockResponse: &ListPostgreSQLInstancesResponse{
				Data: []PostgreSQLInstance{
					{Name: "db-1", SkuSize: "gp-2", Version: DatabaseVersionPostgres17, VPCID: "vpc-a"},
					{Name: "db-2", SkuSize: "2Gi", Version: DatabaseVersionPostgres16, VPCID: "vpc-b"},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			mockResponse:   &ListPostgreSQLInstancesResponse{Data: []PostgreSQLInstance{}},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  0,
		},
		{
			name:           "server error",
			mockResponse:   map[string]string{"error": "internal server error"},
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

			client := newTestAscClient(server.URL, authServer.URL).ManagedDB

			resp, err := client.ListPostgreSQLInstances(context.Background())
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

func TestPostgreSQLClient_UpdateInstance(t *testing.T) {
	tests := []struct {
		name           string
		instanceName   string
		request        UpdatePostgreSQLInstanceRequest
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:         "successful update",
			instanceName: "test-db",
			request: UpdatePostgreSQLInstanceRequest{
				Name:    "test-db",
				SkuSize: "2Gi",
				Version: DatabaseVersionPostgres18,
				VPCID:   "11111111-1111-1111-1111-111111111111",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockResponse: &PostgreSQLInstance{
				Name:    "test-db",
				SkuSize: "2Gi",
				Version: DatabaseVersionPostgres18,
				VPCID:   "11111111-1111-1111-1111-111111111111",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:         "not found",
			instanceName: "nonexistent-db",
			request: UpdatePostgreSQLInstanceRequest{
				Name:    "nonexistent-db",
				SkuSize: "gp-2",
				Version: DatabaseVersionPostgres17,
				VPCID:   "11111111-1111-1111-1111-111111111111",
			},
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:         "bad request",
			instanceName: "test-db",
			request: UpdatePostgreSQLInstanceRequest{
				Name:    "test-db",
				SkuSize: "invalid",
				Version: DatabaseVersionPostgres17,
				VPCID:   "11111111-1111-1111-1111-111111111111",
			},
			mockResponse:   map[string]string{"error": "invalid storage size"},
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

			client := newTestAscClient(server.URL, authServer.URL).ManagedDB

			instance, err := client.UpdatePostgreSQLInstance(context.Background(), tt.instanceName, tt.request)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, instance)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, instance)
				assert.Equal(t, tt.request.Name, instance.Name)
				assert.Equal(t, tt.request.SkuSize, instance.SkuSize)
				assert.Equal(t, tt.request.Version, instance.Version)
				assert.Equal(t, tt.request.VPCID, instance.VPCID)
			}
		})
	}
}

func TestPostgreSQLClient_DeleteInstance(t *testing.T) {
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
			authServer := createMockAuthServer()
			defer authServer.Close()

			server := newMockServer(tt.mockStatusCode, nil)
			defer server.Close()

			client := newTestAscClient(server.URL, authServer.URL).ManagedDB

			err := client.DeletePostgreSQLInstance(context.Background(), tt.instanceName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPostgreSQLClient_CreateInstance_VerifiesRequestBody(t *testing.T) {
	authServer := createMockAuthServer()
	defer authServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected POST request, got %s", r.Method)
		}
		expectedPath := DefaultServiceConfig().ManagedDB.PathPrefix + "/v1/databases"
		if r.URL.Path != expectedPath {
			t.Fatalf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		var req CreatePostgreSQLInstanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		assert.Equal(t, "my-postgres", req.Name)
		assert.Equal(t, "gp-2", req.SkuSize)
		assert.Equal(t, DatabaseVersionPostgres17, req.Version)
		assert.Equal(t, "my-vpc", req.VPCID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(&PostgreSQLInstance{
			Name:    req.Name,
			SkuSize: req.SkuSize,
			Version: req.Version,
			VPCID:   req.VPCID,
		})
	}))
	defer server.Close()

	client := newTestAscClient(server.URL, authServer.URL).ManagedDB
	instance, err := client.CreatePostgreSQLInstance(context.Background(), CreatePostgreSQLInstanceRequest{
		Name:    "my-postgres",
		SkuSize: "gp-2",
		Version: DatabaseVersionPostgres17,
		VPCID:   "my-vpc",
	})

	assert.NoError(t, err)
	assert.Equal(t, "my-postgres", instance.Name)
}
