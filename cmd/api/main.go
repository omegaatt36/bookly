package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api"
)

func action(ctx context.Context) {
	r := http.NewServeMux()

	api.RegisterRouters(r)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", slog.String("error", err.Error()))
		}
	}()

	if err := srv.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", slog.String("error", err.Error()))
	}
}

func main() {
	app := app.App{
		Action: action,
		Flags:  []cli.Flag{},
	}

	app.Run()
}
