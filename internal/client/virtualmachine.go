package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	api requestMaker
}

func NewVirtualMachineService(client requestMaker) *VirtualMachineService {
	return &VirtualMachineService{
		api: client,
	}
}

// CreateVM creates a new virtual machine
func (svc *VirtualMachineService) CreateVM(ctx context.Context, name string) (*VM, error) {
	vm := VM{Name: name}
	resp, err := svc.api.MakeRequest(ctx, http.MethodPost, "/virtualmachine", vm)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var createResp CreateVMResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &VM{Name: createResp.Created}, nil
}

// DeleteVM deletes a virtual machine by name
func (svc *VirtualMachineService) DeleteVM(ctx context.Context, name string) error {
	vm := VM{Name: name}
	resp, err := svc.api.MakeRequest(ctx, http.MethodDelete, "/virtualmachine", vm)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetVM retrieves a virtual machine by name (checks if it exists)
func (svc *VirtualMachineService) GetVM(ctx context.Context, name string) (*VM, error) {
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
	resp, err := svc.api.MakeRequest(ctx, http.MethodGet, "/virtualmachine", nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Warning: failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("API error %d: failed to read response body: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var vms []*VM
	if err := json.NewDecoder(resp.Body).Decode(&vms); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return vms, nil
}
