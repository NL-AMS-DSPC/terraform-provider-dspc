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

type CreateBlockAttachmentResponse struct {
	BlockName string `json:"attached"`
	VMName    string `json:"vm"`
}
