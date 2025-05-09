package database

import (
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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
)

var cnt atomic.Int32

func randomDBName() string {
	return fmt.Sprintf("testing_%v_%d", time.Now().UnixNano(), cnt.Add(1))
}

// TestingInitialize creates new db for testing.
func TestingInitialize(opt ConnectOption) (funcFinalize func()) {
	opt.Testing = true

	if opt.Dialect != "postgres" {
		slog.Error("Only postgres is supported for testing")
		panic("Only postgres is supported for testing")
	}

	// Connect to postgres database to create test database
	adminConnStr := fmt.Sprintf("postgres://%s:%s@%s:%v/postgres?sslmode=disable",
		opt.User, opt.Password, opt.Host, opt.Port)
	
	adminPool, err := pgxpool.New(context.Background(), adminConnStr)
	if err != nil {
		slog.Error("Failed to connect to postgres", "error", err)
		panic(err)
	}
	defer adminPool.Close()

	// Generate random database name for this test
	randomDBName := randomDBName()

	// Create test database
	_, err = adminPool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", randomDBName))
	if err != nil {
		slog.Error("Failed to create test database", "error", err)
		panic(err)
	}

	// Set database name in options
	opt.DBName = randomDBName
	if err := Initialize(opt); err != nil {
		slog.Error("Failed to initialize database", "error", err)
		panic(err)
	}

	// Create UUID extension in new database
	_, err = GetDB().Exec(context.Background(), "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")
	if err != nil {
		slog.Error("Failed to install UUID extension", "error", err)
		panic(err)
	}

	// Return cleanup function
	funcFinalize = func() {
		if err := Finalize(); err != nil {
			slog.Error("Failed to finalize database", "error", err)
			panic(err)
		}

		// Reconnect to postgres database to drop test database
		adminPool, err := pgxpool.New(context.Background(), adminConnStr)
		if err != nil {
			slog.Error("Failed to reconnect to postgres for cleanup", "error", err)
			panic(err)
		}
		defer adminPool.Close()

		// Drop test database
		_, err = adminPool.Exec(context.Background(), fmt.Sprintf("DROP DATABASE %s", randomDBName))
		if err != nil {
			slog.Error("Failed to drop test database", "error", err)
			panic(err)
		}
	}

	return
}