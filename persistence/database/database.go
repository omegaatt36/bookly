package database

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/tracelog"
	pgxslog "github.com/mcosta74/pgx-slog"
)

// database defines database instance.
type database struct {
	Opt ConnectOption
	db  *pgxpool.Pool
}

var mutex sync.Mutex
var db *database

// Initialize inits singleton.
func Initialize(opt ConnectOption) error {
	mutex.Lock()
	defer mutex.Unlock()

	if db != nil {
		slog.Warn("database already initialized")
		return nil
	}

	if err := initializeDB(opt); err != nil {
		return fmt.Errorf("database init error: %w", err)
	}

	return nil
}

// Finalize finalizes singleton.
func Finalize() error {
	mutex.Lock()
	defer mutex.Unlock()

	if db == nil {
		return errors.New("database not initialized")
	}

	db.Close()
	db = nil

	return nil
}

func initializeDB(opt ConnectOption) error {
	db = &database{}
	db.Opt = opt
	return db.Open()
}

// GetDB gets db from singleton.
func GetDB() *pgxpool.Pool {
	return db.db
}

// GetDBContext gets a connection from the pool.
func GetDBContext(ctx context.Context) (*pgx.Conn, error) {
	conn, err := db.db.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	return conn.Conn(), nil
}

// Open opens database connection.
func (db *database) Open() error {
	if db.db != nil {
		return nil
	}

	connStr := db.Opt.ConnStr()
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return fmt.Errorf("failed to parse connection string: %w", err)
	}

	// if db.Opt.Silence {
	pgxLogger := pgxslog.NewLogger(
		slog.Default(),
	)
	tracer := &tracelog.TraceLog{
		Logger:   pgxLogger,
		LogLevel: tracelog.LogLevelDebug,
	}

	config.ConnConfig.Tracer = tracer
	// }

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	db.db = pool
	return nil
}

// Close closes db connection.
func (db *database) Close() {
	if db.db != nil {
		db.db.Close()
		db.db = nil
	}
}
