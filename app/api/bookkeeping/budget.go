package bookkeeping

import (
	"errors"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
)

type jsonBudget struct {
	ID         int32  `json:"id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	UserID     int32  `json:"user_id"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	Amount     string `json:"amount"`
	PeriodType string `json:"period_type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date,omitempty"`
	IsActive   bool   `json:"is_active"`
}

func (b *jsonBudget) fromDomain(budget *domain.Budget) {
	b.ID = budget.ID
	b.CreatedAt = budget.CreatedAt.Format(time.RFC3339)
	b.UpdatedAt = budget.UpdatedAt.Format(time.RFC3339)
	b.UserID = budget.UserID
	b.Name = budget.Name
	b.Category = budget.Category
	b.Amount = budget.Amount.String()
	b.PeriodType = string(budget.PeriodType)
	b.StartDate = budget.StartDate.Format(time.RFC3339)
	if budget.EndDate != nil {
		b.EndDate = budget.EndDate.Format(time.RFC3339)
	}
	b.IsActive = budget.IsActive
}

type jsonBudgetSummary struct {
	Budget      jsonBudget `json:"budget"`
	UsedAmount  string     `json:"used_amount"`
	Percentage  string     `json:"percentage"`
	PeriodStart string     `json:"period_start"`
	PeriodEnd   string     `json:"period_end"`
}

// CreateBudget creates a new budget
func (x *Controller) CreateBudget() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Name       string          `json:"name"`
			Category   string          `json:"category"`
			Amount     decimal.Decimal `json:"amount"`
			PeriodType string          `json:"period_type"`
			StartDate  time.Time       `json:"start_date"`
			EndDate    *time.Time      `json:"end_date"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*jsonBudget, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			if req.Name == "" {
				return nil, app.ParamError(errors.New("name is required"))
			}

			if req.Category == "" {
				return nil, app.ParamError(errors.New("category is required"))
			}

			if req.Amount.IsZero() || req.Amount.IsNegative() {
				return nil, app.ParamError(errors.New("amount must be greater than 0"))
			}

			periodType, err := domain.ParsePeriodType(req.PeriodType)
			if err != nil {
				return nil, app.ParamError(err)
			}

			budgetID, err := x.budgetRepo.CreateBudget(domain.CreateBudgetRequest{
				UserID:     userID,
				Name:       req.Name,
				Category:   req.Category,
				Amount:     req.Amount,
				PeriodType: periodType,
				StartDate:  req.StartDate,
				EndDate:    req.EndDate,
			})
			if err != nil {
				return nil, err
			}

			budget, err := x.budgetRepo.GetBudgetByID(budgetID)
			if err != nil {
				return nil, err
			}

			resp := &jsonBudget{}
			resp.fromDomain(budget)
			return resp, nil
		}).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// GetBudget gets a budget by ID
func (x *Controller) GetBudget() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id int32
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*jsonBudget, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			budget, err := x.budgetRepo.GetBudgetByID(req.id)
			if err != nil {
				if err == domain.ErrBudgetNotFound {
					return nil, app.NotFoundError()
				}
				return nil, err
			}

			if budget.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied"))
			}

			resp := &jsonBudget{}
			resp.fromDomain(budget)
			return resp, nil
		}).Param("id", &req.id).Call(req).ResponseJSON()
	}
}

// GetBudgets gets all budgets for the current user
func (x *Controller) GetBudgets() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			activeOnly bool
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) ([]*jsonBudget, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			var budgets []*domain.Budget
			var err error

			if req.activeOnly {
				budgets, err = x.budgetRepo.GetActiveBudgetsByUserID(userID)
			} else {
				budgets, err = x.budgetRepo.GetBudgetsByUserID(userID)
			}

			if err != nil {
				return nil, err
			}

			resp := make([]*jsonBudget, len(budgets))
			for i, budget := range budgets {
				resp[i] = &jsonBudget{}
				resp[i].fromDomain(budget)
			}

			return resp, nil
		}).Query("active_only", &req.activeOnly).Call(req).ResponseJSON()
	}
}

// UpdateBudget updates a budget
func (x *Controller) UpdateBudget() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id         int32
			Name       *string          `json:"name"`
			Category   *string          `json:"category"`
			Amount     *decimal.Decimal `json:"amount"`
			PeriodType *string          `json:"period_type"`
			StartDate  *time.Time       `json:"start_date"`
			EndDate    *time.Time       `json:"end_date"`
			IsActive   *bool            `json:"is_active"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*jsonBudget, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Get existing budget to check ownership
			budget, err := x.budgetRepo.GetBudgetByID(req.id)
			if err != nil {
				if err == domain.ErrBudgetNotFound {
					return nil, app.NotFoundError()
				}
				return nil, err
			}

			if budget.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied"))
			}

			// Validate period type if provided
			var periodType *domain.PeriodType
			if req.PeriodType != nil {
				pt, err := domain.ParsePeriodType(*req.PeriodType)
				if err != nil {
					return nil, app.ParamError(err)
				}
				periodType = &pt
			}

			// Validate amount if provided
			if req.Amount != nil && (req.Amount.IsZero() || req.Amount.IsNegative()) {
				return nil, app.ParamError(errors.New("amount must be greater than 0"))
			}

			err = x.budgetRepo.UpdateBudget(domain.UpdateBudgetRequest{
				ID:         req.id,
				Name:       req.Name,
				Category:   req.Category,
				Amount:     req.Amount,
				PeriodType: periodType,
				StartDate:  req.StartDate,
				EndDate:    req.EndDate,
				IsActive:   req.IsActive,
			})
			if err != nil {
				return nil, err
			}

			// Get updated budget
			updated, err := x.budgetRepo.GetBudgetByID(req.id)
			if err != nil {
				return nil, err
			}

			resp := &jsonBudget{}
			resp.fromDomain(updated)
			return resp, nil
		}).Param("id", &req.id).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// DeleteBudget deletes a budget
func (x *Controller) DeleteBudget() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id int32
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Get existing budget to check ownership
			budget, err := x.budgetRepo.GetBudgetByID(req.id)
			if err != nil {
				if err == domain.ErrBudgetNotFound {
					return nil, app.NotFoundError()
				}
				return nil, err
			}

			if budget.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied"))
			}

			err = x.budgetRepo.DeleteBudget(req.id)
			if err != nil {
				return nil, err
			}

			return nil, nil
		}).Param("id", &req.id).Call(req).ResponseJSON()
	}
}

// GetBudgetSummary gets budget summary with usage
func (x *Controller) GetBudgetSummary() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id int32
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*jsonBudgetSummary, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Get budget
			budget, err := x.budgetRepo.GetBudgetByID(req.id)
			if err != nil {
				if err == domain.ErrBudgetNotFound {
					return nil, app.NotFoundError()
				}
				return nil, err
			}

			if budget.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied"))
			}

			// Calculate period dates
			now := time.Now()
			var periodStart, periodEnd time.Time

			switch budget.PeriodType {
			case domain.PeriodTypeMonthly:
				year, month, _ := now.Date()
				periodStart = time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
				periodEnd = periodStart.AddDate(0, 1, 0).Add(-time.Second)
			case domain.PeriodTypeYearly:
				year := now.Year()
				periodStart = time.Date(year, 1, 1, 0, 0, 0, 0, now.Location())
				periodEnd = time.Date(year, 12, 31, 23, 59, 59, 999999999, now.Location())
			}

			// Get usage
			usage, err := x.budgetRepo.GetBudgetUsage(userID, budget.Category, periodStart, periodEnd)
			if err != nil {
				return nil, err
			}

			// Calculate percentage
			percentage := decimal.Zero
			if budget.Amount.GreaterThan(decimal.Zero) {
				percentage = usage.Div(budget.Amount).Mul(decimal.NewFromInt(100))
			}

			budgetResp := &jsonBudget{}
			budgetResp.fromDomain(budget)

			return &jsonBudgetSummary{
				Budget:      *budgetResp,
				UsedAmount:  usage.String(),
				Percentage:  percentage.String(),
				PeriodStart: periodStart.Format(time.RFC3339),
				PeriodEnd:   periodEnd.Format(time.RFC3339),
			}, nil
		}).Param("id", &req.id).Call(req).ResponseJSON()
	}
}