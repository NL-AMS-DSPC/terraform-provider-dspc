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
				SkuSize: "gp-2",
				Version: DatabaseVersionMSSQL2022_16,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9a",
			},
			mockResponse: &MSSQLInstance{
				Name:    "test-db",
				SkuSize: "gp-2",
				Version: DatabaseVersionMSSQL2022_16,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9a",
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "creation with tags",
			request: CreateMSSQLInstanceRequest{
				Name:    "tagged-db",
				SkuSize: "c-8",
				Version: DatabaseVersionMSSQL2019_15,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9b",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockResponse: &MSSQLInstance{
				Name:    "tagged-db",
				SkuSize: "c-8",
				Version: DatabaseVersionMSSQL2019_15,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9b",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name: "conflict error",
			request: CreateMSSQLInstanceRequest{
				Name:    "existing-db",
				SkuSize: "gp-2",
				Version: DatabaseVersionMSSQL2022_16,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9a",
			},
			mockResponse:   map[string]string{"error": "already exists"},
			mockStatusCode: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "bad request",
			request: CreateMSSQLInstanceRequest{
				Name:    "bad-db",
				SkuSize: "invalid",
				Version: "UNKNOWN",
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9a",
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

			client := newTestDspcClient(server.URL, authServer.URL).ManagedDB

			instance, err := client.CreateMSSQLInstance(context.Background(), tt.request)
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
				SkuSize: "gp-2",
				Version: DatabaseVersionMSSQL2022_16,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9a",
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
					{Name: "db-1", SkuSize: "gp-2", Version: DatabaseVersionMSSQL2022_16, VPCID: "vpc-a"},
					{Name: "db-2", SkuSize: "gp-4", Version: DatabaseVersionMSSQL2019_15, VPCID: "vpc-b"},
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

			server := newMockServer(tt.mockStatusCode, tt.mockResponse)
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
				SkuSize: "gp-4",
				Version: DatabaseVersionMSSQL2025_17,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9a",
				Tags:    []Tag{{Key: "env", Value: "prod"}},
			},
			mockResponse: &MSSQLInstance{
				Name:    "test-db",
				SkuSize: "gp-4",
				Version: DatabaseVersionMSSQL2025_17,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9a",
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
				SkuSize: "gp-2",
				Version: DatabaseVersionMSSQL2022_16,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9a",
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
				SkuSize: "invalid",
				Version: DatabaseVersionMSSQL2022_16,
				VPCID:   "da1cfaf5-bee5-4f06-8323-e3bd6daead9a",
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

			client := newTestDspcClient(server.URL, authServer.URL).ManagedDB

			instance, err := client.UpdateMSSQLInstance(context.Background(), tt.instanceName, tt.request)
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

			server := newMockServer(tt.mockStatusCode, nil)
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
		assert.Equal(t, "gp-2", req.SkuSize)
		assert.Equal(t, DatabaseVersionMSSQL2022_16, req.Version)
		assert.Equal(t, "11111111-1111-1111-1111-111111111113", req.VPCID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(&MSSQLInstance{
			Name:    req.Name,
			SkuSize: req.SkuSize,
			Version: req.Version,
			VPCID:   req.VPCID,
		})
	}))
	defer server.Close()

	client := newTestDspcClient(server.URL, authServer.URL).ManagedDB
	instance, err := client.CreateMSSQLInstance(context.Background(), CreateMSSQLInstanceRequest{
		Name:    "my-mssql",
		SkuSize: "gp-2",
		Version: DatabaseVersionMSSQL2022_16,
		VPCID:   "11111111-1111-1111-1111-111111111113",
	})

	assert.NoError(t, err)
	assert.Equal(t, "my-mssql", instance.Name)
}
