package workers

import (
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/lileio/image_service/image_service"
	"github.com/lileio/image_service/images"
	"github.com/lileio/image_service/storage"
)

var (
	Queue = make(chan ImageJob)
	store storage.Storage
)

type ImageJob struct {
	Filename     string
	Data         []byte
	Op           *image_service.ImageOperation
	ResponseChan chan image_service.Image
	ErrChan      chan error
	Ctx          context.Context
}

func StartWorkerPool(workerCount int, s storage.Storage) {
	if s == nil {
		log.Fatal("storage for workers must be set")
	}

	store = s

	log.Printf("Starting worker pool with count: %d", workerCount)
	for w := 1; w <= workerCount; w++ {
		go worker(w, Queue)
	}
}

func worker(id int, jobs <-chan ImageJob) {
	for j := range jobs {
		log.Debugf("Starting job on worker: %d", id)

		if j.Op != nil {
			data, err := images.Process(j.Data, j.Op)
			if err != nil {
				j.ErrChan <- err
				return
			}

			j.Data = data
		}

		obj, err := store.Store(j.Ctx, j.Data, j.Filename)
		if err != nil {
			j.ErrChan <- errors.Wrap(err, "storage failed")
			return
		}

		j.ResponseChan <- image_service.Image{Filename: obj.Filename, Url: obj.URL}
	}
}
