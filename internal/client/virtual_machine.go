package client

import (
	"context"
	"fmt"
)

// VM represents a virtual machine in the DSPC API
type VM struct {
	Name string `json:"VMName"`
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
func (api *virtualMachineClient) CreateVM(ctx context.Context, name string) (*VM, error) {
	var response CreateVMResponse
	err := api.post(ctx, "/virtualmachines", VM{Name: name}, &response)
	if err != nil {
		return nil, err
	}
	return &VM{Name: response.Created}, nil
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

func newVirtualMachineClient(endpoint, namespace, apiKey string, timeoutSeconds int64) *virtualMachineClient {
	return &virtualMachineClient{
		newApiClient(endpoint, namespace, apiKey, timeoutSeconds),
	}
}
