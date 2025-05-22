//go:generate go-enum
package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// BudgetPeriod represents a budget period
// ENUM(monthly, yearly)
type BudgetPeriod string

// Budget represents a budget
type Budget struct {
	ID         int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UserID     int32
	Name       string
	Period     BudgetPeriod
	StartDate  time.Time
	EndDate    time.Time
	Amount     decimal.Decimal
	CategoryID int32 // For now, we'll use a simple int32 for CategoryID. We can enhance this later.
}

// CreateBudgetRequest defines the request to create a budget
type CreateBudgetRequest struct {
	UserID     int32
	Name       string
	Period     BudgetPeriod
	StartDate  time.Time
	Amount     decimal.Decimal
	CategoryID int32
}

// UpdateBudgetRequest defines the request to update a budget
type UpdateBudgetRequest struct {
	ID         int32
	Name       *string
	Period     *BudgetPeriod
	StartDate  *time.Time
	Amount     *decimal.Decimal
	CategoryID *int32
}

// BudgetRepository represents a budget repository
type BudgetRepository interface {
	CreateBudget(CreateBudgetRequest) (int32, error)
	GetBudgetByID(int32) (*Budget, error)
	GetBudgetsByUserID(userID int32) ([]*Budget, error)
	UpdateBudget(UpdateBudgetRequest) error
	DeleteBudget(id int32) error
	// Add other necessary methods, e.g., for listing budgets with filters
}

// BudgetUsage represents the usage of a budget
type BudgetUsage struct {
	BudgetID        int32
	BudgetName      string
	BudgetAmount    decimal.Decimal
	SpentAmount     decimal.Decimal
	RemainingAmount decimal.Decimal
	Period          BudgetPeriod
	StartDate       time.Time
	EndDate         time.Time
	CategoryID      int32
}
