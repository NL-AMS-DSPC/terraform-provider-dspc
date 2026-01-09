package client

import "github.com/NL-AMS-DSPC/terraform-provider-dspc/internal/provider"

type VirtualMachineService struct {
}

func (vms *VirtualMachineService) Get(name string) (*provider.VM, error) {
	return nil, nil
}
