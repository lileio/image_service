package storage

import "context"

type StorageObject struct {
	Filename string
	URL      string
}

type Storage interface {
	Store(ctx context.Context, data []byte, filename string) (*StorageObject, error)
	Delete(ctx context.Context, filename string) error
}
