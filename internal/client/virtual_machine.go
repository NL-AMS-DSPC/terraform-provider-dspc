package client

import (
	"context"
	"fmt"
	"net/http"
)

// SKU represents a VM SKU/size in the DSPC API
type SKU struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AutoscalingConfig represents autoscaling configuration for a VM
type AutoscalingConfig struct {
	MinReplicas                       *int64 `json:"minReplicas,omitempty"`
	MaxReplicas                       *int64 `json:"maxReplicas,omitempty"`
	TargetCPUUtilizationPercentage    *int64 `json:"targetCPUUtilizationPercentage,omitempty"`
	TargetMemoryUtilizationPercentage *int64 `json:"targetMemoryUtilizationPercentage,omitempty"`
	EnableScaleToZero                 *bool  `json:"enableScaleToZero,omitempty"`
	ScaleToZeroAfter                  *int64 `json:"scaleToZeroAfter,omitempty"`
}

// VM represents a virtual machine in the DSPC API
type VM struct {
	Name            string             `json:"name"`
	SKU             SKU                `json:"sku"`
	Status          string             `json:"status"`
	AttachedBlocks  []string           `json:"attachedBlocks,omitempty"`
	ResourceVersion string             `json:"resourceVersion,omitempty"`
	Autoscaling     *AutoscalingConfig `json:"autoscaling,omitempty"`
	Replicas        *int32             `json:"replicas,omitempty"`
}

// CreateVMRequest represents the request body for creating a VM
type CreateVMRequest struct {
	Name        string             `json:"name"`
	SKUID       string             `json:"skuID"`
	Autoscaling *AutoscalingConfig `json:"autoscaling,omitempty"`
}

// CreateVMResponse represents the response from creating a VM
type CreateVMResponse struct {
	Created string `json:"created"`
}

// DeleteVMResponse represents the response from deleting a VM
type DeleteVMResponse struct {
	Deleted string `json:"deleted"`
}

type virtualMachineClient struct {
	apiClient
}

// CreateVM creates a new virtual machine
func (api *virtualMachineClient) CreateVM(ctx context.Context, name, skuID string, autoscaling *AutoscalingConfig) (*VM, error) {
	var response CreateVMResponse
	err := api.post(ctx, "/virtualmachines/", CreateVMRequest{
		Name:        name,
		SKUID:       skuID,
		Autoscaling: autoscaling,
	}, &response)
	if err != nil {
		return nil, err
	}
	// Fetch the created VM to get full details
	return api.GetVM(ctx, response.Created)
}

// DeleteVM deletes a virtual machine by name
func (api *virtualMachineClient) DeleteVM(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/virtualmachines/%s", name))
}

// GetVM retrieves a virtual machine by name (checks if it exists)
func (api *virtualMachineClient) GetVM(ctx context.Context, name string) (vm *VM, err error) {
	err = api.get(ctx, fmt.Sprintf("/virtualmachines/%s", name), &vm)
	return
}

// ListVMs retrieves all virtual machines
func (api *virtualMachineClient) ListVMs(ctx context.Context) (virtualMachines []*VM, err error) {
	err = api.get(ctx, "/virtualmachines", &virtualMachines)
	return
}

func newVirtualMachineClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *virtualMachineClient {
	return &virtualMachineClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
