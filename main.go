package main

import (
	"os"

	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
	"github.com/lileio/image_service/image_service"
	"github.com/lileio/image_service/server"
	"github.com/lileio/image_service/storage"
	"github.com/lileio/image_service/workers"
	"github.com/lileio/lile"
)

func main() {
	var store storage.Storage
	if os.Getenv("CLOUD_STORAGE_ADDR") != "" {
		s, err := storage.NewCloudStorage("localhost:8000")
		if err != nil {
			panic(err)
		}

		store = s
	}

	workers.StartWorkerPool(5, store)
	s := &server.Server{}

	impl := func(g *grpc.Server) {
		image_service.RegisterImageServiceServer(g, s)
	}

	workers.StartWorkerPool(5, nil)

	err := lile.NewServer(
		lile.Name("image_service"),
		lile.Implementation(impl),
	).ListenAndServe()

	log.Fatal(err)
}
