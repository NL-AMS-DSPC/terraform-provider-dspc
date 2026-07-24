package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Function represents a function in the ASC API.
// Secrets and RegistryAuth are write-only: the API never returns their values on read.
type Function struct {
	ID                  string          `json:"id,omitempty"`
	TenantID            string          `json:"tenantId,omitempty"`
	Name                string          `json:"name"`
	Image               string          `json:"image"`
	Port                int32           `json:"port,omitempty"`
	Env                 []EnvVar        `json:"env,omitempty"`
	Secrets             []RuntimeSecret `json:"secrets,omitempty"`
	RegistryAuth        *RegistryAuth   `json:"registryAuth,omitempty"`
	Resources           *Resources      `json:"resources,omitempty"`
	Concurrency         *Concurrency    `json:"concurrency,omitempty"`
	HealthChecks        *HealthChecks   `json:"healthChecks,omitempty"`
	Tags                []Tag           `json:"tags,omitempty"`
	URL                 string          `json:"url,omitempty"`
	Status              string          `json:"status,omitempty"`
	LatestReadyRevision string          `json:"latestReadyRevision,omitempty"`
	CreatedAt           *time.Time      `json:"createdAt,omitempty"`
	UpdatedAt           *time.Time      `json:"updatedAt,omitempty"`
}

// EnvVar defines a plain-text environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Resources defines CPU and memory requests and limits
type Resources struct {
	CPURequest    string `json:"cpuRequest,omitempty"`
	CPULimit      string `json:"cpuLimit,omitempty"`
	MemoryRequest string `json:"memoryRequest,omitempty"`
	MemoryLimit   string `json:"memoryLimit,omitempty"`
}

// Concurrency controls the maximum number of concurrent requests handled by a function instance
type Concurrency struct {
	Limit *int64 `json:"limit,omitempty"`
}

// HealthChecks defines optional liveness and readiness probes
type HealthChecks struct {
	Liveness  *Probe `json:"liveness,omitempty"`
	Readiness *Probe `json:"readiness,omitempty"`
}

// Probe defines HTTP probe settings for a function container
type Probe struct {
	Path                string `json:"path,omitempty"`
	Port                int32  `json:"port,omitempty"`
	InitialDelaySeconds int32  `json:"initialDelaySeconds,omitempty"`
	PeriodSeconds       int32  `json:"periodSeconds,omitempty"`
	TimeoutSeconds      int32  `json:"timeoutSeconds,omitempty"`
	FailureThreshold    int32  `json:"failureThreshold,omitempty"`
}

// CreateFunctionRequest represents the request body for creating a function
type CreateFunctionRequest struct {
	Name         string          `json:"name"`
	Image        string          `json:"image"`
	Port         int32           `json:"port,omitempty"`
	Env          []EnvVar        `json:"env,omitempty"`
	Secrets      []RuntimeSecret `json:"secrets,omitempty"`
	RegistryAuth *RegistryAuth   `json:"registryAuth,omitempty"`
	Resources    *Resources      `json:"resources,omitempty"`
	Concurrency  *Concurrency    `json:"concurrency,omitempty"`
	HealthChecks *HealthChecks   `json:"healthChecks,omitempty"`
	Tags         []Tag           `json:"tags,omitempty"`
}

// UpdateFunctionRequest represents the request body for updating a function
type UpdateFunctionRequest struct {
	Image        string          `json:"image"`
	Port         int32           `json:"port,omitempty"`
	Env          []EnvVar        `json:"env,omitempty"`
	Secrets      []RuntimeSecret `json:"secrets,omitempty"`
	RegistryAuth *RegistryAuth   `json:"registryAuth,omitempty"`
	Resources    *Resources      `json:"resources,omitempty"`
	Concurrency  *Concurrency    `json:"concurrency,omitempty"`
	HealthChecks *HealthChecks   `json:"healthChecks,omitempty"`
	Tags         []Tag           `json:"tags,omitempty"`
}

type functionClient struct {
	apiClient
}

// CreateFunction creates a new function. The API returns the created function wrapped in
// {"data":...} with a 201; the name is already known, so we ignore the POST body and
// re-fetch by name to hydrate full details (mirrors the container client).
func (api *functionClient) CreateFunction(ctx context.Context, req CreateFunctionRequest) (*Function, error) {
	if err := api.post(ctx, "/v1/functions/", req, nil); err != nil {
		return nil, err
	}
	return api.GetFunction(ctx, req.Name)
}

// UpdateFunction updates an existing function and re-fetches it for full details.
func (api *functionClient) UpdateFunction(ctx context.Context, name string, req UpdateFunctionRequest) (*Function, error) {
	if err := api.put(ctx, fmt.Sprintf("/v1/functions/%s", name), req, nil); err != nil {
		return nil, err
	}
	return api.GetFunction(ctx, name)
}

// DeleteFunction deletes a function by name
func (api *functionClient) DeleteFunction(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/functions/%s", name))
}

// GetFunction retrieves a function by name. The API wraps the body in {"data":...}.
func (api *functionClient) GetFunction(ctx context.Context, name string) (*Function, error) {
	var wrapper struct {
		Data *Function `json:"data"`
	}
	if err := api.get(ctx, fmt.Sprintf("/v1/functions/%s", name), &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Data, nil
}

// ListFunctions retrieves all functions. The API wraps the list in {"data":{"functions":[...]}}.
func (api *functionClient) ListFunctions(ctx context.Context) ([]*Function, error) {
	var wrapper struct {
		Data struct {
			Functions []*Function `json:"functions"`
		} `json:"data"`
	}
	if err := api.get(ctx, "/v1/functions", &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Data.Functions, nil
}

func newFunctionClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *functionClient {
	return &functionClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
