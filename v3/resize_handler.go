package silverdile

import (
	"context"
	"errors"
	"fmt"
	"image"
	"net/http"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/disintegration/imaging"
	"github.com/vvakame/sdlog/aelog"
)

type ResizeHandlers struct {
	basePath     string
	alterBucket  string
	imageService *ImageService
}

func NewResizeHandlers(ctx context.Context, basePath string, alterBucket string, imageService *ImageService) *ResizeHandlers {
	return &ResizeHandlers{
		basePath:     basePath,
		alterBucket:  alterBucket,
		imageService: imageService,
	}
}

// ResizeHandler is Pathで指定したCloud Storage上のObjectを指定したサイズにResizeして返す
// このまま使うこともできるが、Exampleの意味合いが強い
func (h *ResizeHandlers) ResizeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := aelog.WithHTTPRequest(r.Context(), r)

	// /resize/{BUCKET}/{OBJECT} が来るのを期待している
	path := strings.Replace(r.URL.Path, fmt.Sprintf("%s/resize", h.basePath), "", -1)
	rr, err := BuildResizeRequest(path)
	if errors.Is(err, ErrInvalidArgument) {
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

	resizeObject := fmt.Sprintf("%s_%d", rr.Object, rr.Size)
	if rr.Size > 0 {
		cache, attrs, err := h.imageService.Read(ctx, h.alterBucket, resizeObject)
		if errors.Is(err, storage.ErrObjectNotExist) {
			// cacheがなかった場合は、続きに進む
		} else if err != nil {
			aelog.Errorf(ctx, "failed read cache image. gs://%s/%s %+v\n", h.alterBucket, resizeObject, err)
			// cache読み込みに失敗した場合はとりあえず先に進む
		} else {
			w.Header().Set("last-modified", attrs.Updated.Format(http.TimeFormat))
			w.Header().Set("content-type", attrs.ContentType)
			w.Header().Set("content-length", fmt.Sprintf("%d", attrs.Size))
			w.Header().Set("cache-control", "public, max-age=3600")
			f, err := ContentTypeToImagingFormat(attrs.ContentType)
			if err != nil {
				aelog.Errorf(ctx, "failed ContentTypeToImagingFormat content-type=%s %+v\n", attrs.ContentType, err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
			if err := imaging.Encode(w, cache, f); err != nil {
				aelog.Errorf(ctx, "failed gs://%s/%s write to response %+v\n", h.alterBucket, resizeObject, err)
				return
			}
		}
	}

	var result image.Image
	org, attrs, err := h.imageService.Read(ctx, rr.Bucket, rr.Object)
	if errors.Is(err, storage.ErrObjectNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		aelog.Errorf(ctx, "failed read gs://%s/%s %+v\n", rr.Bucket, rr.Object, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	result = org

	if rr.Size > 0 {
		newImg, err := h.imageService.ResizeToFitLongSide(ctx, org, rr.Size)
		if err != nil {
			aelog.Errorf(ctx, "failed ResizeToFitLongSide gs://%s/%s size=%d %+v\n", rr.Bucket, rr.Object, rr.Size, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		result = newImg

		// 次回RequestのためにResizeしたImageをCacheとしてCloud Storageに書き込む
		err = h.imageService.Write(ctx, h.alterBucket, resizeObject, newImg, &storage.ObjectAttrs{
			ContentType: attrs.ContentType,
		})
		if err != nil {
			aelog.Errorf(ctx, "failed ImageService.Write gs://%s/%s size=%d %+v\n", h.alterBucket, resizeObject, rr.Size, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("last-modified", attrs.Updated.Format(http.TimeFormat))
	w.Header().Set("content-type", attrs.ContentType)
	w.Header().Set("cache-control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	f, err := ContentTypeToImagingFormat(attrs.ContentType)
	if err != nil {
		aelog.Errorf(ctx, "failed ContentTypeToImagingFormat content-type=%s %+v\n", attrs.ContentType, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := imaging.Encode(w, result, f); err != nil {
		aelog.Errorf(ctx, "failed resize image write to response gs://%s/%s size=%d %+v\n", rr.Bucket, rr.Object, rr.Size, err)
		return
	}
}
