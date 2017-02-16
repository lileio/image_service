package images

import (
	"github.com/h2non/bimg"
	"github.com/lileio/image_service/image_service"
	"github.com/pkg/errors"
)

func Process(data []byte, op *image_service.ImageOperation) ([]byte, error) {
	o := optionsFromOperation(op)

	data, err := bimg.NewImage(data).Process(o)
	if err != nil {
		return []byte{}, errors.Wrap(err, "image processing failed")
	}

	return data, nil
}

func optionsFromOperation(op *image_service.ImageOperation) bimg.Options {
	var itype bimg.ImageType
	switch op.Format {
	case image_service.Format_JPEG:
		itype = bimg.JPEG
	case image_service.Format_PNG:
		itype = bimg.PNG
	case image_service.Format_WEBP:
		itype = bimg.WEBP
	}

	return bimg.Options{
		Height:      int(op.Height),
		Width:       int(op.Width),
		Quality:     int(op.Quality),
		Compression: int(op.Compression),
		Crop:        op.Crop,
		Enlarge:     op.Enlarge,
		Flip:        op.Flip,
		Interlace:   op.Interlace,
		Type:        itype,
	}
}
