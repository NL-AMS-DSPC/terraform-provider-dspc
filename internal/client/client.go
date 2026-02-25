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
	"strings"
	"sync"
	"time"
)

// DspcClient contains clients for interacting with different resources
type DspcClient struct {
	VirtualMachines *virtualMachineClient
	BlockStorage    *blockStorageClient
	Network         *networkClient
}

// keycloakTokenResponse represents the response from Keycloak token endpoint
type keycloakTokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
}

// authManager handles JWT token authentication with Keycloak
type authManager struct {
	mu           sync.RWMutex
	httpClient   *http.Client
	authURL      string
	org          string
	username     string
	password     string
	accessToken  string
	expiresAt    time.Time
}

// newAuthManager creates a new authentication manager
func newAuthManager(httpClient *http.Client, authURL, org, username, password string) *authManager {
	return &authManager{
		httpClient: httpClient,
		authURL:    authURL,
		org:        org,
		username:   username,
		password:   password,
	}
}

// getToken returns a valid JWT token, refreshing if necessary
func (a *authManager) getToken(ctx context.Context) (string, error) {
	a.mu.RLock()
	// Check if we have a valid token with at least 30 seconds remaining
	if a.accessToken != "" && time.Now().Add(30*time.Second).Before(a.expiresAt) {
		token := a.accessToken
		a.mu.RUnlock()
		return token, nil
	}
	a.mu.RUnlock()

	// Need to acquire new token
	a.mu.Lock()
	defer a.mu.Unlock()

	// Double-check after acquiring write lock
	if a.accessToken != "" && time.Now().Add(30*time.Second).Before(a.expiresAt) {
		return a.accessToken, nil
	}

	// Request new token from Keycloak
	tokenURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", 
		strings.TrimSuffix(a.authURL, "/"), a.org)

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", a.username)
	data.Set("client_secret", a.password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("authentication failed (status %d): %s", resp.StatusCode, string(body))
	}

	var tokenResp keycloakTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("received empty access token from Keycloak")
	}

	// Store the new token
	a.accessToken = tokenResp.AccessToken
	a.expiresAt = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

	return a.accessToken, nil
}

// NewDspcClient Creates and returns a new DSPC client which can be used to interact with different resources
func NewDspcClient(endpoint, namespace, username, password, authURL, org string, timeoutSeconds int64) *DspcClient {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeoutSeconds == 0 {
		timeout = 30 * time.Second
	}

	httpClient := &http.Client{
		Timeout: timeout,
	}

	authMgr := newAuthManager(httpClient, authURL, org, username, password)

	return &DspcClient{
		VirtualMachines: newVirtualMachineClient(endpoint, namespace, authMgr, httpClient),
		BlockStorage:    newBlockStorageClient(endpoint, namespace, authMgr, httpClient),
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

	// Construct path with Envoy gateway prefix and namespace
	pathURL, err := url.Parse(fmt.Sprintf("%s/v1/namespaces/%s%s", c.pathPrefix, c.namespace, path))
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

	if out == nil && (resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusCreated) {
		_ = resp.Body.Close()
		return nil
	}

	respBody, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
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
