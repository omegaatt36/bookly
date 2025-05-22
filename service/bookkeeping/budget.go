package bookkeeping

import (
	"context"
	"time"

	"github.com/omegaatt36/bookly/domain"
	"github.com/shopspring/decimal"
)

// BudgetService provides budget-related operations
type BudgetService struct {
	budgetRepo      domain.BudgetRepository   // Changed to specific repository
	ledgerRepo      domain.LedgerRepository   // Added specific repository
	categoryService *CategoryService          // To validate category ownership
}

// NewBudgetService creates a new BudgetService
func NewBudgetService(budgetRepo domain.BudgetRepository, ledgerRepo domain.LedgerRepository, categoryService *CategoryService) *BudgetService { // Changed parameters
	return &BudgetService{budgetRepo: budgetRepo, ledgerRepo: ledgerRepo, categoryService: categoryService}
}

// CreateBudget creates a new budget for a user
func (s *BudgetService) CreateBudget(ctx context.Context, req domain.CreateBudgetRequest) (int32, error) {
	if req.Name == "" {
		return 0, domain.ErrBudgetNameRequired // Define this error
	}
	if req.UserID == 0 {
		return 0, domain.ErrUserIDRequired
	}
	if req.CategoryID == 0 {
		return 0, domain.ErrCategoryIDRequired // Define this error
	}
	if !req.Period.IsValid() {
		return 0, domain.ErrInvalidBudgetPeriod
	}
	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return 0, domain.ErrAmountMustBePositive // Define this
	}

	// Validate category ownership
	_, err := s.categoryService.GetCategory(ctx, req.CategoryID, req.UserID)
	if err != nil {
		return 0, err // Could be not found or forbidden
	}
	
	// Default StartDate to the beginning of the current month/year if not provided
    if req.StartDate.IsZero() {
        now := time.Now()
        if req.Period == domain.BudgetPeriodMonthly {
            req.StartDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
        } else if req.Period == domain.BudgetPeriodYearly {
            req.StartDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
        }
    }

	return s.budgetRepo.CreateBudget(ctx, req) // Changed to s.budgetRepo
}

// GetBudget retrieves a budget by ID, ensuring it belongs to the user
func (s *BudgetService) GetBudget(ctx context.Context, budgetID int32, userID int32) (*domain.Budget, error) {
	budget, err := s.budgetRepo.GetBudgetByID(ctx, budgetID) // Changed to s.budgetRepo
	if err != nil {
		return nil, err
	}
	if budget.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return budget, nil
}

// ListBudgets retrieves all budgets for a user
func (s *BudgetService) ListBudgets(ctx context.Context, userID int32) ([]*domain.Budget, error) {
	if userID == 0 {
		return nil, domain.ErrUserIDRequired
	}
	return s.budgetRepo.GetBudgetsByUserID(ctx, userID) // Changed to s.budgetRepo
}

// UpdateBudget updates an existing budget
func (s *BudgetService) UpdateBudget(ctx context.Context, req domain.UpdateBudgetRequest, userID int32) error {
	// Validate ownership of the budget itself
	budget, err := s.budgetRepo.GetBudgetByID(ctx, req.ID) // Changed to s.budgetRepo
	if err != nil {
		return nil, err
	}
	if budget.UserID != userID {
		return nil, domain.ErrForbidden
	}

	// If CategoryID is being updated, validate ownership of the new category
	if req.CategoryID != nil && *req.CategoryID != 0 {
		_, err := s.categoryService.GetCategory(ctx, *req.CategoryID, userID)
		if err != nil {
			return err
		}
	}
    
    // Add other validation as needed (e.g. for amount, period)
    if req.Amount != nil && req.Amount.LessThanOrEqual(decimal.Zero) {
        return domain.ErrAmountMustBePositive
    }
    if req.Period != nil && !req.Period.IsValid() {
        return domain.ErrInvalidBudgetPeriod
    }

	return s.budgetRepo.UpdateBudget(ctx, req) // Changed to s.budgetRepo
}

// DeleteBudget deletes a budget
func (s *BudgetService) DeleteBudget(ctx context.Context, budgetID int32, userID int32) error {
	budget, err := s.budgetRepo.GetBudgetByID(ctx, budgetID) // Changed to s.budgetRepo
	if err != nil {
		return nil, err
	}
	if budget.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return s.budgetRepo.DeleteBudget(ctx, budgetID) // Changed to s.budgetRepo
}

// GetBudgetUsage calculates the usage of a budget
func (s *BudgetService) GetBudgetUsage(ctx context.Context, budgetID int32, userID int32) (*domain.BudgetUsage, error) {
    budget, err := s.GetBudget(ctx, budgetID, userID) // This also handles ownership check
    if err != nil {
        return nil, err
    }

    // The repository method is GetLedgersByUserIDAndDateRangeAndCategory
    // It expects UserID, StartDate, EndDate, CategoryID.
    // The current call structure matches the implemented repository method.
    ledgers, err := s.ledgerRepo.GetLedgersByUserIDAndDateRangeAndCategory( // Changed to s.ledgerRepo
        ctx,
        userID, // This is the UserID of the user requesting the budget usage
        budget.StartDate,
        budget.EndDate,
        budget.CategoryID, // This is the CategoryID from the budget
    )
    if err != nil {
        return nil, err
    }

    var spentAmount decimal.Decimal
    for _, ledger := range ledgers {
        // Assuming 'expense' and 'transfer' out count towards budget spending.
        // This logic might need refinement based on exact requirements.
        if ledger.Type == domain.LedgerTypeExpense || (ledger.Type == domain.LedgerTypeTransfer) { // Need to clarify how transfers affect budget
            spentAmount = spentAmount.Add(ledger.Amount)
        }
    }
    
    // Ensure domain.BudgetUsage is defined
    return &domain.BudgetUsage{
        BudgetID:      budget.ID,
        BudgetName:    budget.Name,
        BudgetAmount:  budget.Amount,
        SpentAmount:   spentAmount,
        RemainingAmount: budget.Amount.Sub(spentAmount),
        Period:        budget.Period,
        StartDate:     budget.StartDate,
        EndDate:       budget.EndDate,
        CategoryID:    budget.CategoryID,
    }, nil
}
