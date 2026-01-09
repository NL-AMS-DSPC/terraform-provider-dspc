package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

type DspcClient struct {
	VirtualMachines *VirtualMachineService
	BlockStorage    *BlockStorageService
}

func NewDspcClient(endpoint, apiKey string, timeoutSeconds int64) *DspcClient {

	vmsClient := newApiClient[VM](endpoint, apiKey, timeoutSeconds)
	blocksClient := newApiClient[Block](endpoint, apiKey, timeoutSeconds)

	return &DspcClient{
		VirtualMachines: NewVirtualMachineService(vmsClient),
		BlockStorage:    NewBlockStorageService(blocksClient),
	}
}

// MakeRequest makes an HTTP request to the DSPC API
func (c *apiClient[T]) MakeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
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

func (c *apiClient[T]) Create(ctx context.Context, path string, body interface{}) (*T, error) {
	resp, err := c.MakeRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var respObj T
	if err := json.NewDecoder(resp.Body).Decode(&respObj); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &respObj, nil
}

func (c *apiClient[T]) Get(ctx context.Context, path string) (*T, error) {
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var entity *T
	if err := json.NewDecoder(resp.Body).Decode(&entity); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return entity, nil
}

func (c *apiClient[T]) List(ctx context.Context, path string) ([]*T, error) {
	resp, err := c.MakeRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var entities []*T
	if err := json.NewDecoder(resp.Body).Decode(&entities); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return entities, nil
}

func (c *apiClient[T]) Delete(ctx context.Context, path string) error {
	resp, err := c.MakeRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

type apiClient[T any] struct {
	httpClient *http.Client
	endpoint   string
	apiKey     string
}

func newApiClient[T any](endpoint, apiKey string, timeoutSeconds int64) *apiClient[T] {
	timeout := time.Duration(timeoutSeconds) * time.Second
	if timeoutSeconds == 0 {
		timeout = 30 * time.Second // default timeout
	}

	return &apiClient[T]{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		endpoint: endpoint,
		apiKey:   apiKey,
	}
}

type requestMaker[T any] interface {
	Create(ctx context.Context, path string, body interface{}) (*T, error)
	Get(ctx context.Context, path string) (*T, error)
	Delete(ctx context.Context, path string) error
	List(ctx context.Context, path string) ([]*T, error)
	MakeRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error)
}
