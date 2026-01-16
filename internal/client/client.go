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

// DspcClient contains clients for interacting with different resources
type DspcClient struct {
	VirtualMachines *virtualMachineClient
	BlockStorage    *blockStorageClient
}

// NewDspcClient Creates and returns a new DSPC client which can be used to interact with different resources
func NewDspcClient(endpoint, namespace, apiKey string, timeoutSeconds int64) *DspcClient {
	return &DspcClient{
		VirtualMachines: newVirtualMachineClient(endpoint, namespace, apiKey, timeoutSeconds),
		BlockStorage:    newBlockStorageClient(endpoint, namespace, apiKey, timeoutSeconds),
	}
}

func (c *apiClient) post(ctx context.Context, path string, body interface{}, out interface{}) error {
	return c.makeRequest(ctx, http.MethodPost, path, body, out)
}

func (c *apiClient) put(ctx context.Context, path string, body interface{}, out interface{}) error {
	return c.makeRequest(ctx, http.MethodPut, path, body, out)
}

func (c *apiClient) get(ctx context.Context, path string, out interface{}) error {
	return c.makeRequest(ctx, http.MethodGet, path, nil, out)
}

func (c *apiClient) delete(ctx context.Context, path string) error {
	return c.makeRequest(ctx, http.MethodDelete, path, nil, nil)
}

// makeRequest makes an HTTP request to the DSPC API
func (c *apiClient) makeRequest(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	// Construct URL properly
	baseURL, err := url.Parse(c.endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint URL: %w", err)
	}

	// add prefixed `/v1/namespaces/{namespace}/` to the url
	pathURL, err := url.Parse(fmt.Sprintf("/v1/namespaces/%s%s", c.namespace, path))
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	finalURL := baseURL.ResolveReference(pathURL)

	req, err := http.NewRequestWithContext(ctx, method, finalURL.String(), reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	var respBody []byte
	shouldReadBody := out != nil || resp.StatusCode != http.StatusOK
	if shouldReadBody {
		respBody, err = io.ReadAll(resp.Body)

		_ = resp.Body.Close()

		if err != nil {
			return err
		}
	}

	_ = resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil {
		if err := json.Unmarshal(respBody, &out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

type apiClient struct {
	httpClient *http.Client
	endpoint   string
	namespace  string
	apiKey     string
}

func newApiClient(endpoint, namespace, apiKey string, timeoutSeconds int64) apiClient {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeoutSeconds == 0 {
		timeout = 30 * time.Second // default timeout
	}

	return apiClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		endpoint:  endpoint,
		namespace: namespace,
		apiKey:    apiKey,
	}
}
