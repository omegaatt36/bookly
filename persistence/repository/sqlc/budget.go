package sqlc

import (
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
)

// CreateBudget creates a new budget
func (r *Repository) CreateBudget(req domain.CreateBudgetRequest) (int32, error) {
	budget, err := r.querier.CreateBudget(r.ctx, sqlcgen.CreateBudgetParams{
		UserID:     req.UserID,
		Name:       req.Name,
		Category:   req.Category,
		Amount:     req.Amount,
		PeriodType: string(req.PeriodType),
		StartDate:  pgtype.Timestamptz{Time: req.StartDate, Valid: true},
		EndDate:    timePtrToPgtype(req.EndDate),
		IsActive:   true,
	})
	if err != nil {
		return 0, err
	}
	return budget.ID, nil
}

// GetBudgetByID gets a budget by ID
func (r *Repository) GetBudgetByID(id int32) (*domain.Budget, error) {
	budget, err := r.querier.GetBudgetByID(r.ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrBudgetNotFound
		}
		return nil, err
	}
	return r.toDomainBudget(budget), nil
}

// GetBudgetsByUserID gets all budgets for a user
func (r *Repository) GetBudgetsByUserID(userID int32) ([]*domain.Budget, error) {
	budgets, err := r.querier.GetBudgetsByUserID(r.ctx, userID)
	if err != nil {
		return nil, err
	}
	return r.toDomainBudgets(budgets), nil
}

// GetActiveBudgetsByUserID gets all active budgets for a user
func (r *Repository) GetActiveBudgetsByUserID(userID int32) ([]*domain.Budget, error) {
	budgets, err := r.querier.GetActiveBudgetsByUserID(r.ctx, userID)
	if err != nil {
		return nil, err
	}
	return r.toDomainBudgets(budgets), nil
}

// GetBudgetsByUserIDAndCategory gets budgets by user ID and category
func (r *Repository) GetBudgetsByUserIDAndCategory(userID int32, category string) ([]*domain.Budget, error) {
	budgets, err := r.querier.GetBudgetsByUserIDAndCategory(r.ctx, sqlcgen.GetBudgetsByUserIDAndCategoryParams{
		UserID:   userID,
		Category: category,
	})
	if err != nil {
		return nil, err
	}
	return r.toDomainBudgets(budgets), nil
}

// GetActiveBudgetByUserIDCategoryAndPeriod gets active budget by user ID, category and period
func (r *Repository) GetActiveBudgetByUserIDCategoryAndPeriod(userID int32, category string, periodType domain.PeriodType, date time.Time) (*domain.Budget, error) {
	budget, err := r.querier.GetActiveBudgetByUserIDCategoryAndPeriod(r.ctx, sqlcgen.GetActiveBudgetByUserIDCategoryAndPeriodParams{
		UserID:     userID,
		Category:   category,
		PeriodType: string(periodType),
		StartDate:  pgtype.Timestamptz{Time: date, Valid: true},
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomainBudget(budget), nil
}

// UpdateBudget updates a budget
func (r *Repository) UpdateBudget(req domain.UpdateBudgetRequest) error {
	// Get the current budget first
	budget, err := r.querier.GetBudgetByID(r.ctx, req.ID)
	if err != nil {
		return err
	}

	params := sqlcgen.UpdateBudgetParams{
		ID:         req.ID,
		Name:       budget.Name,
		Category:   budget.Category,
		Amount:     budget.Amount,
		PeriodType: budget.PeriodType,
		StartDate:  budget.StartDate,
		EndDate:    budget.EndDate,
		IsActive:   budget.IsActive,
	}

	// Update only the fields that are provided
	if req.Name != nil {
		params.Name = *req.Name
	}
	if req.Category != nil {
		params.Category = *req.Category
	}
	if req.Amount != nil {
		params.Amount = *req.Amount
	}
	if req.PeriodType != nil {
		params.PeriodType = string(*req.PeriodType)
	}
	if req.StartDate != nil {
		params.StartDate = pgtype.Timestamptz{Time: *req.StartDate, Valid: true}
	}
	if req.EndDate != nil {
		params.EndDate = timePtrToPgtype(req.EndDate)
	}
	if req.IsActive != nil {
		params.IsActive = *req.IsActive
	}

	_, err = r.querier.UpdateBudget(r.ctx, params)
	return err
}

// DeleteBudget deletes a budget
func (r *Repository) DeleteBudget(id int32) error {
	return r.querier.DeleteBudget(r.ctx, id)
}

// GetBudgetUsage gets the total amount used for a category in a date range
func (r *Repository) GetBudgetUsage(userID int32, category string, startDate, endDate time.Time) (decimal.Decimal, error) {
	result, err := r.querier.GetLedgersSumByCategoryAndDateRange(r.ctx, sqlcgen.GetLedgersSumByCategoryAndDateRangeParams{
		UserID:   userID,
		Category: pgtype.Text{String: category, Valid: true},
		Date:     pgtype.Timestamptz{Time: startDate, Valid: true},
		Date_2:   pgtype.Timestamptz{Time: endDate, Valid: true},
	})
	if err != nil {
		return decimal.Zero, err
	}
	return result, nil
}

// toDomainBudget converts a database budget to domain budget
func (r *Repository) toDomainBudget(budget sqlcgen.Budget) *domain.Budget {
	return &domain.Budget{
		ID:         budget.ID,
		CreatedAt:  budget.CreatedAt.Time,
		UpdatedAt:  budget.UpdatedAt.Time,
		UserID:     budget.UserID,
		Name:       budget.Name,
		Category:   budget.Category,
		Amount:     budget.Amount,
		PeriodType: domain.PeriodType(budget.PeriodType),
		StartDate:  budget.StartDate.Time,
		EndDate:    pgtypeToTimePtr(budget.EndDate),
		IsActive:   budget.IsActive,
	}
}

// toDomainBudgets converts database budgets to domain budgets
func (r *Repository) toDomainBudgets(budgets []sqlcgen.Budget) []*domain.Budget {
	result := make([]*domain.Budget, len(budgets))
	for i, budget := range budgets {
		result[i] = r.toDomainBudget(budget)
	}
	return result
}

// Helper functions
func timePtrToPgtype(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func pgtypeToTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}
