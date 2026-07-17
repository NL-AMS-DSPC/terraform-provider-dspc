package client

import (
	"context"
	"fmt"
	"net/http"
)

// VM represents a virtual machine in the DSPC API
type VM struct {
	URN             string      `json:"urn"`
	Name            string      `json:"name"`
	SKU             SKUResponse `json:"sku"`
	Status          string      `json:"status"`
	LastError       string      `json:"lastError"`
	Tags            []Tag       `json:"tags"`
	AttachedVolumes []string    `json:"attachedVolumes"`
	OS              OSDetails   `json:"os"`
}

// SKUResponse represents basic SKU information in API responses.
type SKUResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Family      string `json:"family"`
	Threads     uint64 `json:"threads"`
	Cores       uint64 `json:"cores"`
	MemoryInMB  uint64 `json:"memoryInMB"`
	StorageInGB uint64 `json:"storageInGB"`
	StorageType string `json:"storageType"`
	GPUCount    uint64 `json:"GPUCount"`
	GPUType     string `json:"GPUType"`
}

// OSDetails represents details about the VM OS
type OSDetails struct {
	ID           string `json:"id"`
	Family       string `json:"family"`
	Distribution string `json:"distribution"`
	Release      string `json:"release"`
	DisplayName  string `json:"displayName"`
}

// CreateVMRequest represents the request body for creating a VM
type CreateVMRequest struct {
	Name          string         `json:"name"`
	SKUID         string         `json:"skuID"`
	VPCID         string         `json:"vpcID"`
	Image         string         `json:"image,omitempty"`
	Tags          []Tag          `json:"tags,omitempty"`
	Customization *Customization `json:"customization,omitempty"`
}

// Customization contains optional VM initialization data.
type Customization struct {
	CloudInit *CloudInitCustomization `json:"cloudInit,omitempty"`
	Ignition  *IgnitionCustomization  `json:"ignition,omitempty"`
}

// CloudInitCustomization contains optional cloud-init NoCloud input.
type CloudInitCustomization struct {
	UserData string `json:"userData,omitempty"`
	MetaData string `json:"metaData,omitempty"`
}

// IgnitionCustomization contains Ignition or Butane configuration input.
type IgnitionCustomization struct {
	Format string `json:"format"`
	Config string `json:"config"`
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
func (api *virtualMachineClient) CreateVM(ctx context.Context, createRequest CreateVMRequest) (VM, error) {
	var response CreateVMResponse
	err := api.post(ctx, api.namespacedPath("/vms/"), createRequest, &response)
	if err != nil {
		return VM{}, err
	}
	// Fetch the created VM to get full details
	return api.GetVM(ctx, createRequest.Name)
}

// DeleteVM deletes a virtual machine by name
func (api *virtualMachineClient) DeleteVM(ctx context.Context, name string) error {
	return api.delete(ctx, api.namespacedPath(fmt.Sprintf("/vms/%s", name)))
}

// GetVM retrieves a virtual machine by name (checks if it exists)
func (api *virtualMachineClient) GetVM(ctx context.Context, name string) (vm VM, err error) {
	err = api.get(ctx, api.namespacedPath(fmt.Sprintf("/vms/%s", name)), &vm)
	return
}

// ListVMs retrieves all virtual machines
func (api *virtualMachineClient) ListVMs(ctx context.Context) (virtualMachines []VM, err error) {
	err = api.get(ctx, api.namespacedPath("/vms"), &virtualMachines)
	return
}

func newVirtualMachineClient(endpoint, namespace, pathPrefix string, authMgr *authManager, httpClient *http.Client) *virtualMachineClient {
	return &virtualMachineClient{
		newAPIClient(endpoint, namespace, pathPrefix, authMgr, httpClient),
	}
}
