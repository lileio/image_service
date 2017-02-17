package server

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/lileio/image_service/image_service"
	"github.com/lileio/image_service/workers"
	uuid "github.com/satori/go.uuid"
)

type Server struct {
	image_service.ImageServiceServer
}

func (s Server) Store(
	req *image_service.ImageStoreRequest,
	stream image_service.ImageService_StoreServer) error {

	expectedImages := len(req.Ops) + 1
	images := make(chan image_service.Image, expectedImages)
	errch := make(chan error, expectedImages)

	id := uuid.NewV1().String()

	go func() {
		// Upload the original image
		workers.Queue <- workers.ImageJob{
			Filename:     id + "-" + req.Filename,
			Data:         req.Data,
			ResponseChan: images,
			ErrChan:      errch,
			Ctx:          stream.Context(),
		}

		// And then the derivatives
		for _, op := range req.Ops {
			workers.Queue <- workers.ImageJob{
				Filename:     filenameWithOpts(id, op),
				Data:         req.Data,
				Op:           op,
				ResponseChan: images,
				ErrChan:      errch,
				Ctx:          stream.Context(),
			}
		}
	}()

	errs := []error{}
	for i := 0; i < expectedImages; i++ {
		select {
		case img := <-images:
			stream.Send(&img)
		case err := <-errch:
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		err := errors.New("error storing images")
		for _, e := range errs {
			err = errors.Wrap(err, e.Error())
		}

		return err
	}

	return nil
}

// filenameWithOpts creates a unique filename for given an a unique id,
// with ops for filetype and version name
func filenameWithOpts(id string, ops *image_service.ImageOperation) string {
	return fmt.Sprintf("%s-%s.%s",
		id,
		strings.ToLower(ops.VersionName),
		strings.ToLower(ops.Format.String()),
	)
}
