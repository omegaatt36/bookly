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
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    string
	Name      string
	Status    AccountStatus
	Currency  string
	Balance   decimal.Decimal
}

// CreateAccountRequest defines the request to create a ledger account
type CreateAccountRequest struct {
	UserID   string
	Name     string
	Currency string
}

// UpdateAccountRequest defines the request to update a ledger account
type UpdateAccountRequest struct {
	ID       string
	UserID   *string
	Name     *string
	Currency *string
	Status   *AccountStatus
}

// AccountRepository represents a ledger account repository interface
type AccountRepository interface {
	CreateAccount(CreateAccountRequest) error
	GetAccountByID(string) (*Account, error)
	UpdateAccount(UpdateAccountRequest) error
	DeactivateAccountByID(string) error
	GetAllAccounts() ([]*Account, error)
	GetAccountsByUserID(string) ([]*Account, error)
}
