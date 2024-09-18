package api

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type middleware func(http.Handler) http.Handler

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startsAt := time.Now()

		wrappedWriter := &wrappedWriter{ResponseWriter: w}

		next.ServeHTTP(wrappedWriter, r)

		ctx := r.Context()

		method := r.Method
		path := r.URL.Path
		duration := time.Since(startsAt)

		slog.InfoContext(ctx, fmt.Sprintf("%s %d %s %s", method, wrappedWriter.statusCode, duration.String(), path))
	})
}

func chain(middlewares ...middleware) middleware {
	return func(next http.Handler) http.Handler {
		for index := len(middlewares) - 1; index >= 0; index-- {
			next = middlewares[index](next)
		}
		return next
	}
}
