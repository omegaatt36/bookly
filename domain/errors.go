package domain

import "errors"

// ErrNotFound indicates that a requested resource was not found.
var ErrNotFound = errors.New("resource not found")

// ErrBudgetNotFound indicates that a requested budget was not found.
var ErrBudgetNotFound = errors.New("budget not found")

// ErrInvalidBudgetPeriod indicates that the budget period is invalid.
var ErrInvalidBudgetPeriod = errors.New("invalid budget period")

// ErrBudgetCategoryRequired indicates that category is required for budget.
var ErrBudgetCategoryRequired = errors.New("category is required for budget")
