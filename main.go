package main

import (
	"os"
	"strconv"

	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
	"github.com/lileio/image_service/image_service"
	"github.com/lileio/image_service/server"
	"github.com/lileio/image_service/storage"
	"github.com/lileio/image_service/workers"
	"github.com/lileio/lile"
)

func main() {
	store := storageFromEnv()

	if os.Getenv("DEBUG") == "true" {
		log.SetLevel(log.DebugLevel)
	}

	poolSize := 5
	if os.Getenv("WORKER_POOL_SIZE") != "" {
		i, err := strconv.Atoi(os.Getenv("WORKER_POOL_SIZE"))
		if err != nil {
			panic(err)
		}

		poolSize = i
	}

	workers.StartWorkerPool(poolSize, store)

	s := &server.Server{}
	impl := func(g *grpc.Server) {
		image_service.RegisterImageServiceServer(g, s)
	}

	err := lile.NewServer(
		lile.Name("image_service"),
		lile.Implementation(impl),
	).ListenAndServe()

	log.Fatal(err)
}

func storageFromEnv() storage.Storage {
	var store storage.Storage
	if os.Getenv("CLOUD_STORAGE_ADDR") != "" {
		s, err := storage.NewCloudStorage(os.Getenv("CLOUD_STORAGE_ADDR"))
		if err != nil {
			panic(err)
		}

		store = s
	}

	if os.Getenv("FILE_LOCATION") != "" {
		s, err := storage.NewFileStorage(os.Getenv("FILE_LOCATION"))
		if err != nil {
			panic(err)
		}

		store = s
	}

	return store
}
