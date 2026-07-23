package client

import (
	"context"
	"net/http"
)

// SKUResponse represents basic SKU information in API responses.
type SKUResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Family      string `json:"family"`
	Threads     uint64 `json:"threads"`
	Cores       uint64 `json:"cores"`
	MemoryInMB  uint64 `json:"memoryInMB"`
	StorageInGB uint64 `json:"storageInGB"`
	StorageType string `json:"storageType"`
	GPUCount    uint64 `json:"GPUCount"`
	GPUType     string `json:"GPUType"`
}

type skuClient struct {
	apiClient
}

// ListSKUs retrieves all available SKUs
func (api *skuClient) ListSKUs(ctx context.Context) (skus []SKUResponse, err error) {
	err = api.get(ctx, api.namespacedPath("/skus"), &skus)
	return
}

func newSkuClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *skuClient {
	return &skuClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
