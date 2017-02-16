package storage

import (
	"context"

	"github.com/lileio/cloud_storage_service/cloud_storage_service"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type CloudStorage struct {
	Storage

	client cloud_storage_service.CloudStorageServiceClient
}

func NewCloudStorage(addr string) (*CloudStorage, error) {
	cs := &CloudStorage{}

	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	cs.client = cloud_storage_service.NewCloudStorageServiceClient(conn)
	return cs, nil
}

func (s *CloudStorage) Store(ctx context.Context, data []byte, filename string) (*StorageObject, error) {
	obj, err := s.client.Store(ctx, &cloud_storage_service.StoreRequest{
		Filename: filename,
		Data:     data,
	})

	if err != nil {
		return nil, errors.Wrap(err, "cloud storage failed")
	}

	return &StorageObject{
		Filename: filename,
		URL:      obj.Url,
	}, nil
}
