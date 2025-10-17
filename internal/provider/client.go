package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

// Client represents the DSPC API client
type Client struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
}

// VM represents a virtual machine in the DSPC API
type VM struct {
	Name string `json:"vmName"`
}

// CreateVMResponse represents the response from creating a VM
type CreateVMResponse struct {
	Created string `json:"created"`
}

// DeleteVMResponse represents the response from deleting a VM
type DeleteVMResponse struct {
	Deleted string `json:"deleted"`
}

// NewClient creates a new DSPC API client
func NewClient(endpoint, apiKey string, timeoutSeconds int64) *Client {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeoutSeconds == 0 {
		timeout = 30 * time.Second // default timeout
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

// NewClientFromConfig creates a client from provider configuration with environment variable fallbacks
func NewClientFromConfig(config DspcProviderModel) (*Client, error) {
	var endpoint, apiKey string
	var timeoutSeconds int64

	// Extract endpoint with environment fallback
	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	}
	if endpoint == "" {
		endpoint = os.Getenv("DSPC_ENDPOINT")
	}

	// Validate that endpoint is provided
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is required but not provided. Please set the 'endpoint' attribute " +
			"in the provider configuration or set the DSPC_ENDPOINT environment variable")
	}

	// Extract API key with environment fallback
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		apiKey = os.Getenv("DSPC_API_KEY")
	}

	// Validate that API key is provided
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required but not provided. Please set the 'api_key' attribute " +
			"in the provider configuration or set the DSPC_API_KEY environment variable")
	}

	// Extract timeout with defaults
	if !config.Timeout.IsNull() {
		timeoutSeconds = config.Timeout.ValueInt64()
	}
	if timeoutSeconds == 0 {
		if envTimeout := os.Getenv("DSPC_TIMEOUT"); envTimeout != "" {
			if parsedTimeout, err := strconv.ParseInt(envTimeout, 10, 64); err == nil {
				timeoutSeconds = parsedTimeout
			}
		}
		if timeoutSeconds == 0 {
			timeoutSeconds = 30 // default
		}
	}

	return NewClient(endpoint, apiKey, timeoutSeconds), nil
}

// makeRequest makes an HTTP request to the DSPC API
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
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

// CreateVM creates a new virtual machine
func (c *Client) CreateVM(ctx context.Context, name string) (*VM, error) {
	vm := VM{Name: name}
	resp, err := c.makeRequest(ctx, "POST", "/virtualmachine", vm)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var createResp CreateVMResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &VM{Name: createResp.Created}, nil
}

// DeleteVM deletes a virtual machine by name
func (c *Client) DeleteVM(ctx context.Context, name string) error {
	vm := VM{Name: name}
	resp, err := c.makeRequest(ctx, "DELETE", "/virtualmachine", vm)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetVM retrieves a virtual machine by name (checks if it exists)
func (c *Client) GetVM(ctx context.Context, name string) (*VM, error) {
	vms, err := c.ListVMs(ctx)
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.Name == name {
			return vm, nil
		}
	}

	return nil, fmt.Errorf("VM '%s' not found. Please verify the VM name exists or check your API endpoint", name)
}

// ListVMs retrieves all virtual machines
func (c *Client) ListVMs(ctx context.Context) ([]*VM, error) {
	resp, err := c.makeRequest(ctx, "GET", "/virtualmachine", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var vms []*VM
	if err := json.NewDecoder(resp.Body).Decode(&vms); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return vms, nil
}
