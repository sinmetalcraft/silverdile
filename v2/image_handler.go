package silverdile

import (
	"context"
	"net/http"
	"strings"

	"github.com/vvakame/sdlog/aelog"
)

type ImageHandlers struct {
	imageService *ImageService
}

func NewImageHandlers(ctx context.Context, imageService *ImageService) *ImageHandlers {
	return &ImageHandlers{
		imageService: imageService,
	}
}

func (h *ImageHandlers) ResizeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := aelog.WithHTTPRequest(r.Context(), r)

	l := strings.Split(r.URL.Path, "/")
	if len(l) < 4 {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("invalid argument"))
		if err != nil {
			aelog.Errorf(ctx, "failed write to response. err%+v", err)
		}
		return
	}

	o, err := BuildImageOption(strings.Join(l[1:], "/"))
	if IsErrInvalidArgument(err) {
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
	if IsErrNotFound(err) {
		aelog.Infof(ctx, "404: bucket=%v,object=%v,err=%+v", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		aelog.Errorf(ctx, "failed ReadAndWrite bucket=%v,object=%v err=%+v\n", o.Bucket, o.Object, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
