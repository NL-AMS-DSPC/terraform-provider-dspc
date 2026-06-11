package client

import (
	"context"
	"fmt"
	"net/http"
)

// ObjectStorage represents an object storage bucket
type ObjectStorage struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Quota         *StorageQuota `json:"quota,omitempty"`
	Tags          []Tag         `json:"tags,omitempty"`
	ReclaimPolicy string        `json:"reclaimPolicy,omitempty"`
	Endpoint      string        `json:"endpoint,omitempty"`
	Region        string        `json:"region,omitempty"`
	CreatedAt     string        `json:"createdAt,omitempty"`
	UpdatedAt     string        `json:"updatedAt,omitempty"`
}

// StorageQuota represents the storage quota for a bucket
type StorageQuota struct {
	MaxSize string `json:"maxSize,omitempty"`
}

// CreateObjectStorageRequest is the request body for creating a bucket
type CreateObjectStorageRequest struct {
	Name          string        `json:"name"`
	Quota         *StorageQuota `json:"quota,omitempty"`
	Tags          []TagDTO      `json:"tags,omitempty"`
	ReclaimPolicy string        `json:"reclaimPolicy,omitempty"`
}

// UpdateBucketRequest is the request body for updating a bucket
type UpdateBucketRequest struct {
	Quota StorageQuota `json:"quota"`
}

// ObjectStorageListResponse represents the response for listing buckets
type ObjectStorageListResponse struct {
	Buckets []ObjectStorage `json:"data"`
}

// StorageCredentialsResponse represents the response for bucket credentials
type StorageCredentialsResponse struct {
	AccessKeyID     string `json:"accessKeyID"`
	SecretAccessKey string `json:"secretAccessKey"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
}

type objectStorageClient struct {
	apiClient
}

// ListBuckets retrieves all object storage buckets
func (api *objectStorageClient) ListBuckets(ctx context.Context) ([]ObjectStorage, error) {
	var response ObjectStorageListResponse
	err := api.get(ctx, "/v1/object-storage", &response)
	if err != nil {
		return nil, err
	}
	return response.Buckets, nil
}

// GetBucket retrieves a specific object storage bucket by ID
func (api *objectStorageClient) GetBucket(ctx context.Context, id string) (*ObjectStorage, error) {
	var response ObjectStorage
	err := api.get(ctx, fmt.Sprintf("/v1/object-storage/%s", id), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// CreateBucket creates a new object storage bucket
func (api *objectStorageClient) CreateBucket(ctx context.Context, req CreateObjectStorageRequest) (*ObjectStorage, error) {
	var response ObjectStorage
	err := api.post(ctx, "/v1/object-storage", req, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// PatchBucket applies tag mutations to a specific object storage bucket
func (api *objectStorageClient) PatchBucket(ctx context.Context, id string, req PatchTagsRequest) (*ObjectStorage, error) {
	var response ObjectStorage
	err := api.patch(ctx, fmt.Sprintf("/v1/object-storage/%s", id), req, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// UpdateBucket updates a specific object storage bucket
func (api *objectStorageClient) UpdateBucket(ctx context.Context, id string, req UpdateBucketRequest) (*ObjectStorage, error) {
	var response ObjectStorage
	err := api.put(ctx, fmt.Sprintf("/v1/object-storage/%s", id), req, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// DeleteBucket deletes an object storage bucket
func (api *objectStorageClient) DeleteBucket(ctx context.Context, id string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/object-storage/%s", id))
}

func newObjectStorageClient(endpoint, pathPrefix string, authMgr *authManager, httpClient *http.Client) *objectStorageClient {
	return &objectStorageClient{
		apiClient: newAPIClient(endpoint, "", pathPrefix, authMgr, httpClient),
	}
}
