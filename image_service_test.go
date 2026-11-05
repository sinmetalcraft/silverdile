package silverdile_test

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/sinmetalcraft/silverdile/v3"
)

const bucket = "sinmetal-ci-silverdile"
const object = "jun0.jpg"

func TestImageService_Read(t *testing.T) {
	ctx := context.Background()

	is := newImageService(t)
	_, attrs, err := is.Read(ctx, bucket, object)
	if err != nil {
		t.Fatal(err)
	}
	if e, g := "image/jpeg", attrs.ContentType; e != g {
		t.Errorf("want attrs.ContentType %s but got %s", e, g)
	}
}

func TestImageService_ResizeToFitLongSide(t *testing.T) {
	ctx := context.Background()

	is := newImageService(t)
	img, _, err := is.Read(ctx, bucket, object)
	if err != nil {
		t.Fatal(err)
	}

	_, err = is.ResizeToFitLongSide(ctx, img, 256)
	if err != nil {
		t.Fatal(err)
	}
}

func TestImageService_Write(t *testing.T) {
	ctx := context.Background()

	// silverdile.WithAlterBucket("alter-sinmetal-ci-silverdile2")
	is := newImageService(t)
	img, attrs, err := is.Read(ctx, bucket, object)
	if err != nil {
		t.Fatal(err)
	}

	size := 10
	newImg, err := is.ResizeToFitLongSide(ctx, img, size)
	if err != nil {
		t.Fatal(err)
	}
	if err := is.Write(ctx, fmt.Sprintf("alter-%s", bucket), fmt.Sprintf("%s_s%d", object, size), newImg, &storage.ObjectAttrs{
		ContentType: attrs.ContentType,
	}); err != nil {
		t.Fatal(err)
	}
}

func TestContentTypeToImagingFormat(t *testing.T) {
	f, err := silverdile.ContentTypeToImagingFormat("image/jpeg")
	if err != nil {
		t.Fatal(err)
	}
	if e, g := imaging.JPEG, f; e != g {
		t.Errorf("want %d but got %d", e, g)
	}
}

func newImageService(t *testing.T) *silverdile.ImageService {
	ctx := context.Background()

	gcs, err := storage.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := gcs.Close(); err != nil {
			// noop
		}
	})

	is, err := silverdile.NewImageService(ctx, gcs)
	if err != nil {
		t.Fatal(err)
	}

	return is
}
