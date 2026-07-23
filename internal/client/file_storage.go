package client

import (
	"context"
	"fmt"
	"net/http"
)

// FileStorage represents a file storage volume in the ASC file service API.
type FileStorage struct {
	Name         string `json:"name"`
	Size         string `json:"size"`
	Status       string `json:"status"`
	NFSMountPath string `json:"nfsMountPath,omitempty"`
}

// CreateFileStorageRequest is the request body for creating a file storage.
type CreateFileStorageRequest struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

// CreateFileStorageResponse is the response body from a create file storage request.
type CreateFileStorageResponse struct {
	Created string `json:"created"`
}

// FileStorageAccess represents an access entry that grants a workload NFS access to a file storage.
type FileStorageAccess struct {
	FileStorageName string `json:"fileStorageName"`
	TargetType      string `json:"targetType"`
	TargetName      string `json:"targetName"`
}

// AssignAccessRequest is the request body for assigning workload access to a file storage.
type AssignAccessRequest struct {
	TargetType string `json:"targetType"`
	TargetName string `json:"targetName"`
}

type fileStorageClient struct {
	apiClient
}

// CreateFileStorage creates a new file storage volume.
func (api *fileStorageClient) CreateFileStorage(ctx context.Context, name, size string) (*FileStorage, error) {
	var resp CreateFileStorageResponse
	if err := api.post(ctx, "/v1/file-storages", CreateFileStorageRequest{Name: name, Size: size}, &resp); err != nil {
		return nil, err
	}
	return api.GetFileStorage(ctx, resp.Created)
}

// GetFileStorage retrieves a file storage by name.
func (api *fileStorageClient) GetFileStorage(ctx context.Context, name string) (fs *FileStorage, err error) {
	err = api.get(ctx, fmt.Sprintf("/v1/file-storages/%s", name), &fs)
	return
}

// ListFileStorages retrieves all file storages.
func (api *fileStorageClient) ListFileStorages(ctx context.Context) (storages []*FileStorage, err error) {
	err = api.get(ctx, "/v1/file-storages", &storages)
	return
}

// DeleteFileStorage deletes a file storage by name.
func (api *fileStorageClient) DeleteFileStorage(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/file-storages/%s", name))
}

// AssignAccess grants a workload NFS access to a file storage.
func (api *fileStorageClient) AssignAccess(ctx context.Context, fileStorageName, targetType, targetName string) (*FileStorageAccess, error) {
	var resp FileStorageAccess
	err := api.post(ctx, fmt.Sprintf("/v1/file-storages/%s/access", fileStorageName), AssignAccessRequest{
		TargetType: targetType,
		TargetName: targetName,
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetAccess retrieves a specific access entry for a file storage.
func (api *fileStorageClient) GetAccess(ctx context.Context, fileStorageName, targetType, targetName string) (access *FileStorageAccess, err error) {
	err = api.get(ctx, fmt.Sprintf("/v1/file-storages/%s/access/%s/%s", fileStorageName, targetType, targetName), &access)
	return
}

// RevokeAccess revokes a workload's NFS access to a file storage.
func (api *fileStorageClient) RevokeAccess(ctx context.Context, fileStorageName, targetType, targetName string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/file-storages/%s/access/%s/%s", fileStorageName, targetType, targetName))
}

func newFileStorageClient(endpoint, pathPrefix string, authMgr *authManager, httpClient *http.Client) *fileStorageClient {
	return &fileStorageClient{
		apiClient: newAPIClient(endpoint, "", pathPrefix, authMgr, httpClient),
	}
}
