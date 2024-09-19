package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// Server represents a router
type Server struct {
	jwtSalt       *string
	jwtSecret     *string
	internalToken *string

	port int

	router http.Handler
}

// NewServer creates a new router.
func NewServer(options ...Option) *Server {
	server := Server{
		port: 8080,
	}

	for _, option := range options {
		option.apply(&server)
	}

	server.registerRouters()

	return &server
}

// Run starts the server.
func (s *Server) Run(ctx context.Context) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", slog.String("error", err.Error()))
		}
	}()

	slog.Info("starting api server", slog.String("addr", srv.Addr))

	if err := srv.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", slog.String("error", err.Error()))
	}
}
