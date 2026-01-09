package client

type BlockStorageService struct {
	api requestMaker
}

func NewBlockStorageService(client requestMaker) *BlockStorageService {
	return &BlockStorageService{api: client}
}
