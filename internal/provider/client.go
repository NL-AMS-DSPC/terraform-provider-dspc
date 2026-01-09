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
