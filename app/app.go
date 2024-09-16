package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/urfave/cli/v2"
)

// App is cli wrapper that do some common operation and creates signal handler.
type App struct {
	Flags  []cli.Flag
	Action func(ctx context.Context)
}

func (a *App) action(c *cli.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		slog.DebugContext(ctx, fmt.Sprintf("received signal: %s", sig.String()))
		cancel()
	}()

	// Panic handling.
	defer func() {
		if r := recover(); r != nil {
			slog.ErrorContext(ctx, "Main recovered", slog.Any("panic", r))
			debug.PrintStack()
		}
	}()

	a.Action(ctx)
	slog.InfoContext(ctx, "terminated")

	return nil
}

// Run setups everything and runs Main.
func (a *App) Run() {
	app := cli.NewApp()
	app.Flags = a.Flags
	app.Action = a.action

	err := app.Run(os.Args)
	if err != nil {
		slog.Error("app error", slog.String("error", err.Error()))
	}
}
