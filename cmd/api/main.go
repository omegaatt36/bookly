package main

import (
	"context"
	"log/slog"
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

	internalTokenOption api.InternalTokenOption
	jwtOption           api.JWTOption
	portOption          api.PortOption
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
	server := api.NewServer(
		&config.jwtOption,
		&config.internalTokenOption,
		&config.portOption,
	)

	server.Run(ctx)
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
	cliFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "jwt-salt",
			EnvVars:     []string{"JWT_SALT"},
			Value:       "salt",
			Required:    true,
			Destination: &config.jwtOption.JWTSalt,
		},
		&cli.StringFlag{
			Name:        "jwt-secret-key",
			EnvVars:     []string{"JWT_SECRET_KEY"},
			Value:       "secret",
			Required:    true,
			Destination: &config.jwtOption.JWTSecretKey,
		}, &cli.StringFlag{
			Name:        "internal-token",
			EnvVars:     []string{"INTERNAL_TOKEN"},
			Value:       "secret",
			Required:    true,
			Destination: &config.internalTokenOption.InternalToken,
		}, &cli.IntFlag{
			Name:        "port",
			EnvVars:     []string{"PORT"},
			Value:       8080,
			DefaultText: "8080",
			Destination: &config.portOption.Port,
		},
	}
	cliFlags = append(cliFlags, config.databaseConnectionOption.CliFlags()...)

	server := &app.App{
		Action: action,
		Before: before,
		After:  after,
		Flags:  cliFlags,
	}

	server.Run()
}
