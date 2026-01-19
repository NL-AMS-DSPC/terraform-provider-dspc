package client

import (
	"context"
	"fmt"
)

// Block represents a block and all of its properties
type Block struct {
	Name         string            `json:"name" example:"my-block"`
	Size         string            `json:"size" example:"10Gi"`
	StorageClass string            `json:"storageClass" example:"standard"`
	AccessMode   string            `json:"accessMode" example:"ReadWriteOnce" enum:"ReadWriteOnce,ReadWriteMany,ReadOnlyMany"`
	VolumeMode   string            `json:"volumeMode" example:"Filesystem" enum:"Filesystem,Block"`
	Status       string            `json:"status" example:"Bound" enum:"Pending,Bound,Lost"`
	Namespace    string            `json:"namespace,omitempty" example:"default"`
	AttachedToVM string            `json:"attachedToVM,omitempty" example:"my-vm"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

// CreateBlockRequest contains parameters used on creation of a block
type CreateBlockRequest struct {
	Name         string `json:"name"`
	Size         string `json:"size"`
	StorageClass string `json:"storageClass"`
}

// CreateBlockResponse contains result form a CreateBlockRequest
type CreateBlockResponse struct {
	Created string `json:"created"`
}

// UpdateBlockRequest contains the request parameters for updating a block
type UpdateBlockRequest struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

// UpdateBlockResponse contains result from the UpdateBlockRequest call
type UpdateBlockResponse struct {
	Name string `json:"name"`
}

// BlockStorageAttachment represents a connection between a block storage volume and a virtual machine.
type BlockStorageAttachment struct {
	BlockName string `json:"blockName"`
	VMName    string `json:"vmName"`
}

// CreateBlockAttachmentResponse represents the API response when creating a block storage attachment.
type CreateBlockAttachmentResponse struct {
	BlockName string `json:"attached"`
	VMName    string `json:"vm"`
}

// ListBlockAttachmentsForVmResponse represents a block storage volume attached to a virtual machine.
type ListBlockAttachmentsForVmResponse struct {
	Name         string            `json:"name" example:"my-block"`
	Size         string            `json:"size" example:"10Gi"`
	StorageClass string            `json:"storageClass" example:"standard"`
	AccessMode   string            `json:"accessMode" example:"ReadWriteOnce" enum:"ReadWriteOnce,ReadWriteMany,ReadOnlyMany"`
	VolumeMode   string            `json:"volumeMode" example:"Filesystem" enum:"Filesystem,Block"`
	Status       string            `json:"status" example:"Bound" enum:"Pending,Bound,Lost"`
	Namespace    string            `json:"namespace,omitempty" example:"default"`
	AttachedToVM string            `json:"attachedToVM,omitempty" example:"my-vm"`
	Labels       map[string]string `json:"labels,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

type blockStorageClient struct {
	apiClient
}

// CreateAttachment creates a new attachment between a block storage volume and a virtual machine.
func (api *blockStorageClient) CreateAttachment(ctx context.Context, blockName, vmName string) (*BlockStorageAttachment, error) {
	path := fmt.Sprintf("/blocks/%s/attach/%s", blockName, vmName)

	var response CreateBlockAttachmentResponse
	err := api.post(ctx, path, nil, &response)
	if err != nil {
		return nil, err
	}

	return &BlockStorageAttachment{
		BlockName: response.BlockName,
		VMName:    response.VMName,
	}, nil
}

// GetAttachment retrieves an attachment between a block storage volume and a virtual machine.
func (api *blockStorageClient) GetAttachment(ctx context.Context, blockName, vmName string) (*BlockStorageAttachment, error) {
	path := fmt.Sprintf("/virtualmachines/%s/blocks", vmName)
	var attachments []ListBlockAttachmentsForVmResponse
	err := api.get(ctx, path, &attachments)
	if err != nil {
		return nil, err
	}
	for _, attachment := range attachments {
		if attachment.Name == blockName {
			return &BlockStorageAttachment{
				BlockName: attachment.Name,
				VMName:    attachment.AttachedToVM,
			}, nil
		}
	}
	return nil, fmt.Errorf("attachment not found for block %s on VM %s", blockName, vmName)
}

// DeleteAttachment deletes an attachment between a block storage volume and a virtual machine.
func (api *blockStorageClient) DeleteAttachment(ctx context.Context, blockName, vmName string) error {
	path := fmt.Sprintf("/blocks/%s/attach/%s", blockName, vmName)
	return api.delete(ctx, path)
}

// ListBlocks retrieves all blocks
func (api *blockStorageClient) ListBlocks(ctx context.Context) (blocks []*Block, err error) {
	err = api.get(ctx, "/blocks", &blocks)
	return
}

// CreateBlock creates a new block
func (api *blockStorageClient) CreateBlock(ctx context.Context, req CreateBlockRequest) (response *CreateBlockResponse, err error) {
	err = api.post(ctx, "/blocks", req, &response)
	return
}

// UpdateBlock updates a block
func (api *blockStorageClient) UpdateBlock(ctx context.Context, req UpdateBlockRequest) (response *UpdateBlockResponse, err error) {
	err = api.put(ctx, fmt.Sprintf("/blocks/%s", req.Name), req, &response)
	return
}

// GetBlock retrieves a block
func (api *blockStorageClient) GetBlock(ctx context.Context, name string) (block *Block, err error) {
	err = api.get(ctx, fmt.Sprintf("/blocks/%s", name), &block)
	return
}

// DeleteBlock deletes a block
func (api *blockStorageClient) DeleteBlock(ctx context.Context, name string) error {
	return api.delete(ctx, fmt.Sprintf("/blocks/%s", name))
}

func newBlockStorageClient(endpoint, namespace, apiKey string, timeoutSeconds int64) *blockStorageClient {
	return &blockStorageClient{
		newApiClient(endpoint, namespace, apiKey, timeoutSeconds),
	}
}
