package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"

	slogzap "github.com/samber/slog-zap/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api"
	"github.com/omegaatt36/bookly/persistence/database"
)

var config struct {
	databaseConnectionOption database.ConnectOption
	logLevel                 string
}

func before(_ *cli.Context) error {
	if err := initSLog(config.logLevel); err != nil {
		return err
	}

	return database.Initialize(config.databaseConnectionOption)
}

func after(_ *cli.Context) error {
	return database.Finalize()
}

func action(ctx context.Context) {
	r := api.NewRouter()

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

func initSLog(logLevel string) error {
	level := zapcore.DebugLevel
	if err := level.Set(logLevel); err != nil {
		level = zapcore.DebugLevel // default level
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), level),
	)

	zapLogger := zap.New(core, zap.AddCaller())

	slog.SetDefault(slog.New(slogzap.Option{Level: slog.LevelDebug, Logger: zapLogger}.NewZapHandler()))

	return nil
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
