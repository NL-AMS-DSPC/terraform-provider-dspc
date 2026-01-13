package client

import (
	"context"
	"fmt"
)

// Block represents a block and all of its properties
type Block struct {
	Name         string            `json:"name" example:"my-pvc"`
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
	Name string `json:"name"`
	Size string `json:"size"`
}

// CreateBlockResponse contains result form a CreateBlockRequest
type CreateBlockResponse struct {
	Created string `json:"created"`
}

type UpdateBlockRequest struct {
	Size string `json:"size"`
}

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
	Name         string            `json:"name" example:"my-pvc"`
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

type blockStorageService struct {
	api requestMaker
}

// CreateAttachment creates a new attachment between a block storage volume and a virtual machine.
func (svc *blockStorageService) CreateAttachment(ctx context.Context, blockName, vmName string) (*BlockStorageAttachment, error) {
	path := fmt.Sprintf("/pvcs/%s/attach/%s", blockName, vmName)

	var response CreateBlockAttachmentResponse
	err := svc.api.Create(ctx, path, nil, &response)
	if err != nil {
		return nil, err
	}

	return &BlockStorageAttachment{
		BlockName: response.BlockName,
		VMName:    response.VMName,
	}, nil
}

// GetAttachment retrieves an attachment between a block storage volume and a virtual machine.
func (svc *blockStorageService) GetAttachment(ctx context.Context, blockName, vmName string) (*BlockStorageAttachment, error) {
	path := fmt.Sprintf("/virtualmachines/%s/pvcs", vmName)
	var attachments []ListBlockAttachmentsForVmResponse
	err := svc.api.Get(ctx, path, &attachments)
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
func (svc *blockStorageService) DeleteAttachment(ctx context.Context, blockName, vmName string) error {
	path := fmt.Sprintf("/pvcs/%s/attach/%s", blockName, vmName)
	return svc.api.Delete(ctx, path)
}

// ListBlocks retrieves all blocks
func (svc *blockStorageService) ListBlocks(ctx context.Context) ([]*Block, error) {
	var blocks []*Block
	err := svc.api.Get(ctx, "/pvcs", &blocks)
	if err != nil {
		return nil, err
	}
	fmt.Println("got all", blocks[0])
	return blocks, nil
}

// CreateBlock creates a new block
func (svc *blockStorageService) CreateBlock(ctx context.Context, req CreateBlockRequest) (*CreateBlockResponse, error) {
	var response CreateBlockResponse
	err := svc.api.Create(ctx, "/pvcs", req, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
func (svc *blockStorageService) UpdateBlock(ctx context.Context, name string, req UpdateBlockRequest) (*UpdateBlockResponse, error) {
	var response UpdateBlockResponse
	err := svc.api.Update(ctx, "/pvcs", req, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

func (svc *blockStorageService) GetBlock(ctx context.Context, name string) (*Block, error) {
	var block Block
	err := svc.api.Get(ctx, fmt.Sprintf("/pvcs/%s", name), &block)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (svc *blockStorageService) DeleteBlock(ctx context.Context, name string) error {
	return svc.api.Delete(ctx, fmt.Sprintf("/pvcs/%s", name))
}

// newBlockStorageService creates a new blockStorageService with the provided request maker.
func newBlockStorageService(client requestMaker) *blockStorageService {
	return &blockStorageService{api: client}
}
