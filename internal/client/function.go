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
	SKUID       string             `json:"skuID"`
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
func (api *functionClient) CreateFunction(ctx context.Context, name, skuID string) (*Function, error) {
	var response CreateFunctionResponse
	err := api.post(ctx, api.namespacedPath("/virtualmachines/"), CreateFunctionRequest{
		Name:        name,
		SKUID:       skuID,
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
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/virtualmachines/%s", name)))
}

// GetFunction retrieves a function by name (checks if it exists)
func (api *functionClient) GetFunction(ctx context.Context, name string) (function *Function, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/virtualmachines/%s", name)), &function)
	return
}

// ListFunctions retrieves all functions
func (api *functionClient) ListFunctions(ctx context.Context) (functions []*Function, err error) {
	err = api.get(ctx, api.namespacedPath("/virtualmachines"), &functions)
	return
}

func newFunctionClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *functionClient {
	return &functionClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
