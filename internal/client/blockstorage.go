package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type CreateBlockRequest struct {
	Name string
	Size string
}

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

type BlockStorageAttachment struct {
	BlockName string
	VMName    string
}

type CreateBlockResponse struct {
	Created string `json:"created"`
}

type CreateBlockAttachmentResponse struct {
	BlockName string `json:"attached"`
	VMName    string `json:"vm"`
}

type BlockStorageService struct {
	api requestMaker[Block]
}

func NewBlockStorageService(client requestMaker[Block]) *BlockStorageService {
	return &BlockStorageService{api: client}
}

func (c *BlockStorageService) CreateAttachment(ctx context.Context, blockName, vmName string) (*BlockStorageAttachment, error) {
	path := fmt.Sprintf("/pvcs/%s/attach/%s", blockName, vmName)
	resp, err := c.api.MakeRequest(ctx, http.MethodPost, path, nil)
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

	var respBody CreateBlockAttachmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &BlockStorageAttachment{
		BlockName: blockName,
		VMName:    vmName,
	}, nil
}

func (svc *BlockStorageService) ListBlocks(ctx context.Context) ([]*Block, error) {
	return svc.api.List(ctx, "/pvcs")
}

// CreateBlock creates a new block
func (svc *BlockStorageService) CreateBlock(ctx context.Context, req CreateBlockRequest) (*CreateBlockResponse, error) {
	resp, err := svc.api.MakeRequest(ctx, http.MethodPost, "/pvcs", req)
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

	var createResponse CreateBlockResponse
	if err := json.NewDecoder(resp.Body).Decode(&createResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &createResponse, nil
}

func (svc *BlockStorageService) GetBlock(ctx context.Context, name string) (*Block, error) {
	return svc.api.Get(ctx, fmt.Sprintf("/pvcs/%s", name))
}

func (svc *BlockStorageService) DeleteBlock(ctx context.Context, name string) error {
	return svc.api.Delete(ctx, fmt.Sprintf("/pvcs/%s", name))
}
