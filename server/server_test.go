package server

import (
	"context"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/lileio/image_service/image_service"
	"github.com/lileio/image_service/storage"
	"github.com/lileio/image_service/workers"
	"github.com/lileio/lile"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

var client image_service.ImageServiceClient

func init() {
	log.SetLevel(log.DebugLevel)

	if os.Getenv("FILE_LOCATION") == "" {
		panic("FILE_LOCATION must be set")
	}

	cs, err := storage.NewFileStorage(os.Getenv("FILE_LOCATION"))
	if err != nil {
		panic(err)
	}

	workers.StartWorkerPool(5, cs)

	var s = Server{}
	impl := func(g *grpc.Server) {
		image_service.RegisterImageServiceServer(g, s)
	}

	serv := lile.NewServer(
		lile.Port(":9999"),
		lile.Implementation(impl),
	)

	go serv.ListenAndServe()
	conn := dialWithRetry()
	client = image_service.NewImageServiceClient(conn)
}

func dialWithRetry() *grpc.ClientConn {
	conn, err := grpc.Dial("localhost:9999", grpc.WithInsecure())
	if err != nil {
		log.Infof("failed to dial: %v. Retrying..", err)
		time.Sleep(1)
	}
	return conn
}

func TestStore(t *testing.T) {
	b, err := ioutil.ReadFile("../test/pic.jpg")
	assert.Nil(t, err)

	ctx := context.Background()
	req := &image_service.ImageStoreRequest{
		Filename: "pic.jpg",
		Data:     b,
		Ops: []*image_service.ImageOperation{
			&image_service.ImageOperation{Crop: true, Width: 100, Height: 200, VersionName: "1"},
			&image_service.ImageOperation{Crop: true, Width: 200, Height: 300, VersionName: "2"},
			&image_service.ImageOperation{Crop: true, Width: 300, Height: 400, VersionName: "3"},
		},
	}

	stream, err := client.Store(ctx, req)
	assert.Nil(t, err)

	images := []*image_service.Image{}

	for {
		img, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			assert.Fail(t, err.Error())
		}

		images = append(images, img)
	}

	assert.Equal(t, len(req.Ops)+1, len(images))

	for _, img := range images {
		_, err := client.Delete(ctx, &image_service.DeleteRequest{
			Filename: img.Filename,
		})

		if err != nil {
			assert.Fail(t, err.Error())
		}
	}
}
