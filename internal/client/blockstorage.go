package client

type BlockStorageService struct {
	api blockStorageConsumer
}

func NewBlockStorageService(consumer blockStorageConsumer) *BlockStorageService {
	return &BlockStorageService{api: consumer}
}

type blockStorageConsumer interface{}
