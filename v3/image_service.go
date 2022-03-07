package silverdile

import (
	"context"
	"fmt"
	"image"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/sinmetalcraft/silverdile/v3/internal/trace"
)

type ImageService struct {
	alterBucket string
	GCS         *storage.Client
}

func NewImageService(ctx context.Context, gcs *storage.Client) (*ImageService, error) {
	s := &ImageService{
		GCS: gcs,
	}

	return s, nil
}

// ResizeToFitLongSide is 長辺が指定したサイズになるようにアスペクト比を維持してリサイズする
func (s *ImageService) ResizeToFitLongSide(ctx context.Context, src image.Image, size int) (dst image.Image, err error) {
	ctx = trace.StartSpan(ctx, "ImageService.ResizeToFitLongSide")
	defer func() { trace.EndSpan(ctx, err) }()

	if size < 1 {
		return nil, fmt.Errorf("invalid size")
	}

	// 長辺に合わせてResizeする
	if src.Bounds().Size().X > src.Bounds().Size().Y {
		return imaging.Resize(src, size, 0, imaging.Lanczos), nil
	}
	return imaging.Resize(src, 0, size, imaging.Lanczos), nil
}

func (s *ImageService) Write(ctx context.Context, bucket string, object string, img image.Image, meta *ImageMeta, attrs *storage.ObjectAttrs) (err error) {
	ctx = trace.StartSpan(ctx, "ImageService.Write")
	defer func() { trace.EndSpan(ctx, err) }()

	w := s.GCS.Bucket(bucket).Object(object).NewWriter(ctx)
	if attrs != nil {
		w.ObjectAttrs = *attrs
	}
	w.ObjectAttrs.Name = object
	w.ObjectAttrs.ContentType = meta.ContentType
	if err := imaging.Encode(w, img, meta.FormatType.ImagingFormat()); err != nil {
		return err
	}
	return nil
}

func (s *ImageService) Read(ctx context.Context, bucket string, object string) (img image.Image, meta *ImageMeta, err error) {
	ctx = trace.StartSpan(ctx, "ImageService.Read")
	defer func() { trace.EndSpan(ctx, err) }()

	attrs, err := s.GCS.Bucket(bucket).Object(object).Attrs(ctx)
	if err != nil {
		return nil, nil, err
	}

	r, err := s.GCS.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := r.Close(); err != nil {
			fmt.Printf("failed GCS Reader Close(). err=%+v\n", err)
		}
	}()
	dst, err := imaging.Decode(r)
	if err != nil {
		return nil, nil, err
	}

	var ft FormatType
	switch attrs.ContentType {
	case "image/png":
		ft = PNG
	case "image/jpeg":
		ft = JPEG
	}

	return dst, &ImageMeta{
		ContentType: attrs.ContentType,
		FormatType:  ft,
	}, nil
}
