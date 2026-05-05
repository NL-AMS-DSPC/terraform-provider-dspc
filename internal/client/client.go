// Package client provides a client for interacting with the DSPC API.
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

const (
	minTokenLifetime     = 30 * time.Second
	defaultClientTimeout = 30 * time.Second
)

// DspcClient contains clients for interacting with different resources
type DspcClient struct {
	VirtualMachines *virtualMachineClient
	BlockStorage    *blockStorageClient
	Network         *networkClient
	ManagedDB       *managedDatabaseClient
	Authorization   *authorizationClient
	Functions       *functionClient
	Containers      *containerClient
}

// NewDspcClient Creates and returns a new DSPC client which can be used to interact with different resources
func NewDspcClient(endpoint, namespace, username, password, authURL, org string, timeoutSeconds int64) *DspcClient {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeoutSeconds == 0 {
		timeout = defaultClientTimeout
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	authMgr := newAuthManager(httpClient, authURL, org, username, password)

	// Load service configuration with environment variable overrides
	config := LoadServiceConfig()

	return &DspcClient{
		VirtualMachines: newVirtualMachineClient(endpoint, namespace, config.VM.PathPrefix, authMgr, httpClient),
		BlockStorage:    newBlockStorageClient(endpoint, namespace, config.BlockStorage.PathPrefix, authMgr, httpClient),
		Network:         newNetworkClient(endpoint, namespace, config.Network.PathPrefix, authMgr, httpClient),
		ManagedDB:       newManagedDatabaseClient(endpoint, namespace, config.ManagedDB.PathPrefix, authMgr, httpClient),
		Authorization:   newAuthorizationClient(endpoint, config.Authorization.PathPrefix, authMgr, httpClient),
		Functions:       newFunctionClient(endpoint, namespace, config.Function.PathPrefix, authMgr, httpClient),
		Containers:      newContainerClient(endpoint, namespace, config.Container.PathPrefix, authMgr, httpClient),
	}
}

func (c *apiClient) post(ctx context.Context, path string, body any, out any) error {
	return c.makeRequest(ctx, http.MethodPost, path, body, out)
}

func (c *apiClient) put(ctx context.Context, path string, body any, out any) error {
	return c.makeRequest(ctx, http.MethodPut, path, body, out)
}

func (c *apiClient) get(ctx context.Context, path string, out any) error {
	return c.makeRequest(ctx, http.MethodGet, path, nil, out)
}

func (c *apiClient) delete(ctx context.Context, path string) error {
	return c.makeRequest(ctx, http.MethodDelete, path, nil, nil)
}

// makeRequest makes an HTTP request to the DSPC API
func (c *apiClient) makeRequest(ctx context.Context, method, path string, body any, out any) error {
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

	// Construct path with gateway prefix
	pathURL, err := url.Parse(fmt.Sprintf("%s%s", c.pathPrefix, path))
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

	// Get JWT token for authorization
	token, err := c.authManager.getToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get authentication token: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	// #nosec G704 -- Endpoint is from trusted Terraform provider configuration, not user input
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}

	if out == nil && isSuccessStatus(resp.StatusCode) {
		_ = resp.Body.Close()
		return nil
	}

	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
	}

	if resp.StatusCode == http.StatusNotFound {
		// Map 404 responses to ErrResourceNotFound so callers can use errors.Is.
		if len(respBody) == 0 {
			return ErrResourceNotFound
		}
		return fmt.Errorf("%w: %s", ErrResourceNotFound, string(respBody))
	}
	if !isSuccessStatus(resp.StatusCode) {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

type apiClient struct {
	httpClient  *http.Client
	endpoint    string
	namespace   string
	pathPrefix  string
	authManager *authManager
}

func newAPIClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) apiClient {
	return apiClient{
		httpClient:  httpClient,
		endpoint:    endpoint,
		namespace:   namespace,
		pathPrefix:  pathPrefix,
		authManager: authMgr,
	}
}

// Prefixes path with /v1/namespaces/{namespace}
func (c *apiClient) namespacedPath(path string) string {
	return fmt.Sprintf("/v1/namespaces/%s%s", c.namespace, path)
}

// Prefixes path with /v1 without a namespace segment.
func (c *apiClient) path(path string) string {
	return "/v1" + path
}

// isSuccessStatus reports whether the HTTP status code indicates a successful response (2xx).
func isSuccessStatus(statusCode int) bool {
	return statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices
}
