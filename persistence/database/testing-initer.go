package database

import (
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	// PostgresOpt is default connection option for postgres.
	PostgresOpt = ConnectOption{
		Dialect:  "postgres",
		Host:     "localhost",
		DBName:   "postgres",
		Port:     5432,
		User:     "postgres",
		Password: "postgres",
	}

	// SQLiteOpt is shared in-memory database.
	SQLiteOpt = ConnectOption{
		Dialect: "sqlite3",
		Host:    "file::memory:?cache=shared",
	}
)

var cnt atomic.Int32

func randomDBName() string {
	return fmt.Sprintf("testing_%v_%d", time.Now().UnixNano(), cnt.Add(1))
}

// TestingInitialize creates new db for testing.
func TestingInitialize(opt ConnectOption) (funcFinalize func()) {
	opt.Config.DisableForeignKeyConstraintWhenMigrating = true
	opt.Testing = true

	if opt.Dialect != "postgres" {
		if err := Initialize(opt); err != nil {
			panic(err)
		}

		return func() {
			if err := Finalize(); err != nil {
				slog.Error("Failed to finalize database", "error", err)
				panic(err)
			}
		}
	}

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
		opt.Host, opt.Port, opt.User, opt.Password)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to postgres", "error", err)
		panic(err)
	}

	randomDBName := randomDBName()

	if err := db.Exec(fmt.Sprintf("CREATE DATABASE %s", randomDBName)).Error; err != nil {
		slog.Error("Failed to create test database", "error", err)
		panic(err)
	}

	opt.DBName = randomDBName
	if err := Initialize(opt); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		panic(err)
	}

	funcFinalize = func() {
		if err := Finalize(); err != nil {
			slog.Error("Failed to finalize database", "error", err)
			panic(err)
		}

		if err := db.Exec(fmt.Sprintf("DROP DATABASE %s", randomDBName)).Error; err != nil {
			slog.Error("Failed to drop test database", "error", err)
			panic(err)
		}
	}

	if err := GetDB().Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		slog.Error("Failed to install UUID extension", "error", err)
		panic(err)
	}

	return
}
