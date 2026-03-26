package client

import (
	"context"
	"fmt"
	"net/http"
)

// ContainerTag represents a key-value tag on a container deployment
type ContainerTag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Container represents a container deployment in the DSPC container API.
// Used for both create requests (ID is omitted) and API responses.
type Container struct {
	ID         string         `json:"id,omitempty"`
	Name       string         `json:"name"`
	Image      string         `json:"image"`
	Command    string         `json:"command,omitempty"`
	Port       int32          `json:"port"`
	Args       []string       `json:"args,omitempty"`
	Env        []string       `json:"env,omitempty"`
	WorkingDir string         `json:"working_dir,omitempty"`
	User       string         `json:"user,omitempty"`
	Group      string         `json:"group,omitempty"`
	Replicas   int32          `json:"replicas,omitempty"`
	Tags       []ContainerTag `json:"tags,omitempty"`
}

type containerClient struct {
	apiClient
}

// CreateDeployment creates a new container deployment
func (api *containerClient) CreateDeployment(ctx context.Context, req Container) (container *Container, err error) {
	err = api.post(ctx, api.namespacedPath("/deployments"), req, &container)
	return
}

// GetDeployment retrieves a container deployment by name
func (api *containerClient) GetDeployment(ctx context.Context, name string) (container *Container, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/deployments/%s", name)), &container)
	return
}

// DeleteDeployment deletes a container deployment by name
func (api *containerClient) DeleteDeployment(ctx context.Context, name string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/deployments/%s", name)))
}

func newContainerClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *containerClient {
	return &containerClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
