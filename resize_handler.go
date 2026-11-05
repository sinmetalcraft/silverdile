package silverdile

import (
	"context"
	"errors"
	"fmt"
	"io"
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

func (h *ResizeHandlers) bucketOfAlteredObject(bucket string) string {
	if h.alterBucket != "" {
		return h.alterBucket
	}
	return fmt.Sprintf("alter-%s", bucket)
}

func writeReaderToResponse(w http.ResponseWriter, r *storage.Reader) error {
	if !r.Attrs.LastModified.IsZero() {
		w.Header().Set("last-modified", r.Attrs.LastModified.Format(http.TimeFormat))
	}
	if r.Attrs.ContentType != "" {
		w.Header().Set("content-type", r.Attrs.ContentType)
	}
	if r.Attrs.Size != 0 {
		w.Header().Set("content-length", fmt.Sprintf("%d", r.Attrs.Size))
	}
	w.Header().Set("cache-control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)

	_, err := io.Copy(w, r)
	return err
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

	if rr.Size <= 0 {
		or, err := h.imageService.NewReader(ctx, rr.Bucket, rr.Object)
		if errors.Is(err, storage.ErrObjectNotExist) {
			aelog.Infof(ctx, "404: bucket=%v,object=%v,err=%+v", rr.Bucket, rr.Object, err)
			w.WriteHeader(http.StatusNotFound)
			return
		} else if err != nil {
			aelog.Errorf(ctx, "failed read gs://%s/%s %+v\n", rr.Bucket, rr.Object, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer func() {
			if err := or.Close(); err != nil {
				aelog.Errorf(ctx, "failed close reader gs://%s/%s %+v\n", rr.Bucket, rr.Object, err)
			}
		}()

		if err := writeReaderToResponse(w, or); err != nil {
			aelog.Errorf(ctx, "failed gs://%s/%s write to response %+v\n", rr.Bucket, rr.Object, err)
			return
		}
		return
	}

	alterBucket := h.bucketOfAlteredObject(rr.Bucket)
	resizeObject := fmt.Sprintf("%s_s%d", rr.Object, rr.Size)
	cr, err := h.imageService.NewReader(ctx, alterBucket, resizeObject)
	if errors.Is(err, storage.ErrObjectNotExist) {
		// cacheがなかった場合は、続きに進む
	} else if err != nil {
		aelog.Errorf(ctx, "failed read cache image. gs://%s/%s %+v\n", alterBucket, resizeObject, err)
		// cache読み込みに失敗した場合はとりあえず先に進む
		// errorを返すか、Retryをするかなど選択肢が色々ある
	} else {
		defer func() {
			if err := cr.Close(); err != nil {
				aelog.Errorf(ctx, "failed close reader gs://%s/%s %+v\n", alterBucket, resizeObject, err)
			}
		}()

		if err := writeReaderToResponse(w, cr); err != nil {
			aelog.Errorf(ctx, "failed gs://%s/%s write to response %+v\n", alterBucket, resizeObject, err)
			return
		}
		return
	}

	org, attrs, err := h.imageService.Read(ctx, rr.Bucket, rr.Object)
	if errors.Is(err, storage.ErrObjectNotExist) {
		aelog.Infof(ctx, "404: bucket=%v,object=%v,err=%+v", rr.Bucket, rr.Object, err)
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		aelog.Errorf(ctx, "failed read gs://%s/%s %+v\n", rr.Bucket, rr.Object, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	newImg, err := h.imageService.ResizeToFitLongSide(ctx, org, rr.Size)
	if err != nil {
		aelog.Errorf(ctx, "failed ResizeToFitLongSide gs://%s/%s size=%d %+v\n", rr.Bucket, rr.Object, rr.Size, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 次回RequestのためにResizeしたImageをCacheとしてCloud Storageに書き込む
	err = h.imageService.Write(ctx, alterBucket, resizeObject, newImg, &storage.ObjectAttrs{
		ContentType: attrs.ContentType,
	})
	if err != nil {
		aelog.Errorf(ctx, "failed ImageService.Write gs://%s/%s size=%d %+v\n", alterBucket, resizeObject, rr.Size, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	f, err := ContentTypeToImagingFormat(attrs.ContentType)
	if err != nil {
		aelog.Errorf(ctx, "failed ContentTypeToImagingFormat content-type=%s %+v\n", attrs.ContentType, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("last-modified", attrs.Updated.Format(http.TimeFormat))
	w.Header().Set("content-type", attrs.ContentType)
	w.Header().Set("cache-control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	if err := imaging.Encode(w, newImg, f); err != nil {
		aelog.Errorf(ctx, "failed resize image write to response gs://%s/%s size=%d %+v\n", rr.Bucket, rr.Object, rr.Size, err)
		return
	}
}
