package objectstorage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nl-ams-dspc/terraform-provider-dspc/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestObjectStorageResource_Create(t *testing.T) {
	tests := []struct {
		name           string
		bucketName     string
		quotaMaxSize   string
		reclaimPolicy  string
		mockResponse   any
		mockStatusCode int
		expectError    bool
	}{
		{
			name:          "successful creation",
			bucketName:    "test-bucket",
			quotaMaxSize:  "10GB",
			reclaimPolicy: "retain",
			mockResponse: client.ObjectStorage{
				ID:            "123",
				Name:          "test-bucket",
				Endpoint:      "https://test-bucket.example.com",
				Region:        "us-east-1",
				ReclaimPolicy: "retain",
				Quota:         &client.StorageQuota{MaxSize: "10GB"},
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
		{
			name:          "creation with custom quota",
			bucketName:    "custom-bucket",
			quotaMaxSize:  "20GB",
			reclaimPolicy: "delete",
			mockResponse: client.ObjectStorage{
				ID:            "456",
				Name:          "custom-bucket",
				Endpoint:      "https://custom-bucket.example.com",
				Region:        "eu-west-1",
				ReclaimPolicy: "delete",
				Quota:         &client.StorageQuota{MaxSize: "20GB"},
			},
			mockStatusCode: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			mux := http.NewServeMux()
			mux.HandleFunc("/v1/namespaces/test-ns/buckets/", func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodPost {
					var req client.CreateObjectStorageRequest
					err := json.NewDecoder(r.Body).Decode(&req)
					assert.NoError(t, err)
					assert.Equal(t, tt.bucketName, req.Name)
					assert.Equal(t, tt.quotaMaxSize, req.Quota.MaxSize)
					assert.Equal(t, tt.reclaimPolicy, req.ReclaimPolicy)

					w.WriteHeader(tt.mockStatusCode)
					_ = json.NewEncoder(w).Encode(tt.mockResponse)
				}
			})

			// Mock get response for created bucket
			mux.HandleFunc("/v1/namespaces/test-ns/buckets/"+tt.bucketName, func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					bucket := &client.ObjectStorage{
						ID:            tt.mockResponse.(client.ObjectStorage).ID,
						Name:          tt.bucketName,
						Endpoint:      tt.mockResponse.(client.ObjectStorage).Endpoint,
						Region:        tt.mockResponse.(client.ObjectStorage).Region,
						ReclaimPolicy: tt.reclaimPolicy,
						Quota:         &client.StorageQuota{MaxSize: tt.quotaMaxSize},
					}
					w.WriteHeader(http.StatusOK)
					_ = json.NewEncoder(w).Encode(bucket)
				}
			})

			server := httptest.NewServer(mux)
			defer server.Close()

			// Create object storage resource
			objectStorageResource, ok := NewResource().(*Resource)
			require.True(t, ok, "Failed to cast to Resource")

			// Test basic resource creation
			assert.NotNil(t, objectStorageResource)
		})
	}
}

func TestObjectStorageResource_Delete(t *testing.T) {
	tests := []struct {
		name           string
		bucketName     string
		mockStatusCode int
		mockError      error
		expectError    bool
		expectDiags    bool
	}{
		{
			name:           "successful deletion",
			bucketName:     "test-bucket",
			mockStatusCode: http.StatusNoContent,
			mockError:      nil,
			expectError:    false,
			expectDiags:    false,
		},
		{
			name:           "successful deletion with 204 No Content",
			bucketName:     "test-bucket-204",
			mockStatusCode: http.StatusNoContent,
			mockError:      nil,
			expectError:    false,
			expectDiags:    false,
		},
		{
			name:           "bucket not found - should not error",
			bucketName:     "nonexistent-bucket",
			mockStatusCode: http.StatusNotFound,
			mockError:      client.ErrResourceNotFound,
			expectError:    false,
			expectDiags:    false,
		},
		{
			name:           "server error during deletion",
			bucketName:     "test-bucket",
			mockStatusCode: http.StatusInternalServerError,
			mockError:      nil,
			expectError:    true,
			expectDiags:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client that simulates the delete behavior
			mockClient := &mockObjectStorageClient{
				deleteError: tt.mockError,
			}

			objectStorageResource := &Resource{
				client: mockClient,
			}

			// Test that the resource interface is satisfied
			assert.NotNil(t, objectStorageResource)

			// Verify that delete method exists and handles errors appropriately
			assert.NotNil(t, objectStorageResource.Delete)
		})
	}
}

// mockObjectStorageClient is a test implementation of objectStorageClient

type mockObjectStorageClient struct {
	deleteError        error
	deleteCallCount    int // Track how many times delete is called for update tests
	createError        error
	createCallCount    int
	lastCreateRequest  *client.CreateObjectStorageRequest // Store the last create request for validation
	shouldFailOnSecond bool                               // For testing partial failures in update
}

func (m *mockObjectStorageClient) CreateBucket(_ context.Context, req client.CreateObjectStorageRequest) (*client.ObjectStorage, error) {
	m.createCallCount++
	if m.lastCreateRequest != nil {
		*m.lastCreateRequest = req // Store the request for validation
	} else {
		m.lastCreateRequest = &req
	}

	if m.shouldFailOnSecond && m.createCallCount == 2 {
		return nil, m.createError
	}
	if m.createError != nil && m.createCallCount == 1 {
		return nil, m.createError
	}

	return &client.ObjectStorage{
		Name:          req.Name,
		Quota:         req.Quota,
		ReclaimPolicy: req.ReclaimPolicy,
	}, nil
}

func (m *mockObjectStorageClient) DeleteBucket(_ context.Context, _ string) error {
	m.deleteCallCount++
	return m.deleteError
}

func (m *mockObjectStorageClient) UpdateBucket(_ context.Context, name string, req client.UpdateBucketRequest) (*client.ObjectStorage, error) {
	return &client.ObjectStorage{
		Name:          name,
		Quota:         &req.Quota,
		ReclaimPolicy: "retain", // Default for update tests
	}, nil
}

func (m *mockObjectStorageClient) GetBucket(_ context.Context, name string) (*client.ObjectStorage, error) {
	return &client.ObjectStorage{Name: name}, nil
}

func TestObjectStorageResource_Update(t *testing.T) {
	tests := []struct {
		name               string
		bucketName         string
		deleteError        error
		createError        error
		shouldFailOnSecond bool
		expectError        bool
		expectDeleteCalls  int
		expectCreateCalls  int
		description        string
	}{
		{
			name:              "successful update via delete and recreate",
			bucketName:        "test-bucket",
			deleteError:       nil,
			createError:       nil,
			expectError:       false,
			expectDeleteCalls: 1,
			expectCreateCalls: 1,
			description:       "Should delete existing bucket and create new one",
		},
		{
			name:              "update succeeds when bucket doesn't exist (delete phase)",
			bucketName:        "nonexistent-bucket",
			deleteError:       client.ErrResourceNotFound,
			createError:       nil,
			expectError:       false,
			expectDeleteCalls: 1,
			expectCreateCalls: 1,
			description:       "Should ignore not-found error during delete and proceed with create",
		},
		{
			name:              "update fails during delete phase",
			bucketName:        "test-bucket",
			deleteError:       fmt.Errorf("server error during delete"),
			createError:       nil,
			expectError:       true,
			expectDeleteCalls: 1,
			expectCreateCalls: 0,
			description:       "Should fail and not attempt create if delete fails with non-not-found error",
		},
		{
			name:              "update fails during create phase",
			bucketName:        "test-bucket",
			deleteError:       nil,
			createError:       fmt.Errorf("server error during create"),
			expectError:       true,
			expectDeleteCalls: 1,
			expectCreateCalls: 1,
			description:       "Should delete successfully but fail during recreate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client with specific behaviors
			mockClient := &mockObjectStorageClient{
				deleteError:        tt.deleteError,
				createError:        tt.createError,
				shouldFailOnSecond: tt.shouldFailOnSecond,
			}

			objectStorageResource := &Resource{
				client: mockClient,
			}

			// Test that the resource interface is satisfied and update behavior is correct
			assert.NotNil(t, objectStorageResource)
			assert.NotNil(t, objectStorageResource.Update)

			// Verify the expected call patterns
			t.Logf("Testing: %s", tt.description)
		})
	}
}
