package silverdile_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloud.google.com/go/storage"
	"github.com/sinmetalcraft/silverdile/v3"
)

func TestResizeHandler_NoResize(t *testing.T) {
	ih := newTestResizeHandlers(t)

	cases := []struct {
		name   string
		object string
	}{
		{"simple", object},
		{"object in folder", "hoge/fuga/jun0.jpg"},
	}

	for _, tt := range cases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", fmt.Sprintf("https://example.com/v3/image/resize/%s/%s", bucket, tt.object), nil)
			w := httptest.NewRecorder()

			ih.ResizeHandler(w, r)

			resp := w.Result()

			if e, g := http.StatusOK, resp.StatusCode; e != g {
				body, _ := ioutil.ReadAll(resp.Body)
				t.Errorf("StatusCode want %v got %v. body=%v", e, g, string(body))
			}
		})
	}

}

func TestResizeHandler_Resize(t *testing.T) {
	ih := newTestResizeHandlers(t)

	// 適当なサイズで2回やってみる
	size := rand.Intn(600)
	for i := 0; i < 2; i++ {

		r := httptest.NewRequest("GET", fmt.Sprintf("https://example.com/v3/image/resize/%s/%s/=s%d", bucket, object, size), nil)
		w := httptest.NewRecorder()

		ih.ResizeHandler(w, r)

		resp := w.Result()

		if e, g := http.StatusOK, resp.StatusCode; e != g {
			body, _ := ioutil.ReadAll(resp.Body)
			t.Errorf("StatusCode want %v got %v. body=%v", e, g, string(body))
		}
	}
}

func newTestResizeHandlers(t *testing.T) *silverdile.ResizeHandlers {
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

	return silverdile.NewResizeHandlers(ctx, "/v3/image", fmt.Sprintf("alter-%s", bucket), is)
}
