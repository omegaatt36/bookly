//go:generate go-enum

package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// AccountStatus represents a ledger account status
// ENUM(active, closed, archived)
type AccountStatus string

// Account represents a ledger account
type Account struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    int32
	Name      string
	Status    AccountStatus
	Currency  string
	Balance   decimal.Decimal
	DeletedAt *time.Time
}

// CreateAccountRequest defines the request to create a ledger account
type CreateAccountRequest struct {
	UserID   int32
	Name     string
	Currency string
}

// UpdateAccountRequest defines the request to update a ledger account
type UpdateAccountRequest struct {
	ID       int32
	UserID   *int32
	Name     *string
	Currency *string
	Status   *AccountStatus
}

// AccountRepository represents a ledger account repository interface
type AccountRepository interface {
	CreateAccount(CreateAccountRequest) error
	GetAccountByID(int32) (*Account, error)
	UpdateAccount(UpdateAccountRequest) error
	DeactivateAccountByID(int32) error
	DeleteAccount(int32) error
	GetAllAccounts() ([]*Account, error)
	GetAccountsByUserID(int32) ([]*Account, error)
}
