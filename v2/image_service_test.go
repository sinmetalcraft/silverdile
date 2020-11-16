package silverdile_test

import (
	"context"
	"testing"

	"github.com/sinmetalcraft/silverdile/v2"
)

func TestWithAlterBucket(t *testing.T) {
	ctx := context.Background()

	const bucket = "Hello"
	is, err := silverdile.NewImageService(ctx, nil, nil, silverdile.WithAlterBucket(bucket))
	if err != nil {
		t.Fatal(err)
	}

	if e, g := bucket, silverdile.GetImageServiceAlterBucket(t, is); e != g {
		t.Errorf("want %v but got %v", e, g)
	}
}
