package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type BlockStorageAttachment struct {
	BlockName string
	VMName    string
}

type BlockStorageService struct {
	api requestMaker
}

func NewBlockStorageService(client requestMaker) *BlockStorageService {
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

func (c *BlockStorageService) GetAttachment(ctx context.Context, blockName, vmName string) (*BlockStorageAttachment, error) {
	path := fmt.Sprintf("/virtualmachines/%s/pvcs", vmName)
	resp, err := c.api.MakeRequest(ctx, http.MethodGet, path, nil)
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

	var attachments []ListBlockAttachmentsForVmResponse
	if err := json.NewDecoder(resp.Body).Decode(&attachments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
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

func (c *BlockStorageService) DeleteAttachment(ctx context.Context, blockName, vmName string) error {
	path := fmt.Sprintf("/pvcs/%s/attach/%s", blockName, vmName)
	resp, err := c.api.MakeRequest(ctx, http.MethodDelete, path, nil)
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

type CreateBlockAttachmentResponse struct {
	BlockName string `json:"attached"`
	VMName    string `json:"vm"`
}

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
