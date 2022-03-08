package silverdile_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sinmetalcraft/silverdile/v3"
)

func TestBuildResizeInfo(t *testing.T) {
	cases := []struct {
		name string
		path string
		want *silverdile.ResizeRequest
	}{
		{"simple", "/hoge/moge.jpg", &silverdile.ResizeRequest{Bucket: "hoge", Object: "moge.jpg", Size: 0}},
		{"s32", "/hoge/fuga/=s32", &silverdile.ResizeRequest{Bucket: "hoge", Object: "fuga", Size: 32}},
		{"object in folder", "/bucket/hoge/fuga/moge.jpg", &silverdile.ResizeRequest{Bucket: "bucket", Object: "hoge/fuga/moge.jpg", Size: 0}},
		{"object in folder with s32", "/bucket/hoge/fuga/moge.jpg/=s32", &silverdile.ResizeRequest{Bucket: "bucket", Object: "hoge/fuga/moge.jpg", Size: 32}},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := silverdile.BuildResizeRequest(tt.path)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("want %+v but got %+v", tt.want, got)
			}
		})
	}
}

func TestBuildResizeInfoError(t *testing.T) {
	cases := []struct {
		name string
		path string
		want error
	}{
		{"empty", "", silverdile.ErrInvalidArgument},
		{"invalid argument", "/", silverdile.ErrInvalidArgument},
		{"invalid argument bucket only", "/hoge/", silverdile.ErrInvalidArgument},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := silverdile.BuildResizeRequest(tt.path)
			if !errors.Is(err, tt.want) {
				t.Errorf("want %+v but got %+v", tt.want, err)
			}
		})
	}
}
