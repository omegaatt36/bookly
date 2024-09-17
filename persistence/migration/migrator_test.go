package migration_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/migration"
	apimigration "github.com/omegaatt36/bookly/persistence/migration/api"
)

func TestMigrateAPI(t *testing.T) {
	s := assert.New(t)

	finalize := database.TestingInitialize(database.PostgresOpt)
	defer finalize()

	db := database.GetDB()

	mg := migration.NewMigrator(db, []any{}, apimigration.MigrationList)

	s.NoError(mg.Upgrade())
}
