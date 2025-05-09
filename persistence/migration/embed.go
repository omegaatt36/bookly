package migration

import "embed"

//go:embed sql/*.sql
// EmbeddedMigrations contains the SQL migration files embedded in the binary
var EmbeddedMigrations embed.FS
