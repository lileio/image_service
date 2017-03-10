package storage

import (
	"context"
	"os"
)

type StorageObject struct {
	Filename string
	URL      string
}

type Storage interface {
	Store(ctx context.Context, data []byte, filename string) (*StorageObject, error)
	Delete(ctx context.Context, filename string) error
}

func StorageFromEnv() Storage {
	var store Storage

	if os.Getenv("CLOUD_STORAGE_ADDR") != "" {
		store = NewCloudStorage(os.Getenv("CLOUD_STORAGE_ADDR"))
	}

	if os.Getenv("FILE_LOCATION") != "" {
		s, err := NewFileStorage(os.Getenv("FILE_LOCATION"))
		if err != nil {
			panic(err)
		}

		store = s
	}

	return store
}
