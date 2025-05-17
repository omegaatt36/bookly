package domain

import (
	"time"
)

// BankAccount represents a bank account linked to a ledger account
type BankAccount struct {
	ID            int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
	AccountID     int32
	AccountNumber string
	BankName      string
	BranchName    string
	SwiftCode     string
	DeletedAt     *time.Time
}

// CreateBankAccountRequest defines the request to create a bank account
type CreateBankAccountRequest struct {
	AccountID     int32
	AccountNumber string
	BankName      string
	BranchName    string
	SwiftCode     string
}

// UpdateBankAccountRequest defines the request to update a bank account
type UpdateBankAccountRequest struct {
	ID            int32
	AccountNumber *string
	BankName      *string
	BranchName    *string
	SwiftCode     *string
}

// BankAccountRepository represents a bank account repository interface
type BankAccountRepository interface {
	CreateBankAccount(CreateBankAccountRequest) error
	GetBankAccountByID(int32) (*BankAccount, error)
	GetBankAccountByAccountID(int32) (*BankAccount, error)
	UpdateBankAccount(UpdateBankAccountRequest) error
	DeleteBankAccount(int32) error
}