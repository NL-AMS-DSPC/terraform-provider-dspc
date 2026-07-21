package client

import (
	"context"
	"fmt"
	"net/http"
)

// VMGroup represents a virtual machine group.
type VMGroup struct {
	URN               string             `json:"urn"`
	Name              string             `json:"name"`
	SKU               SKUResponse        `json:"sku"`
	Status            string             `json:"status"`
	LastError         string             `json:"lastError"`
	Tags              []Tag              `json:"tags"`
	AttachedVolumes   []string           `json:"attachedVolumes"`
	AutoscalingPolicy *AutoscalingPolicy `json:"autoscalingPolicy"`
}

// AutoscalingPolicy contains all available scaler configurations
type AutoscalingPolicy struct {
	MinReplicas int       `json:"min_replicas"`
	MaxReplicas int       `json:"max_replicas"`
	CronRule    *CronRule `json:"cron,omitempty"`
}

// CronRule represents cron-based scheduling configuration for scaling
type CronRule struct {
	Timezone        string `json:"timezone,omitempty"`
	Start           string `json:"start,omitempty"`
	End             string `json:"end,omitempty"`
	DesiredReplicas int    `json:"desiredReplicas,omitempty"`
}

// CreateVMGroupRequest represents the API request body for creating a virtual machine group.
type CreateVMGroupRequest struct {
	Name              string            `json:"name"`
	SKUID             string            `json:"skuID"`
	VPCID             string            `json:"vpcID"`
	Image             string            `json:"image,omitempty"`
	AutoscalingPolicy AutoscalingPolicy `json:"autoscalers,omitempty"`
	Tags              []Tag             `json:"tags,omitempty"`
	Customization     *Customization    `json:"customization,omitempty"`
}
type vmGroupClient struct {
	apiClient
}

// CreateVMGroup creates a new virtual machine group
func (api *vmGroupClient) CreateVMGroup(ctx context.Context, createRequest CreateVMGroupRequest) (VMGroup, error) {
	err := api.post(ctx, api.namespacedPath("/vm-groups/"), createRequest, nil)
	if err != nil {
		return VMGroup{}, err
	}
	// Fetch the created VM to get full details
	return api.GetVMGroup(ctx, createRequest.Name)
}

// DeleteVMGroup deletes a virtual machine group by name
func (api *vmGroupClient) DeleteVMGroup(ctx context.Context, name string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/vm-groups/%s", name)))
}

// GetVMGroup retrieves a virtual machine group by name (checks if it exists)
func (api *vmGroupClient) GetVMGroup(ctx context.Context, name string) (vmGroup VMGroup, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/vm-groups/%s", name)), &vmGroup)
	return
}

// ListVMGroups retrieves all virtual machine groups
func (api *vmGroupClient) ListVMGroups(ctx context.Context) (vmGroups []VMGroup, err error) {
	err = api.get(ctx, api.namespacedPath("/vm-groups"), &vmGroups)
	return
}

func newVMGroupClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *vmGroupClient {
	return &vmGroupClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
