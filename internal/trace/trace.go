package trace

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/status"
)

func StartSpan(ctx context.Context, name string) context.Context {
	ctx, _ = otel.Tracer("github.com/sinmetalcraft/silverdile/v3").Start(ctx, fmt.Sprintf("/github.com/sinmetalcraft/silverdile/%s", name))
	return ctx
}

func EndSpan(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if err != nil {
		span.SetStatus(toStatus(err))
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}

// toStatus interrogates an error and converts it to an appropriate
// OpenTelemetry status.
func toStatus(err error) (code codes.Code, description string) {
	if err2, ok := err.(*googleapi.Error); ok {
		return codes.Error, fmt.Sprintf("%d:%s", err2.Code, err2.Message)
	} else if s, ok := status.FromError(err); ok {
		return codes.Error, s.Message()
	} else {
		return codes.Error, err.Error()
	}
}

func SetAttributesKV(ctx context.Context, kv map[string]interface{}) {
	span := trace.SpanFromContext(ctx)
	for k, v := range kv {
		switch v := v.(type) {
		case string:
			span.SetAttributes(attribute.String(k, v))
		case bool:
			span.SetAttributes(attribute.Bool(k, v))
		case int:
			span.SetAttributes(attribute.Int(k, v))
		case int64:
			span.SetAttributes(attribute.Int64(k, v))
		case float32:
			span.SetAttributes(attribute.Float64(k, float64(v)))
		case float64:
			span.SetAttributes(attribute.Float64(k, v))
		default:
			span.SetAttributes(attribute.String(k, fmt.Sprintf("%#v", v)))
		}
	}
}
