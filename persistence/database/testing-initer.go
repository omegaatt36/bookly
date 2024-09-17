package database

import (
	"log/slog"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
)

var (
	// PostgresOpt is default connection option for postgres.
	PostgresOpt = ConnectOption{
		Dialect:  "postgres",
		Host:     "localhost",
		DBName:   "postgres",
		Port:     5433,
		User:     "tester",
		Password: "tester",
	}

	// SQLiteOpt is shared in-memory database.
	SQLiteOpt = ConnectOption{
		Dialect: "sqlite3",
		Host:    "file::memory:?cache=shared",
	}
)

// TestingInitialize creates new db for testing.
func TestingInitialize(opt ConnectOption) (funcFinalize func()) {
	opt.Config.DisableForeignKeyConstraintWhenMigrating = true
	opt.Testing = true

	if opt.Dialect != "postgres" {
		Initialize(opt)

		return func() {
			if err := Finalize(); err != nil {
				slog.Error("Failed to finalize database", "error", err)
				panic(err)
			}
		}
	}

	pg := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(uint32(opt.Port)).
		Username(opt.User).
		Password(opt.Password).
		Database(opt.DBName))
	if err := pg.Start(); err != nil {
		slog.Error("Failed to start postgres", "error", err)
		panic(err)
	}

	if err := Initialize(opt); err != nil {
		if err := pg.Stop(); err != nil {
			slog.Error("Failed to stop postgres", "error", err)
		}

		slog.Error("Failed to initialize database", "error", err)
		panic(err)
	}

	funcFinalize = func() {
		if err := pg.Stop(); err != nil {
			slog.Error("Failed to stop postgres", "error", err)
		}

		if err := Finalize(); err != nil {
			slog.Error("Failed to finalize database", "error", err)
			panic(err)
		}
	}

	if err := GetDB().Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		slog.Error("Failed to install UUID extension", "error", err)
		panic(err)
	}

	return
}
