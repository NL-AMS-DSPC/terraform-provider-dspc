package client

import (
	"net/http"
	"time"
)

type DspcClient struct {
	VirtualMachines *VirtualMachineService
	BlockStorage    *BlockStorageService
}

func NewDspcClient(apiKey string) *DspcClient {
	apiClient := newApiClient("", apiKey, 30)
	return &DspcClient{
		BlockStorage: NewBlockStorageService(apiClient),
	}
}

type ApiClient struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
}

func newApiClient(endpoint, apiKey string, timeoutSeconds int64) *ApiClient {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeoutSeconds == 0 {
		timeout = 30 * time.Second // default timeout
	}

	return &ApiClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}
