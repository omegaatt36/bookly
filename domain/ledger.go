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
	ID           int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
	AccountID    int32
	Date         time.Time
	Type         LedgerType
	Currency     string
	Amount       decimal.Decimal
	Note         string
	IsAdjustment bool
	AdjustedFrom *int32
	IsVoided     bool
	VoidedAt     *time.Time
	CategoryID   *int32 // Added CategoryID
}

// CreateLedgerRequest defines the request to create a ledger
type CreateLedgerRequest struct {
	AccountID  int32
	Date       time.Time
	Type       LedgerType
	Amount     decimal.Decimal
	Note       string
	CategoryID *int32 // Added CategoryID
}

// UpdateLedgerRequest defines the request to update a ledger
type UpdateLedgerRequest struct {
	ID         int32
	Date       *time.Time
	Type       *LedgerType
	Amount     *decimal.Decimal
	Note       *string
	CategoryID *int32 // Added CategoryID
}

// LedgerRepository represents a ledger repository
type LedgerRepository interface {
	CreateLedger(CreateLedgerRequest) (int32, error)
	GetLedgerByID(int32) (*Ledger, error)
	GetLedgersByAccountID(int32) ([]*Ledger, error)
	UpdateLedger(UpdateLedgerRequest) error
	VoidLedger(id int32) error
	AdjustLedger(originalID int32, adjustment CreateLedgerRequest) error
	DeleteLedger(id int32) error
	GetLedgersByUserIDAndDateRangeAndCategory(ctx context.Context, userID int32, startDate time.Time, endDate time.Time, categoryID int32) ([]*Ledger, error)
}
