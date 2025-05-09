package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
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

func rateLimiter(rate float64, capacity int) middleware {
	type bucket struct {
		tokens        float64
		lastTimestamp time.Time
	}

	buckets := make(map[string]*bucket)
	var mu sync.Mutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			defer mu.Unlock()

			b, exists := buckets[ip]
			if !exists {
				b = &bucket{tokens: float64(capacity), lastTimestamp: time.Now()}
				buckets[ip] = b
			}

			now := time.Now()
			timePassed := now.Sub(b.lastTimestamp).Seconds()
			b.tokens = math.Min(float64(capacity), b.tokens+timePassed*rate)
			b.lastTimestamp = now

			if b.tokens < 1 {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			b.tokens--
			next.ServeHTTP(w, r)
		})
	}
}

func authenticated(authenticator domain.Authenticator) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			abortWithUnauthorized := func(message string) {
				bs, err := json.Marshal(engine.ResponseError{
					Code:    app.CodeUnauthorized,
					Message: message,
				})
				if err != nil {
					panic(err)
				}

				http.Error(w, string(bs), http.StatusUnauthorized)
			}

			authToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if authToken == "" {
				abortWithUnauthorized("token is required")
				return
			}

			valid, err := authenticator.ValidateToken(domain.ValidateTokenRequest{
				Token: authToken,
			})
			if err != nil {
				abortWithUnauthorized(err.Error())
				return
			}

			if !valid {
				abortWithUnauthorized("invalid token")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func onlyInternal(internalToken string) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if internalToken != r.Header.Get("INTERNAL-TOKEN") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func chainMiddleware(middlewares ...middleware) middleware {
	return func(next http.Handler) http.Handler {
		for index := len(middlewares) - 1; index >= 0; index-- {
			next = middlewares[index](next)
		}
		return next
	}
}
