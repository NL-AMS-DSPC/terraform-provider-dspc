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

type virtualMachineService struct {
	api requestMaker
}

// CreateVM creates a new virtual machine
func (svc *virtualMachineService) CreateVM(ctx context.Context, name string) (*VM, error) {
	vm := VM{Name: name}
	var response CreateVMResponse
	err := svc.api.Create(ctx, "/virtualmachines", vm, &response)
	if err != nil {
		return nil, err
	}
	// TODO: this is not correct according to api def? it just returns the vm
	return &VM{Name: response.Created}, nil
}

// DeleteVM deletes a virtual machine by name
func (svc *virtualMachineService) DeleteVM(ctx context.Context, name string) error {
	return svc.api.Delete(ctx, fmt.Sprintf("/virtualmachines/%s", name))
}

// GetVM retrieves a virtual machine by name (checks if it exists)
func (svc *virtualMachineService) GetVM(ctx context.Context, name string) (*VM, error) {
	// TODO: why not use the get endpoint?
	vms, err := svc.ListVMs(ctx)
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if vm.Name == name {
			return vm, nil
		}
	}

	return nil, fmt.Errorf("VM '%s' not found. Please verify the VM name exists or check your API endpoint", name)
}

// ListVMs retrieves all virtual machines
func (svc *virtualMachineService) ListVMs(ctx context.Context) ([]*VM, error) {
	var virtualMachines []*VM
	err := svc.api.Get(ctx, "/virtualmachines", &virtualMachines)
	if err != nil {
		return nil, err
	}
	return virtualMachines, nil
}

func NewVirtualMachineService(client requestMaker) *virtualMachineService {
	return &virtualMachineService{
		api: client,
	}
}
