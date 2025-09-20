package trace

import "context"

type ctxKey string

const TraceIDKey ctxKey = "trace_id"

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

func TraceIDFromContext(ctx context.Context) string {
	if v := ctx.Value(TraceIDKey); v != nil {
		if tid, ok := v.(string); ok {
			return tid
		}
	}
	return ""
}
