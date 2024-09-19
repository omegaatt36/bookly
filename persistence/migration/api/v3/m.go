package v3

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Identity represents the database model for an identity
type Identity struct {
	ID         int    `gorm:"primary_key"`
	UserID     string `gorm:"type:uuid;not null;uniqueIndex:idx_user_provider_identifier"`
	Provider   string `gorm:"type:varchar(20);not null;uniqueIndex:idx_user_provider_identifier;uniqueIndex:idx_provider_identifier"`
	Identifier string `gorm:"type:varchar(255);not null;uniqueIndex:idx_user_provider_identifier;uniqueIndex:idx_provider_identifier"`
	Credential string `gorm:"type:varchar(255);not null"`
	LastUsedAt time.Time
}

// CreateIdentity defines the migration, which creates the users.
var CreateIdentity = gormigrate.Migration{
	ID: "2024-09-19:create-identities",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&Identity{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable(&Identity{})
	},
}
