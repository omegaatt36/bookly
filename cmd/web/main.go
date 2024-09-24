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
	"github.com/omegaatt36/bookly/app/web"
)

var config struct {
	logLevel string

	serverURLOption web.ServerURLOption
	portOption      web.PortOption
}

func before(_ *cli.Context) error {
	return initSLog(config.logLevel)
}

func action(ctx context.Context) {
	server := web.NewServer(
		&config.portOption,
		&config.serverURLOption,
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
		&cli.IntFlag{
			Name:        "port",
			EnvVars:     []string{"PORT"},
			Value:       3000,
			DefaultText: "3000",
			Destination: &config.portOption.Port,
		},
		&cli.StringFlag{
			Name:        "server-url",
			EnvVars:     []string{"SERVER_URL"},
			Value:       "http://localhost:8080",
			Destination: &config.serverURLOption.ServerURL,
		},
		&cli.StringFlag{
			Name:        "log-level",
			EnvVars:     []string{"LOG_LEVEL"},
			Value:       "debug",
			Destination: &config.logLevel,
		},
	}

	server := &app.App{
		Action: action,
		Before: before,
		Flags:  cliFlags,
	}

	server.Run()
}
