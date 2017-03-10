package server

import (
	"fmt"
	"strings"

	"golang.org/x/net/context"

	"github.com/pkg/errors"

	"github.com/lileio/image_service"
	"github.com/lileio/image_service/workers"
	uuid "github.com/satori/go.uuid"
)

type Server struct {
	image_service.ImageServiceServer
}

func (s Server) Store(
	req *image_service.ImageStoreRequest,
	stream image_service.ImageService_StoreServer) error {

	_, err := store(stream.Context(), req, stream)
	return err
}

func (s Server) StoreSync(
	ctx context.Context,
	req *image_service.ImageStoreRequest,
) (*image_service.ImageSyncResponse, error) {

	images, err := store(ctx, req, nil)
	return &image_service.ImageSyncResponse{
		Images: images,
	}, err
}

func (s Server) Delete(
	ctx context.Context,
	req *image_service.DeleteRequest,
) (*image_service.DeleteResponse, error) {

	images := make(chan image_service.Image)
	errch := make(chan error)
	defer close(images)
	defer close(errch)

	go func() {
		workers.Queue <- workers.ImageJob{
			Filename:     req.Filename,
			ResponseChan: images,
			ErrChan:      errch,
			Ctx:          ctx,
			Delete:       true,
		}
	}()

	select {
	case img := <-images:
		return &image_service.DeleteResponse{Filename: img.Filename}, nil
	case err := <-errch:
		return nil, err
	}
}

func store(
	ctx context.Context,
	req *image_service.ImageStoreRequest,
	stream image_service.ImageService_StoreServer,
) ([]*image_service.Image, error) {

	expectedImages := len(req.Ops) + 1
	images := make(chan image_service.Image, expectedImages)
	errch := make(chan error, expectedImages)
	defer close(images)
	defer close(errch)

	id := uuid.NewV1().String()

	go func() {
		// Upload the original image
		workers.Queue <- workers.ImageJob{
			Filename:     id + "-" + req.Filename,
			Data:         req.Data,
			ResponseChan: images,
			ErrChan:      errch,
			Ctx:          ctx,
		}

		// And then the derivatives
		for _, op := range req.Ops {
			workers.Queue <- workers.ImageJob{
				Filename:     filenameWithOpts(id, op),
				Data:         req.Data,
				Op:           op,
				ResponseChan: images,
				ErrChan:      errch,
				Ctx:          ctx,
			}
		}
	}()

	errs := []error{}
	imgs := []*image_service.Image{}
	for i := 0; i < expectedImages; i++ {
		select {
		case img := <-images:
			imgs = append(imgs, &img)
			if stream != nil {
				stream.Send(&img)
			}
		case err := <-errch:
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		err := errors.New("error storing images: ")
		for _, e := range errs {
			err = errors.Wrap(err, e.Error())
		}

		return imgs, err
	}

	return imgs, nil
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
