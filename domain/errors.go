package domain

import "errors"

// ErrNotFound indicates that a requested resource was not found.
var ErrNotFound = errors.New("resource not found")

// General validation errors
var (
	ErrCategoryNameRequired = errors.New("category name is required")
	ErrUserIDRequired       = errors.New("user ID is required")
	ErrForbidden            = errors.New("forbidden")
	ErrBudgetNameRequired   = errors.New("budget name is required")
	ErrCategoryIDRequired   = errors.New("category ID is required")
	ErrAmountMustBePositive = errors.New("amount must be positive")
	ErrLedgerNotFound       = errors.New("ledger not found")    // Potentially duplicates ErrNotFound if not distinguished
	ErrAccountNotFound      = errors.New("account not found")  // Potentially duplicates ErrNotFound if not distinguished
)
