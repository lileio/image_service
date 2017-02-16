package storage

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileStorage struct {
	Storage
	root string
}

func NewFileStorage(rootDir string) (*FileStorage, error) {
	path, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	s, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if !s.Mode().IsDir() {
		return nil, errors.New("root dir for file storage must be a dir")
	}

	return &FileStorage{root: rootDir}, nil
}

func (s *FileStorage) Store(ctx context.Context, data []byte, filename string) (*StorageObject, error) {
	url := s.root + "/" + filename
	err := ioutil.WriteFile(url, data, os.ModePerm)
	if err != nil {
		return nil, err
	}

	return &StorageObject{
		Filename: filename,
		URL:      url,
	}, nil
}
