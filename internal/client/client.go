package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type DspcClient struct {
	VirtualMachines *VirtualMachineService
	BlockStorage    *BlockStorageService
}

func NewDspcClient(endpoint, apiKey string, timeoutSeconds int64) *DspcClient {
	client := newApiClient(endpoint, apiKey, timeoutSeconds)
	return &DspcClient{
		VirtualMachines: NewVirtualMachineService(client),
		BlockStorage:    NewBlockStorageService(client),
	}
}

// MakeRequest makes an HTTP request to the DSPC API
func (c *apiClient) MakeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Construct URL properly
	baseURL, err := url.Parse(c.endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid endpoint URL: %w", err)
	}

	pathURL, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}

	finalURL := baseURL.ResolveReference(pathURL)

	req, err := http.NewRequestWithContext(ctx, method, finalURL.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	return resp, nil
}

type apiClient struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
}

func newApiClient(endpoint, apiKey string, timeoutSeconds int64) *apiClient {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeoutSeconds == 0 {
		timeout = 30 * time.Second // default timeout
	}

	return &apiClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

type requestMaker interface {
	MakeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error)
}
