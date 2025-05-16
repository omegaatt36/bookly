package sqlc

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SchemaFS contains the SQL schema files embedded in the binary.
//
//go:embed schema/*.sql
var SchemaFS embed.FS

// MigrateForTest migrates the database schema for testing purposes.
func MigrateForTest(ctx context.Context, db *pgxpool.Pool) error {
	// Get all schema SQL files
	schemaFiles, err := SchemaFS.ReadDir("schema")
	if err != nil {
		return err
	}

	for _, file := range schemaFiles {
		if file.IsDir() {
			continue
		}

		sqlBytes, err := SchemaFS.ReadFile("schema/" + file.Name())
		if err != nil {
			return err
		}

		if _, err := db.Exec(ctx, string(sqlBytes)); err != nil {
			return err
		}
	}

	return nil
}
