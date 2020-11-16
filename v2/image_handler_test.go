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
	"github.com/sinmetalcraft/goma"
	"github.com/sinmetalcraft/silverdile/v2"
)

func TestImageHandlerV2_NoResize(t *testing.T) {
	ih := newTestImageHandlers(t)

	// 適当なサイズで2回やってみる
	r := httptest.NewRequest("GET", "https://example.com/v2/sinmetal-ci-silverdile/jun0.jpg", nil)
	w := httptest.NewRecorder()

	ih.ResizeHandler(w, r)

	resp := w.Result()

	if e, g := http.StatusOK, resp.StatusCode; e != g {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Errorf("StatusCode want %v got %v. body=%v", e, g, string(body))
	}
}

func TestImageHandlerV2_Resize(t *testing.T) {
	ih := newTestImageHandlers(t)

	// 適当なサイズで2回やってみる
	size := rand.Intn(600)
	for i := 0; i < 2; i++ {

		r := httptest.NewRequest("GET", fmt.Sprintf("https://example.com/v2/sinmetal-ci-silverdile/jun0.jpg/=s%d", size), nil)
		w := httptest.NewRecorder()

		ih.ResizeHandler(w, r)

		resp := w.Result()

		if e, g := http.StatusOK, resp.StatusCode; e != g {
			body, _ := ioutil.ReadAll(resp.Body)
			t.Errorf("StatusCode want %v got %v. body=%v", e, g, string(body))
		}
	}
}

func newTestImageHandlers(t *testing.T) *silverdile.ImageHandlers {
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

	return silverdile.NewImageHandlers(ctx, is)
}
