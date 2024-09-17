package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/urfave/cli/v2"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api"
	"github.com/omegaatt36/bookly/persistence/database"
)

var config struct {
	databaseConnectionOption database.ConnectOption
}

func before(_ *cli.Context) error {
	return database.Initialize(config.databaseConnectionOption)
}

func after(_ *cli.Context) error {
	return database.Finalize()
}

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
	cliFlags := make([]cli.Flag, 0)
	cliFlags = append(cliFlags, config.databaseConnectionOption.CliFlags()...)

	app := app.App{
		Action: action,
		Before: before,
		After:  after,
		Flags:  cliFlags,
	}

	app.Run()
}
