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

type VirtualMachineService struct {
	api requestMaker[VM]
}

func NewVirtualMachineService(client requestMaker[VM]) *VirtualMachineService {
	return &VirtualMachineService{
		api: client,
	}
}

// CreateVM creates a new virtual machine
func (svc *VirtualMachineService) CreateVM(ctx context.Context, name string) (*VM, error) {
	vm := VM{Name: name}
	return svc.api.Create(ctx, "/virtualmachine", vm)
}

// DeleteVM deletes a virtual machine by name
func (svc *VirtualMachineService) DeleteVM(ctx context.Context, name string) error {
	return svc.api.Delete(ctx, fmt.Sprintf("/virtualmachine/%s", name))
}

// GetVM retrieves a virtual machine by name (checks if it exists)
func (svc *VirtualMachineService) GetVM(ctx context.Context, name string) (*VM, error) {
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
func (svc *VirtualMachineService) ListVMs(ctx context.Context) ([]*VM, error) {
	return svc.api.List(ctx, "/virtualmachine")
}
