package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Function represents a function in the DSPC API
type Function struct {
	ID                  string         `json:"id,omitempty"`
	Name                string         `json:"name"`
	Image               string         `json:"image"`
	Port                int32          `json:"port,omitempty"`
	Env                 []EnvVar       `json:"env,omitempty"`
	Secrets             []SecretEnvVar `json:"secrets,omitempty"`
	Resources           *Resources     `json:"resources,omitempty"`
	Concurrency         *Concurrency   `json:"concurrency,omitempty"`
	HealthChecks        *HealthChecks  `json:"healthChecks,omitempty"`
	Tags                []Tag          `json:"tags,omitempty"`
	URL                 string         `json:"url,omitempty"`
	Status              string         `json:"status,omitempty"`
	LatestReadyRevision string         `json:"latestReadyRevision,omitempty"`
	CreatedAt           *time.Time     `json:"createdAt,omitempty"`
	UpdatedAt           *time.Time     `json:"updatedAt,omitempty"`
}

// EnvVar defines a plain-text environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SecretEnvVar defines an environment variable sourced from a secret key
type SecretEnvVar struct {
	Name    string `json:"name"`
	Key     string `json:"key"`
	EnvName string `json:"envName"`
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
	Name         string         `json:"name"`
	Image        string         `json:"image"`
	Port         int32          `json:"port,omitempty"`
	Env          []EnvVar       `json:"env,omitempty"`
	Secrets      []SecretEnvVar `json:"secrets,omitempty"`
	Resources    *Resources     `json:"resources,omitempty"`
	Concurrency  *Concurrency   `json:"concurrency,omitempty"`
	HealthChecks *HealthChecks  `json:"healthChecks,omitempty"`
	Tags         []Tag          `json:"tags,omitempty"`
}

// CreateFunctionResponse represents the response from creating a function
type CreateFunctionResponse struct {
	Created string `json:"created"`
}

// UpdateFunctionRequest represents the request body for updating a function
type UpdateFunctionRequest struct {
	Image        string         `json:"image"`
	Port         int32          `json:"port,omitempty"`
	Env          []EnvVar       `json:"env,omitempty"`
	Secrets      []SecretEnvVar `json:"secrets,omitempty"`
	Resources    *Resources     `json:"resources,omitempty"`
	Concurrency  *Concurrency   `json:"concurrency,omitempty"`
	HealthChecks *HealthChecks  `json:"healthChecks,omitempty"`
	Tags         []Tag          `json:"tags,omitempty"`
}

// UpdateFunctionResponse represents the response from updating a function
type UpdateFunctionResponse struct {
	Updated string `json:"updated"`
}

// DeleteFunctionResponse represents the response from deleting a function
type DeleteFunctionResponse struct {
	Deleted string `json:"deleted"`
}

type functionClient struct {
	apiClient
}

// CreateFunction creates a new function
func (api *functionClient) CreateFunction(ctx context.Context, req CreateFunctionRequest) (*Function, error) {
	var response CreateFunctionResponse
	err := api.post(ctx, "/v1/functions/", req, &response)
	if err != nil {
		return nil, err
	}
	// Fetch the created function to get full details
	return api.GetFunction(ctx, response.Created)
}

// UpdateFunction updates an existing function
func (api *functionClient) UpdateFunction(ctx context.Context, name string, req UpdateFunctionRequest) (*Function, error) {
	var response UpdateFunctionResponse
	err := api.put(ctx, fmt.Sprintf("/v1/functions/%s", name), req, &response)
	if err != nil {
		return nil, err
	}
	// Fetch the updated function to get full details
	return api.GetFunction(ctx, name)
}

// DeleteFunction deletes a function by name
func (api *functionClient) DeleteFunction(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/v1/functions/%s", name))
}

// GetFunction retrieves a function by name
func (api *functionClient) GetFunction(ctx context.Context, name string) (function *Function, err error) {
	err = api.get(ctx, fmt.Sprintf("/v1/functions/%s", name), &function)
	return
}

// ListFunctions retrieves all functions
func (api *functionClient) ListFunctions(ctx context.Context) (functions []*Function, err error) {
	err = api.get(ctx, "/v1/functions", &functions)
	return
}

func newFunctionClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *functionClient {
	return &functionClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
