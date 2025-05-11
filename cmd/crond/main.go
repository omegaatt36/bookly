package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	slogzap "github.com/samber/slog-zap/v2"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/repository"
	"github.com/omegaatt36/bookly/service/bookkeeping"
)

var config struct {
	databaseConnectionOption database.ConnectOption
	logLevel                 string
}

// initSLog initializes structured logging using slog and zap.
func initSLog(logLevel string) error {
	level := zapcore.DebugLevel
	if err := level.Set(logLevel); err != nil {
		level = zapcore.DebugLevel // default level
	}

	encoderConfig := zap.NewDevelopmentEncoderConfig()
	// Customize time format for console encoder
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewTee(
		zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), zapcore.AddSync(os.Stdout), level),
	)

	zapLogger := zap.New(core, zap.AddCaller())

	slog.SetDefault(slog.New(slogzap.Option{Level: slog.LevelDebug, Logger: zapLogger}.NewZapHandler()))

	return nil
}

// before runs before the action, initializing logging and database.
func before(_ *cli.Context) error {
	if err := initSLog(config.logLevel); err != nil {
		return err
	}

	return database.Initialize(config.databaseConnectionOption)
}

// after runs after the action, finalizing database connections.
func after(_ *cli.Context) error {
	return database.Finalize()
}

// action is the main entry point for the crond application logic.
func action(ctx context.Context) {
	slog.Info("Starting bookkeeping crond")

	// Get database connection
	db := database.GetDB()

	// Create repository instance
	repo := repository.NewSQLCRepository(db)

	service := bookkeeping.NewService(bookkeeping.NewServiceRequest{
		AccountRepo:              repo,
		LedgerRepo:               repo,
		RecurringTransactionRepo: repo,
		ReminderRepo:             repo,
	})

	funcProcessDueTransactions := func() {
		slog.Info("Running scheduled ProcessDueTransactions")
		ctx, cancle := context.WithTimeout(ctx, time.Second*30)
		defer cancle()

		if err := service.ProcessDueTransactions(ctx); err != nil {
			slog.Error("Failed to process recurring transactions", "error", err)
			return
		}

		slog.Info("Successfully processed recurring transactions")
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		slog.Error("Failed to create scheduler", "error", err)
		panic(err) // Crond cannot function without a scheduler
	}
	defer func() {
		slog.Info("Scheduler deferred shutdown")
		if err := s.Shutdown(); err != nil {
			slog.Error("Failed to shutdown scheduler", "error", err)
			return
		}

		slog.Info("Scheduler stopped, crond exiting")
	}()

	job, err := s.NewJob(
		gocron.DurationJob(1*time.Hour), // Schedule every hour
		gocron.NewTask(funcProcessDueTransactions),
		gocron.WithSingletonMode(gocron.LimitModeWait), // Ensure only one instance runs at a time
	)
	if err != nil {
		slog.Error("Failed to schedule job", "error", err)
		// Cannot schedule the main job, crond is non-functional
		panic(err)
	}
	slog.Info("Job scheduled", "job_id", job.ID().String(), "schedule", "every 1 hour")

	// Start the scheduler in a goroutine
	s.Start()
	slog.Info("Scheduler started")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	slog.Info("Received shutdown signal, stopping scheduler...")
}

func main() {
	cliFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			EnvVars:     []string{"LOG_LEVEL"},
			Value:       "debug",
			Destination: &config.logLevel,
		},
	}
	// Add database connection flags from the database package
	cliFlags = append(cliFlags, config.databaseConnectionOption.CliFlags()...)

	crondApp := &app.App{
		Action: action,
		Before: before,
		After:  after,
		Flags:  cliFlags,
	}

	// Run the CLI application
	crondApp.Run()
}
