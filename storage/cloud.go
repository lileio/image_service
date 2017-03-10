package storage

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/grpc-opentracing/go/otgrpc"
	"github.com/lileio/cloud_storage_service/cloud_storage_service"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type CloudStorage struct {
	Storage

	addr   string
	client cloud_storage_service.CloudStorageServiceClient
}

func NewCloudStorage(addr string) *CloudStorage {
	return &CloudStorage{addr: addr}
}

func (cs *CloudStorage) connect() error {
	if cs.client != nil {
		return nil
	}

	t := opentracing.GlobalTracer()

	conn, err := grpc.Dial(
		cs.addr,
		grpc.WithInsecure(),
		grpc.WithTimeout(1*time.Second),
		grpc.WithUnaryInterceptor(otgrpc.OpenTracingClientInterceptor(t)),
	)
	if err != nil {
		return err
	}

	cs.client = cloud_storage_service.NewCloudStorageServiceClient(conn)
	return nil
}

func (s *CloudStorage) Store(ctx context.Context, data []byte, filename string) (*StorageObject, error) {
	err := s.connect()
	if err != nil {
		return nil, err
	}

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

func (s *CloudStorage) Delete(ctx context.Context, filename string) error {
	err := s.connect()
	if err != nil {
		return err
	}

	_, err = s.client.Delete(ctx, &cloud_storage_service.DeleteRequest{
		Filename: filename,
	})

	if err != nil {
		return err
	}

	return nil
}
