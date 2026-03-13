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

// CPUScaler represents CPU-based horizontal pod autoscaling configuration
type CPUScaler struct {
	TargetUtilizationPercentage *int64 `json:"targetUtilizationPercentage,omitempty"`
}

// MemoryScaler represents memory-based horizontal pod autoscaling configuration
type MemoryScaler struct {
	TargetUtilizationPercentage *int64 `json:"targetUtilizationPercentage,omitempty"`
}

// CronScaler represents cron-based scheduling configuration for scaling
type CronScaler struct {
	Timezone        string `json:"timezone,omitempty"`
	Start           string `json:"start,omitempty"`
	End             string `json:"end,omitempty"`
	DesiredReplicas *int32 `json:"desiredReplicas,omitempty"`
}

// ScaleToZeroScaler represents scale-to-zero configuration (KEDA-based)
type ScaleToZeroScaler struct {
	Enabled           *bool  `json:"enabled,omitempty"`
	IdleReplicaCount  *int32 `json:"idleReplicaCount,omitempty"`
	CooldownPeriodSec *int32 `json:"cooldownPeriodSec,omitempty"`
}

// Scalers contains all available scaler configurations
type Scalers struct {
	CPU         *CPUScaler         `json:"cpu,omitempty"`
	Memory      *MemoryScaler      `json:"memory,omitempty"`
	Cron        *CronScaler        `json:"cron,omitempty"`
	ScaleToZero *ScaleToZeroScaler `json:"scaleToZero,omitempty"`
}

// HasCPUScaler returns true if CPU scaler is configured
func (s *Scalers) HasCPUScaler() bool {
	return s != nil && s.CPU != nil
}

// HasMemoryScaler returns true if Memory scaler is configured
func (s *Scalers) HasMemoryScaler() bool {
	return s != nil && s.Memory != nil
}

// HasCronScaler returns true if Cron scaler is configured
func (s *Scalers) HasCronScaler() bool {
	return s != nil && s.Cron != nil
}

// HasScaleToZeroScaler returns true if ScaleToZero scaler is configured
func (s *Scalers) HasScaleToZeroScaler() bool {
	return s != nil && s.ScaleToZero != nil
}

// AutoscalingConfig represents autoscaling configuration for a VM
type AutoscalingConfig struct {
	MinReplicas *int64   `json:"minReplicas,omitempty"`
	MaxReplicas *int64   `json:"maxReplicas,omitempty"`
	Scalers     *Scalers `json:"scalers,omitempty"`
}

// HasScalers returns true if any scalers are configured
func (a *AutoscalingConfig) HasScalers() bool {
	return a != nil && a.Scalers != nil
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
	err := api.post(ctx, api.namespacedPath("/virtualmachines/"), CreateVMRequest{
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
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/virtualmachines/%s", name)))
}

// GetVM retrieves a virtual machine by name (checks if it exists)
func (api *virtualMachineClient) GetVM(ctx context.Context, name string) (vm *VM, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/virtualmachines/%s", name)), &vm)
	return
}

// ListVMs retrieves all virtual machines
func (api *virtualMachineClient) ListVMs(ctx context.Context) (virtualMachines []*VM, err error) {
	err = api.get(ctx, api.namespacedPath("/virtualmachines"), &virtualMachines)
	return
}

func newVirtualMachineClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *virtualMachineClient {
	return &virtualMachineClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
