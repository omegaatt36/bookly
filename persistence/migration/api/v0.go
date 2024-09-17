package api

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// CreateExtension defines the initial migration, which creates the extension.
var CreateExtension = gormigrate.Migration{
	ID: "2024-09-17:create-extension-uuid-ossp",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
			return fmt.Errorf("create extension: %w", err)
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
