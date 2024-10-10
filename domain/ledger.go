//go:generate go-enum

package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// EditableDuration is the duration after which a ledger is editable
const EditableDuration = time.Minute * 15

// LedgerType represents a ledger type
// ENUM(balance, income, expense, transfer)
type LedgerType string

// Ledger represents a ledger
type Ledger struct {
	ID           string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	AccountID    string
	Date         time.Time
	Type         LedgerType
	Currency     string
	Amount       decimal.Decimal
	Note         string
	IsAdjustment bool
	AdjustedFrom *string
	IsVoided     bool
	VoidedAt     *time.Time
}

// CreateLedgerRequest defines the request to create a ledger
type CreateLedgerRequest struct {
	AccountID string
	Date      time.Time
	Type      LedgerType
	Amount    decimal.Decimal
	Note      string
}

// UpdateLedgerRequest defines the request to update a ledger
type UpdateLedgerRequest struct {
	ID     string
	Date   *time.Time
	Type   *LedgerType
	Amount *decimal.Decimal
	Note   *string
}

// LedgerRepository represents a ledger repository
type LedgerRepository interface {
	CreateLedger(CreateLedgerRequest) error
	GetLedgerByID(string) (*Ledger, error)
	GetLedgersByAccountID(string) ([]*Ledger, error)
	UpdateLedger(UpdateLedgerRequest) error
	VoidLedger(id string) error
	AdjustLedger(originalID string, adjustment CreateLedgerRequest) error
}
