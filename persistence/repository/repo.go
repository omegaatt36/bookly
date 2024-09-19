package repository

import "gorm.io/gorm"

// GORMRepository represents a generic gorm repository which implements repository interface.
type GORMRepository struct {
	db *gorm.DB
}

// NewGORMRepository creates a new gorm repository.
func NewGORMRepository(db *gorm.DB) *GORMRepository {
	return &GORMRepository{db: db}
}

// AutoMigrate migrates tables.
func (r *GORMRepository) AutoMigrate() error {
	return r.db.AutoMigrate(
		&Account{},
		&Ledger{},
		&User{},
		&Identity{},
	)
}
