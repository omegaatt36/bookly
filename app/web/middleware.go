package web

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/omegaatt36/bookly/app/web/api"
)

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

type contextKey struct{}

var userIDKey = contextKey{}

func authenticatedHandler(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("token")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		tokenStr := cookie.Value

		token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), userIDKey, int32(userID)))

		next.ServeHTTP(w, r)
	}
}

func (s *Server) getUserFromContext(r *http.Request) *api.User {
	userID, ok := r.Context().Value(userIDKey).(int32)
	if !ok {
		slog.Error("userID not found in context")
		return &api.User{}
	}

	return &api.User{
		ID:       userID,
		Name:     "name",
		Nickname: "nickname",
	}
}
