package silverdile_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/xerrors"

	"github.com/sinmetalcraft/silverdile/v2"
)

func TestBuildImageOption(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want *silverdile.ImageOption
	}{
		{"s32", "/hoge/fuga/=s32", &silverdile.ImageOption{Bucket: "hoge", Object: "fuga", Size: 32}},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := silverdile.BuildImageOption(tt.url)
			if err != nil {
				t.Fatal(err)
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("want %+v but got %+v", tt.want, got)
			}
		})
	}
}

func TestBuildImageOptionError(t *testing.T) {
	cases := []struct {
		name string
		url  string
		want error
	}{
		{"invalid argument", "/", silverdile.ErrInvalidArgument},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			_, err := silverdile.BuildImageOption(tt.url)
			if !xerrors.Is(err, tt.want) {
				t.Errorf("want %+v but got %+v", err, tt.want)
			}
		})
	}
}
