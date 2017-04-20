package workers

import (
	"context"

	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

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

		span := opentracing.SpanFromContext(j.Ctx)
		var wspan opentracing.Span
		if span != nil {
			wspan = opentracing.StartSpan(
				"image_worker: "+versionName(j),
				opentracing.ChildOf(span.Context()),
			)

			j.Ctx = opentracing.ContextWithSpan(j.Ctx, wspan)
		}

		i, err := process(j)
		if err != nil {
			j.ErrChan <- err
			continue
		}

		j.ResponseChan <- *i

		if span != nil {
			wspan.Finish()
		}
		log.Debugf("Finished job on worker: %d", id)
	}
}

func process(j ImageJob) (*image_service.Image, error) {
	if j.Delete {
		err := store.Delete(j.Ctx, j.Filename)
		if err != nil {
			return nil, errors.Wrap(err, "delete failed")
		}

		return &image_service.Image{
			Filename: j.Filename,
		}, nil
	}

	if j.Op != nil {
		data, err := images.Process(j.Data, j.Op)
		if err != nil {
			return nil, errors.Wrap(err, "processing failed")
		}

		j.Data = data
	}

	obj, err := store.Store(j.Ctx, j.Data, j.Filename)
	if err != nil {
		return nil, errors.Wrap(err, "storage failed")
	}

	return &image_service.Image{
		Filename:    obj.Filename,
		Url:         obj.URL,
		VersionName: versionName(j),
	}, nil
}

func versionName(i ImageJob) string {
	if i.Op != nil {
		return i.Op.VersionName
	}
	return "original"
}
