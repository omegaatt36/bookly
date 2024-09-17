package migration

import (
	"fmt"
	"log/slog"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var migrationOptions = gormigrate.Options{
	UseTransaction: true,
}

// Migrator runs migration.
type Migrator struct {
	db         *gorm.DB
	models     []any
	migrations []*gormigrate.Migration
}

// NewMigrator creates migrator.
func NewMigrator(db *gorm.DB, initModels []any, migrations []*gormigrate.Migration) *Migrator {
	return &Migrator{db: db, models: initModels, migrations: migrations}
}

// Upgrade upgrades db schema version.
func (m *Migrator) Upgrade() error {
	if len(m.migrations) == 0 {
		return nil
	}

	mg := gormigrate.New(m.db, &migrationOptions, m.migrations)
	err := mg.Migrate()
	if err != nil {
		return err
	}

	slog.Info(fmt.Sprintf("upgraded to version \"%s\"", m.migrations[len(m.migrations)-1].ID))
	return nil
}

// Rollback rollbacks the last migration.
func (m *Migrator) Rollback() error {
	mg := gormigrate.New(m.db, &migrationOptions, m.migrations)
	if err := mg.RollbackLast(); err != nil {
		return fmt.Errorf("rollback last: %w", err)
	}

	slog.Info("rollback to last")
	return nil
}
