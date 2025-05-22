package sqlc

import (
	"context"
	// "database/sql" // No longer needed directly for sql.DB

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
	"github.com/shopspring/decimal"
)

// BudgetRepository implements domain.BudgetRepository
type BudgetRepository struct {
	querier sqlcgen.Querier
	db      sqlcgen.DBTX // Changed from *sql.DB to sqlcgen.DBTX
}

// NewBudgetRepository creates a new BudgetRepository
func NewBudgetRepository(db sqlcgen.DBTX, querier sqlcgen.Querier) domain.BudgetRepository { // Changed parameter type
	return &BudgetRepository{
		querier: querier,
		db:      db,
	}
}

// CreateBudget creates a new budget
func (r *BudgetRepository) CreateBudget(ctx context.Context, req domain.CreateBudgetRequest) (int32, error) {
	// Calculate EndDate based on StartDate and Period
	endDate := req.StartDate
	if req.Period == domain.BudgetPeriodMonthly {
		endDate = req.StartDate.AddDate(0, 1, -1)
	} else if req.Period == domain.BudgetPeriodYearly {
		endDate = req.StartDate.AddDate(1, 0, -1)
	}

	dbBudget, err := r.querier.CreateBudget(ctx, r.db, sqlcgen.CreateBudgetParams{
		UserID:     req.UserID,
		Name:       req.Name,
		Period:     req.Period.String(),
		StartDate:  req.StartDate,
		EndDate:    endDate,
		Amount:     req.Amount.String(), // sqlc expects string for decimal
		CategoryID: req.CategoryID,
	})
	if err != nil {
		return 0, err
	}
	return dbBudget.ID, nil
}

// GetBudgetByID retrieves a budget by its ID
func (r *BudgetRepository) GetBudgetByID(ctx context.Context, id int32) (*domain.Budget, error) {
	dbBudget, err := r.querier.GetBudgetByID(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	amount, _ := decimal.NewFromString(dbBudget.Amount)
	return &domain.Budget{
		ID:         dbBudget.ID,
		CreatedAt:  dbBudget.CreatedAt,
		UpdatedAt:  dbBudget.UpdatedAt,
		UserID:     dbBudget.UserID,
		Name:       dbBudget.Name,
		Period:     domain.BudgetPeriod(dbBudget.Period),
		StartDate:  dbBudget.StartDate,
		EndDate:    dbBudget.EndDate,
		Amount:     amount,
		CategoryID: dbBudget.CategoryID,
	}, nil
}

// GetBudgetsByUserID retrieves budgets by user ID
func (r *BudgetRepository) GetBudgetsByUserID(ctx context.Context, userID int32) ([]*domain.Budget, error) {
	dbBudgets, err := r.querier.ListBudgetsByUserID(ctx, r.db, userID)
	if err != nil {
		return nil, err
	}
	budgets := make([]*domain.Budget, len(dbBudgets))
	for i, dbBudget := range dbBudgets {
		amount, _ := decimal.NewFromString(dbBudget.Amount)
		budgets[i] = &domain.Budget{
			ID:         dbBudget.ID,
			CreatedAt:  dbBudget.CreatedAt,
			UpdatedAt:  dbBudget.UpdatedAt,
			UserID:     dbBudget.UserID,
			Name:       dbBudget.Name,
			Period:     domain.BudgetPeriod(dbBudget.Period),
			StartDate:  dbBudget.StartDate,
			EndDate:    dbBudget.EndDate,
			Amount:     amount,
			CategoryID: dbBudget.CategoryID,
		}
	}
	return budgets, nil
}

// UpdateBudget updates a budget
func (r *BudgetRepository) UpdateBudget(ctx context.Context, req domain.UpdateBudgetRequest) error {
	// Fetch the existing budget to get current values for StartDate and Period if not provided
	// This is important for recalculating EndDate if StartDate or Period changes
	existingBudget, err := r.GetBudgetByID(ctx, req.ID)
	if err != nil {
		return err // Budget not found or other error
	}

	// Use existing values if new ones are not provided
	name := existingBudget.Name
	if req.Name != nil {
		name = *req.Name
	}
	period := existingBudget.Period
	if req.Period != nil {
		period = *req.Period
	}
	startDate := existingBudget.StartDate
	if req.StartDate != nil {
		startDate = *req.StartDate
	}
	amount := existingBudget.Amount
	if req.Amount != nil {
		amount = *req.Amount
	}
	categoryID := existingBudget.CategoryID
	if req.CategoryID != nil {
		categoryID = *req.CategoryID
	}

	// Recalculate EndDate
	endDate := startDate
	if period == domain.BudgetPeriodMonthly {
		endDate = startDate.AddDate(0, 1, -1)
	} else if period == domain.BudgetPeriodYearly {
		endDate = startDate.AddDate(1, 0, -1)
	}

	_, err = r.querier.UpdateBudget(ctx, r.db, sqlcgen.UpdateBudgetParams{
		ID:         req.ID,
		Name:       name,
		Period:     period.String(),
		StartDate:  startDate,
		EndDate:    endDate,
		Amount:     amount.String(),
		CategoryID: categoryID,
	})
	return err
}

// DeleteBudget deletes a budget by its ID
func (r *BudgetRepository) DeleteBudget(ctx context.Context, id int32) error {
	return r.querier.DeleteBudget(ctx, r.db, id)
}
