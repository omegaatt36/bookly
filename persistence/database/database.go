package database

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// database defines database instance.
type database struct {
	Opt ConnectOption
	db  *gorm.DB
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
	return db.Close()
}

func initializeDB(opt ConnectOption) error {
	db = &database{}
	db.Opt = opt
	if err := db.Open(); err != nil {
		return err
	}

	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}

	if opt.Dialect == "postgres" {
		sqlDB.SetConnMaxLifetime(10 * time.Minute)
		sqlDB.SetMaxIdleConns(20)
		sqlDB.SetMaxOpenConns(20)
	}

	return nil
}

// GetDB gets db from singleton.
func GetDB() *gorm.DB {
	return db.getDB()
}

// AutoMigrate migrates table.
func AutoMigrate(models []any) {
	for _, m := range models {
		err := db.getDB().AutoMigrate(m)
		if err != nil {
			slog.Error("AutoMigrate error", slog.String("error", err.Error()))
		}
	}
}

// getDB get gorm db instance.
func (db *database) getDB() *gorm.DB {
	if db.db == nil {
		panic("database is not initialized.")
	}

	return db.db.Session(&gorm.Session{})
}

// Open opens database connection.
func (db *database) Open() error {
	if db.db != nil {
		return nil
	}

	dialector := db.Opt.Dialector()
	if dialector == nil {
		return fmt.Errorf("gorm driver open dialector fail, connect str: (%v)", db.Opt.ConnStr())
	}

	if db.Opt.Silence {
		db.Opt.Config.Logger = logger.Discard
	} else if db.Opt.Logger != nil {
		db.Opt.Config.Logger = db.Opt.Logger
	}
	conn, err := gorm.Open(dialector, &db.Opt.Config)
	if err != nil {
		return fmt.Errorf("sql.Open(%v): %w", db.Opt.ConnStr(), err)
	}

	db.db = conn

	return nil
}

// Close closes db connection.
func (db *database) Close() error {
	realConn, err := db.db.DB()
	if err != nil {
		return fmt.Errorf("get db connection when close db error: %w", err)
	}

	if err := realConn.Close(); err != nil {
		return fmt.Errorf("close db connection error: %w", err)
	}

	db.db = nil

	return nil
}
