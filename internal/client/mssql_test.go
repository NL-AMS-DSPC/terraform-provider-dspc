package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMSSQLClient_CreateInstance(t *testing.T) {
	tests := []struct {
		name           string
		request        CreateMSSQLInstanceRequest
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name: "successful creation",
			request: CreateMSSQLInstanceRequest{
				Name:    "test-db",
				Size:    "1Gi",
				Version: DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockResponse: &MSSQLInstance{
				Name:    "test-db",
				Size:    "1Gi",
				Version: DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "creation with tags",
			request: CreateMSSQLInstanceRequest{
				Name:    "tagged-db",
				Size:    "500Mi",
				Version: DatabaseVersionMSSQL2019_15,
				VPC:     "prod-vpc",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockResponse: &MSSQLInstance{
				Name:    "tagged-db",
				Size:    "500Mi",
				Version: DatabaseVersionMSSQL2019_15,
				VPC:     "prod-vpc",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "conflict error",
			request: CreateMSSQLInstanceRequest{
				Name:    "existing-db",
				Size:    "1Gi",
				Version: DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockResponse:   map[string]string{"error": "already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "bad request",
			request: CreateMSSQLInstanceRequest{
				Name:    "bad-db",
				Size:    "invalid",
				Version: "UNKNOWN",
				VPC:     "test-vpc",
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

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).ManagedDB

			instance, err := client.CreateMSSQLInstance(context.Background(), tt.request)
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

func TestMSSQLClient_GetInstance(t *testing.T) {
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
			mockResponse: &MSSQLInstance{
				Name:    "test-db",
				Size:    "1Gi",
				Version: DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
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

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).ManagedDB

			instance, err := client.GetMSSQLInstance(context.Background(), tt.instanceName)
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

func TestMSSQLClient_ListInstances(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
		expectedCount  int
	}{
		{
			name: "successful list",
			mockResponse: &ListMSSQLInstancesResponse{
				Data: []MSSQLInstance{
					{Name: "db-1", Size: "1Gi", Version: DatabaseVersionMSSQL2022_16, VPC: "vpc-a"},
					{Name: "db-2", Size: "2Gi", Version: DatabaseVersionMSSQL2019_15, VPC: "vpc-b"},
				},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
			expectedCount:  2,
		},
		{
			name:           "empty list",
			mockResponse:   &ListMSSQLInstancesResponse{Data: []MSSQLInstance{}},
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

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).ManagedDB

			resp, err := client.ListMSSQLInstances(context.Background())
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

func TestMSSQLClient_UpdateInstance(t *testing.T) {
	tests := []struct {
		name           string
		instanceName   string
		request        UpdateMSSQLInstanceRequest
		mockResponse   interface{}
		mockStatusCode int
		expectError    bool
	}{
		{
			name:         "successful update",
			instanceName: "test-db",
			request: UpdateMSSQLInstanceRequest{
				Name:    "test-db",
				Size:    "2Gi",
				Version: DatabaseVersionMSSQL2025_17,
				VPC:     "test-vpc",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockResponse: &MSSQLInstance{
				Name:    "test-db",
				Size:    "2Gi",
				Version: DatabaseVersionMSSQL2025_17,
				VPC:     "test-vpc",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name:         "not found",
			instanceName: "nonexistent-db",
			request: UpdateMSSQLInstanceRequest{
				Name:    "nonexistent-db",
				Size:    "1Gi",
				Version: DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
			},
			mockResponse:   map[string]string{"error": "not found"},
			mockStatusCode: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:         "bad request",
			instanceName: "test-db",
			request: UpdateMSSQLInstanceRequest{
				Name:    "test-db",
				Size:    "invalid",
				Version: DatabaseVersionMSSQL2022_16,
				VPC:     "test-vpc",
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

			server := newServer(tt.mockStatusCode, tt.mockResponse)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).ManagedDB

			instance, err := client.UpdateMSSQLInstance(context.Background(), tt.instanceName, tt.request)
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

func TestMSSQLClient_DeleteInstance(t *testing.T) {
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

			server := newServer(tt.mockStatusCode, nil)
			defer server.Close()

			client := newTestDspcClient(server.URL, authServer.URL).ManagedDB

			err := client.DeleteMSSQLInstance(context.Background(), tt.instanceName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMSSQLClient_CreateInstance_VerifiesRequestBody(t *testing.T) {
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

		var req CreateMSSQLInstanceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}

		assert.Equal(t, "my-mssql", req.Name)
		assert.Equal(t, "1Gi", req.Size)
		assert.Equal(t, DatabaseVersionMSSQL2022_16, req.Version)
		assert.Equal(t, "my-vpc", req.VPC)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(&MSSQLInstance{
			Name:    req.Name,
			Size:    req.Size,
			Version: req.Version,
			VPC:     req.VPC,
		})
	}))
	defer server.Close()

	client := newTestDspcClient(server.URL, authServer.URL).ManagedDB
	instance, err := client.CreateMSSQLInstance(context.Background(), CreateMSSQLInstanceRequest{
		Name:    "my-mssql",
		Size:    "1Gi",
		Version: DatabaseVersionMSSQL2022_16,
		VPC:     "my-vpc",
	})

	assert.NoError(t, err)
	assert.Equal(t, "my-mssql", instance.Name)
}
