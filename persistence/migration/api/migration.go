package api

import (
	"github.com/go-gormigrate/gormigrate/v2"

	v0 "github.com/omegaatt36/bookly/persistence/migration/api/v0"
	v1 "github.com/omegaatt36/bookly/persistence/migration/api/v1"
	v2 "github.com/omegaatt36/bookly/persistence/migration/api/v2"
)

// MigrationList is list of migrations.
var MigrationList = []*gormigrate.Migration{
	&v0.CreateExtension,
	&v1.CreateAccountAndLedger,
	&v2.CreateUser,
}
