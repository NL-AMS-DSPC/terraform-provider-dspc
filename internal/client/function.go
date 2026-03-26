package client

import (
	"context"
	"fmt"
	"net/http"
)

// Function represents a function in the DSPC API
type Function struct {
	Name   string `json:"name"`
	SKU    SKU    `json:"sku"`
	Status string `json:"status"`
}

// CreateFunctionRequest represents the request body for creating a function
type CreateFunctionRequest struct {
	Name        string             `json:"name"`
	Image       string             `json:"image"`
	Autoscaling *AutoscalingConfig `json:"autoscaling,omitempty"`
}

// CreateFunctionResponse represents the response from creating a function
type CreateFunctionResponse struct {
	Created string `json:"created"`
}

// DeleteFunctionResponse represents the response from deleting a function
type DeleteFunctionResponse struct {
	Deleted string `json:"deleted"`
}

type functionClient struct {
	apiClient
}

// CreateFunction creates a new function
func (api *functionClient) CreateFunction(ctx context.Context, name, image string) (*Function, error) {
	var response CreateFunctionResponse
	err := api.post(ctx, api.namespacedPath("/functions/"), CreateFunctionRequest{
		Name:        name,
		Image:       image,
		Autoscaling: nil, // Functions don't use autoscaling
	}, &response)
	if err != nil {
		return nil, err
	}
	// Fetch the created function to get full details
	return api.GetFunction(ctx, response.Created)
}

// DeleteFunction deletes a function by name
func (api *functionClient) DeleteFunction(ctx context.Context, name string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/functions/%s", name)))
}

// GetFunction retrieves a function by name (checks if it exists)
func (api *functionClient) GetFunction(ctx context.Context, name string) (function *Function, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/functions/%s", name)), &function)
	return
}

// GetFunctionInNamespace retrieves a function by name from a specific namespace
func (api *functionClient) GetFunctionInNamespace(ctx context.Context, name, namespace string) (function *Function, err error) {
	err = api.get(ctx, api.customNamespacedPath(namespace, fmt.Sprintf("/functions/%s", name)), &function)
	return
}

// CreateFunctionInNamespace creates a new function in a specific namespace
func (api *functionClient) CreateFunctionInNamespace(ctx context.Context, name, image, namespace string) (*Function, error) {
	var response CreateFunctionResponse
	err := api.post(ctx, api.customNamespacedPath(namespace, "/functions/"), CreateFunctionRequest{
		Name:        name,
		Image:       image,
		Autoscaling: nil, // Functions don't use autoscaling
	}, &response)
	if err != nil {
		return nil, err
	}
	// Fetch the created function to get full details
	return api.GetFunctionInNamespace(ctx, response.Created, namespace)
}

// DeleteFunctionInNamespace deletes a function by name from a specific namespace
func (api *functionClient) DeleteFunctionInNamespace(ctx context.Context, name, namespace string) error {
	return api.delete(ctx, api.customNamespacedPath(namespace, fmt.Sprintf("/functions/%s", name)))
}

// ListFunctions retrieves all functions
func (api *functionClient) ListFunctions(ctx context.Context) (functions []*Function, err error) {
	err = api.get(ctx, api.namespacedPath("/functions"), &functions)
	return
}

// ListFunctionsInNamespace retrieves all functions from a specific namespace
func (api *functionClient) ListFunctionsInNamespace(ctx context.Context, namespace string) (functions []*Function, err error) {
	err = api.get(ctx, api.customNamespacedPath(namespace, "/functions"), &functions)
	return
}

// customNamespacedPath creates a path with a custom namespace instead of the client's default namespace
func (api *functionClient) customNamespacedPath(namespace, path string) string {
	return fmt.Sprintf("/v1/namespaces/%s%s", namespace, path)
}

func newFunctionClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *functionClient {
	return &functionClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
