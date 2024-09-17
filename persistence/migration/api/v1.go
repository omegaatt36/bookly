package api

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// Account represents the database model for an account
type Account struct {
	ID        string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string          `gorm:"type:varchar(255);not null"`
	Status    string          `gorm:"type:varchar(20);not null"`
	Currency  string          `gorm:"type:varchar(3);not null"`
	Balance   decimal.Decimal `gorm:"type:decimal(20,2);not null"`

	Ledgers []Ledger
}

// Ledger represents the database model for a ledger entry
type Ledger struct {
	ID           string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	AccountID    string          `gorm:"type:uuid;not null"`
	Date         time.Time       `gorm:"not null"`
	Type         string          `gorm:"type:varchar(20);not null"`
	Amount       decimal.Decimal `gorm:"type:decimal(20,2);not null"`
	Note         string          `gorm:"type:text"`
	IsAdjustment bool            `gorm:"not null"`
	AdjustedFrom *string         `gorm:"type:uuid"`
	IsVoided     bool            `gorm:"not null"`
	VoidedAt     *time.Time
}

// CreateAccountAndLedger defines the migration, which creates the accounts and ledgers.
var CreateAccountAndLedger = gormigrate.Migration{
	ID: "2024-09-17:create-accounts-and-ledgers",
	Migrate: func(tx *gorm.DB) error {
		return tx.AutoMigrate(&Account{}, &Ledger{})
	},
	Rollback: func(tx *gorm.DB) error {
		return tx.Migrator().DropTable(&Ledger{}, &Account{})
	},
}
