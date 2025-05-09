package migration

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrator runs migration.
type Migrator struct {
	db        *pgxpool.Pool
	migrTable string
}

// NewMigrator creates migrator.
func NewMigrator(db *pgxpool.Pool) *Migrator {
	return &Migrator{
		db:        db,
		migrTable: "schema_migrations",
	}
}

// Upgrade upgrades db schema version by applying all pending migration files.
func (m *Migrator) Upgrade() error {
	slog.Info("Starting database migration")

	// Ensure migration table exists
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	// Get applied migrations
	appliedMigrations, err := m.getAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Get all migration files
	migrationFiles, err := m.getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Apply migrations
	for _, file := range migrationFiles {
		if _, applied := appliedMigrations[file]; applied {
			slog.Debug("Migration already applied", slog.String("migration", file))
			continue
		}

		slog.Info("Applying migration", slog.String("migration", file))
		if err := m.applyMigration(file); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", file, err)
		}

		slog.Info("Applied migration", slog.String("migration", file))
	}

	slog.Info("Database migration completed")
	return nil
}

// Rollback rolls back the latest migration.
func (m *Migrator) Rollback() error {
	slog.Info("Rolling back latest migration")

	// Ensure migration table exists
	if err := m.ensureMigrationTable(); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	// Get last applied migration
	ctx := context.Background()
	var lastMigration string
	err := m.db.QueryRow(ctx, `
		SELECT migration FROM `+m.migrTable+`
		ORDER BY applied_at DESC LIMIT 1
	`).Scan(&lastMigration)
	if err != nil {
		return fmt.Errorf("failed to get last applied migration: %w", err)
	}

	// Remove migration from the table
	_, err = m.db.Exec(ctx, `
		DELETE FROM `+m.migrTable+`
		WHERE migration = $1
	`, lastMigration)
	if err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	slog.Info("Rolled back migration", slog.String("migration", lastMigration))
	return nil
}

// ensureMigrationTable creates the migration table if it doesn't exist.
func (m *Migrator) ensureMigrationTable() error {
	ctx := context.Background()
	_, err := m.db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS `+m.migrTable+` (
			migration TEXT PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

// getAppliedMigrations returns a map of applied migrations.
func (m *Migrator) getAppliedMigrations() (map[string]bool, error) {
	ctx := context.Background()
	rows, err := m.db.Query(ctx, `
		SELECT migration FROM `+m.migrTable+`
		ORDER BY applied_at
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	migrations := make(map[string]bool)
	for rows.Next() {
		var migration string
		if err := rows.Scan(&migration); err != nil {
			return nil, err
		}
		migrations[migration] = true
	}

	return migrations, rows.Err()
}

// getMigrationFiles returns a sorted list of migration files.
func (m *Migrator) getMigrationFiles() ([]string, error) {
	entries, err := EmbeddedMigrations.ReadDir("sql")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, entry.Name())
	}

	sort.Strings(files)
	return files, nil
}

// applyMigration applies a migration to the database.
func (m *Migrator) applyMigration(filename string) error {
	ctx := context.Background()
	filePath := "sql/" + filename

	// Read migration file
	content, err := EmbeddedMigrations.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Start transaction
	tx, err := m.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			errRollback := tx.Rollback(ctx)
			if errRollback != nil {
				slog.Error("Failed to rollback transaction", slog.String("error", errRollback.Error()))
			}
		}
	}()

	// Execute migration
	_, err = tx.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	// Record migration
	_, err = tx.Exec(ctx, `
		INSERT INTO `+m.migrTable+` (migration) VALUES ($1)
	`, filename)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
