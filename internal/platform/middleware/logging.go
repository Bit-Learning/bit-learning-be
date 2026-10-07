package middleware

import (
	"log/slog"
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/otel/trace"
)

func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			wrapped := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(wrapped, r)
			span := trace.SpanFromContext(r.Context()).SpanContext()
			logger.LogAttrs(r.Context(), slog.LevelInfo, "request completed",
				slog.String("method", r.Method), slog.String("path", r.URL.Path),
				slog.Int("status", wrapped.Status()), slog.Int64("duration_ms", time.Since(started).Milliseconds()),
				slog.String("request_id", chimiddleware.GetReqID(r.Context())), slog.String("trace_id", span.TraceID().String()))
		})
	}
}
