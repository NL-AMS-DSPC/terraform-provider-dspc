package client

import (
	"context"
)

type BlockStorageService struct {
	api requestMaker
}

func NewBlockStorageService(client requestMaker) *BlockStorageService {
	return &BlockStorageService{api: client}
}

func (c *BlockStorageService) CreateAttachment(ctx context.Context, pvcName, vmName string) (*CreateBlockAttachmentResponse, error) {
	//TODO implement me
	panic("implement me")
}

// todo: add json things
type CreateBlockAttachmentResponse struct {
	PvcName string
	VMName  string
}
