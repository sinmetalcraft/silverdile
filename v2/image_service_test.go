package silverdile_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/sinmetalcraft/goma"
	"golang.org/x/xerrors"

	"github.com/sinmetalcraft/silverdile/v2"
)

func TestImageService_ExistObject(t *testing.T) {
	ctx := context.Background()

	is := newImageService(t)

	cases := []struct {
		name   string
		object string
		want   error
	}{
		{"exist", object, nil},
		{"not found", "momomomomo", silverdile.ErrNotFound},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := is.ExistObject(ctx, &silverdile.ImageOption{
				Bucket: bucket,
				Object: tt.object,
			})
			if !xerrors.Is(got, tt.want) {
				t.Errorf("want %v but got %v", tt.want, got)
			}
		})
	}
}

func TestImageService_ReadAndWriteWithoutResize(t *testing.T) {
	ctx := context.Background()

	is := newImageService(t)

	const existSize = 100
	{ // 先に existSize の Object は作っておく
		w := httptest.NewRecorder()
		err := is.ReadAndWrite(ctx, w, &silverdile.ImageOption{
			Bucket: bucket,
			Object: object,
			Size:   existSize,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	cases := []struct {
		name   string
		object string
		size   int
		want   error
	}{
		{"exist", object, existSize, nil},
		{"not found", object, 666, silverdile.ErrNeedConvert},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			err := is.ReadAndWriteWithoutResize(ctx, w, &silverdile.ImageOption{
				Bucket: bucket,
				Object: object,
				Size:   tt.size,
			})
			if !xerrors.Is(err, tt.want) {
				t.Errorf("wang %v but got %v", tt.want, err)
			}
		})
	}
}

func TestImageService_WithAlterBucket(t *testing.T) {
	ctx := context.Background()

	const bucket = "Hello"
	gcs, err := storage.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}
	gomas := goma.NewStorageService(ctx, gcs)

	is, err := silverdile.NewImageService(ctx, gcs, gomas, silverdile.WithAlterBucket(bucket))
	if err != nil {
		t.Fatal(err)
	}

	if e, g := bucket, silverdile.GetImageServiceAlterBucket(t, is); e != g {
		t.Errorf("want %v but got %v", e, g)
	}
}

func newImageService(t *testing.T) *silverdile.ImageService {
	ctx := context.Background()

	gcs, err := storage.NewClient(ctx)
	if err != nil {
		t.Fatal(err)
	}

	gomas := goma.NewStorageService(ctx, gcs)
	is, err := silverdile.NewImageService(ctx, gcs, gomas)
	if err != nil {
		t.Fatal(err)
	}

	return is
}
