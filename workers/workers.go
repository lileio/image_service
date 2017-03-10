package workers

import (
	"context"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/lileio/image_service"
	"github.com/lileio/image_service/images"
	"github.com/lileio/image_service/storage"
)

var (
	Queue = make(chan ImageJob)
	store storage.Storage
)

type ImageJob struct {
	ResponseChan chan image_service.Image
	ErrChan      chan error
	Ctx          context.Context

	Filename string

	// If doing a resize
	Data []byte
	Op   *image_service.ImageOperation

	// If doing a delete
	Delete bool
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

		if j.Delete {
			err := store.Delete(j.Ctx, j.Filename)
			if err != nil {
				j.ErrChan <- errors.Wrap(err, "delete failed")
				continue
			}

			j.ResponseChan <- image_service.Image{Filename: j.Filename}
			continue
		}

		if j.Op != nil {
			data, err := images.Process(j.Data, j.Op)
			if err != nil {
				j.ErrChan <- err
				continue
			}

			j.Data = data
		}

		obj, err := store.Store(j.Ctx, j.Data, j.Filename)
		if err != nil {
			j.ErrChan <- errors.Wrap(err, "storage failed")
			continue
		}

		j.ResponseChan <- image_service.Image{
			Filename:    obj.Filename,
			Url:         obj.URL,
			VersionName: versionName(j),
		}

		log.Debugf("Finished job on worker: %d", id)
	}
}

func versionName(i ImageJob) string {
	if i.Op != nil {
		return i.Op.VersionName
	}
	return "original"
}
