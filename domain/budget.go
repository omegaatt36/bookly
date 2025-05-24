package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// PeriodType represents a budget period type
// ENUM(monthly, yearly)
type PeriodType string

// Budget represents a budget
type Budget struct {
	ID         int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UserID     int32
	Name       string
	Category   string
	Amount     decimal.Decimal
	PeriodType PeriodType
	StartDate  time.Time
	EndDate    *time.Time
	IsActive   bool
}

// CreateBudgetRequest defines the request to create a budget
type CreateBudgetRequest struct {
	UserID     int32
	Name       string
	Category   string
	Amount     decimal.Decimal
	PeriodType PeriodType
	StartDate  time.Time
	EndDate    *time.Time
}

// UpdateBudgetRequest defines the request to update a budget
type UpdateBudgetRequest struct {
	ID         int32
	Name       *string
	Category   *string
	Amount     *decimal.Decimal
	PeriodType *PeriodType
	StartDate  *time.Time
	EndDate    *time.Time
	IsActive   *bool
}

// BudgetSummary represents a budget summary with usage
type BudgetSummary struct {
	Budget      *Budget
	UsedAmount  decimal.Decimal
	Percentage  decimal.Decimal
	PeriodStart time.Time
	PeriodEnd   time.Time
}

// BudgetRepository represents a budget repository
type BudgetRepository interface {
	CreateBudget(CreateBudgetRequest) (int32, error)
	GetBudgetByID(int32) (*Budget, error)
	GetBudgetsByUserID(int32) ([]*Budget, error)
	GetActiveBudgetsByUserID(int32) ([]*Budget, error)
	GetBudgetsByUserIDAndCategory(userID int32, category string) ([]*Budget, error)
	GetActiveBudgetByUserIDCategoryAndPeriod(userID int32, category string, periodType PeriodType, date time.Time) (*Budget, error)
	UpdateBudget(UpdateBudgetRequest) error
	DeleteBudget(id int32) error
	GetBudgetUsage(userID int32, category string, startDate, endDate time.Time) (decimal.Decimal, error)
}