package silverdile

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"go.opencensus.io/trace"
)

// startSpan is Start Trace Span
func startSpan(ctx context.Context, name string) context.Context {
	ctx, _ = trace.StartSpan(ctx, fmt.Sprintf("github.com/sinmetalcraft/silverdile/%s", name))
	return ctx
}

// endSpan is End Trace Span
func endSpan(ctx context.Context, err error) {
	span := trace.FromContext(ctx)
	if err != nil {
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeInternal, // TODO sinmetal : もう少し丁寧にCode入れても良いかもしれない
			Message: err.Error(),
		})
	}
	span.End()
}

// addObjectAttribute is Cloud Storage Object の情報を Trace Span に入れる
func addObjectAttribute(ctx context.Context, attrs *storage.ObjectAttrs, bucket string) {
	span := trace.FromContext(ctx)
	if span == nil {
		return
	}

	span.AddAttributes(trace.StringAttribute("BucketName", bucket))
	span.AddAttributes(trace.StringAttribute("ObjectName", attrs.Name))
	span.AddAttributes(trace.Int64Attribute("ObjectSize", attrs.Size))
}

// addImageOption is ImageOption の情報を Trace Span に入れる
func addImageOption(ctx context.Context, o *ImageOption) {
	span := trace.FromContext(ctx)
	if span == nil {
		return
	}

	span.AddAttributes(trace.Int64Attribute("ResizeLongSide", int64(o.Size)))
}
