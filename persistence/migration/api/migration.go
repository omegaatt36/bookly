package api

import (
	"github.com/go-gormigrate/gormigrate/v2"
)

// MigrationList is list of migrations.
var MigrationList = []*gormigrate.Migration{
	&CreateExtension,
}
