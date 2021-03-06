package silverdile

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/vvakame/sdlog/aelog"
	"golang.org/x/xerrors"
)

type ImageHandlers struct {
	basePath     string
	imageService *ImageService
}

func NewImageHandlers(ctx context.Context, basePath string, imageService *ImageService) *ImageHandlers {
	return &ImageHandlers{
		basePath:     basePath,
		imageService: imageService,
	}
}

func (h *ImageHandlers) ResizeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := aelog.WithHTTPRequest(r.Context(), r)

	// /resize/{BUCKET}/{OBJECT} が来るのを期待している
	path := strings.Replace(r.URL.Path, fmt.Sprintf("%s/resize", h.basePath), "", -1)
	o, err := BuildImageOption(path)
	if xerrors.Is(err, ErrInvalidArgument) {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("invalid argument"))
		if err != nil {
			aelog.Errorf(ctx, "failed write to response. err%+v", err)
		}
		return
	} else if err != nil {
		aelog.Errorf(ctx, "failed %+v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	o.CacheControlMaxAge = 3600

	err = h.imageService.ReadAndWrite(ctx, w, o)
	if xerrors.Is(err, ErrNotFound) {
		aelog.Infof(ctx, "404: bucket=%v,object=%v,err=%+v", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		aelog.Errorf(ctx, "failed ReadAndWrite bucket=%v,object=%v err=%+v\n", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
