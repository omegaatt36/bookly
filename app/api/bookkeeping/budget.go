package bookkeeping

import (
	"errors"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
	"github.com/shopspring/decimal"
)

type jsonBudget struct {
	ID         int32           `json:"id"`
	Name       string          `json:"name"`
	Period     string          `json:"period"`
	StartDate  time.Time       `json:"start_date"`
	EndDate    time.Time       `json:"end_date"`
	Amount     decimal.Decimal `json:"amount"`
	CategoryID int32           `json:"category_id"`
	UserID     int32           `json:"user_id"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type jsonBudgetUsage struct {
    BudgetID        int32           `json:"budget_id"`
    BudgetName      string          `json:"budget_name"`
    BudgetAmount    decimal.Decimal `json:"budget_amount"`
    SpentAmount     decimal.Decimal `json:"spent_amount"`
    RemainingAmount decimal.Decimal `json:"remaining_amount"`
    Period          string          `json:"period"`
    StartDate       time.Time       `json:"start_date"`
    EndDate         time.Time       `json:"end_date"`
    CategoryID      int32           `json:"category_id"`
}

func (jb *jsonBudget) fromDomain(budget *domain.Budget) {
	jb.ID = budget.ID
	jb.Name = budget.Name
	jb.Period = budget.Period.String()
	jb.StartDate = budget.StartDate
	jb.EndDate = budget.EndDate
	jb.Amount = budget.Amount
	jb.CategoryID = budget.CategoryID
	jb.UserID = budget.UserID
	jb.CreatedAt = budget.CreatedAt
	jb.UpdatedAt = budget.UpdatedAt
}

func (jbu *jsonBudgetUsage) fromDomain(usage *domain.BudgetUsage) {
    jbu.BudgetID = usage.BudgetID
    jbu.BudgetName = usage.BudgetName
    jbu.BudgetAmount = usage.BudgetAmount
    jbu.SpentAmount = usage.SpentAmount
    jbu.RemainingAmount = usage.RemainingAmount
    jbu.Period = usage.Period.String()
    jbu.StartDate = usage.StartDate
    jbu.EndDate = usage.EndDate
    jbu.CategoryID = usage.CategoryID
}

// CreateBudget handles the creation of a new budget
func (x *Controller) CreateBudget() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Name       string          `json:"name"`
			Period     string          `json:"period"` // "monthly", "yearly"
			StartDate  time.Time       `json:"start_date"` // Optional, defaults in service
			Amount     decimal.Decimal `json:"amount"`
			CategoryID int32           `json:"category_id"`
		}
		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*jsonBudget, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}
			
			period, err := domain.ParseBudgetPeriod(req.Period)
			if err != nil {
				return nil, app.ParamError(errors.New("invalid period type"))
			}

			budgetID, err := x.budgetService.CreateBudget(ctx, domain.CreateBudgetRequest{
				UserID:     userID,
				Name:       req.Name,
				Period:     period,
				StartDate:  req.StartDate, // Service will default if zero
				Amount:     req.Amount,
				CategoryID: req.CategoryID,
			})
			if err != nil {
				return nil, err
			}
			createdBudget, err := x.budgetService.GetBudget(ctx, budgetID, userID)
			if err != nil {
			    return nil, err // Should ideally not happen if create succeeded
			}
			var res jsonBudget
			res.fromDomain(createdBudget)
			return &res, nil
		}).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// ListBudgets handles listing all budgets for the authenticated user
func (x *Controller) ListBudgets() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]jsonBudget, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}
			budgets, err := x.budgetService.ListBudgets(ctx, userID)
			if err != nil {
				return nil, err
			}
			jsonBudgets := make([]jsonBudget, len(budgets))
			for i, b := range budgets {
				jsonBudgets[i].fromDomain(b)
			}
			return jsonBudgets, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetBudgetByID handles retrieving a specific budget by its ID
func (x *Controller) GetBudgetByID() func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        var budgetID int32
        engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*jsonBudget, error) {
            userID := ctx.GetUserID()
            if userID == 0 {
                return nil, app.Unauthorized(errors.New("user not authenticated"))
            }
            budget, err := x.budgetService.GetBudget(ctx, budgetID, userID)
            if err != nil {
                return nil, err
            }
            var res jsonBudget
            res.fromDomain(budget)
            return &res, nil
        }).Param("budget_id", &budgetID).Call(&engine.Empty{}).ResponseJSON()
    }
}

// GetBudgetUsageByID handles retrieving usage for a specific budget
func (x *Controller) GetBudgetUsageByID() func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        var budgetID int32
        engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*jsonBudgetUsage, error) {
            userID := ctx.GetUserID()
            if userID == 0 {
                return nil, app.Unauthorized(errors.New("user not authenticated"))
            }
            usage, err := x.budgetService.GetBudgetUsage(ctx, budgetID, userID)
            if err != nil {
                return nil, err
            }
            var res jsonBudgetUsage
            res.fromDomain(usage)
            return &res, nil
        }).Param("budget_id", &budgetID).Call(&engine.Empty{}).ResponseJSON()
    }
}

// UpdateBudget handles updating a budget
func (x *Controller) UpdateBudget() func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        type request struct {
            Name       *string          `json:"name"`
            Period     *string          `json:"period"`
            StartDate  *time.Time       `json:"start_date"`
            Amount     *decimal.Decimal `json:"amount"`
            CategoryID *int32           `json:"category_id"`
        }
        var req request
        var budgetID int32
        engine.Chain(r, w, func(ctx *engine.Context, req request) (*jsonBudget, error) {
            userID := ctx.GetUserID()
            if userID == 0 {
                return nil, app.Unauthorized(errors.New("user not authenticated"))
            }

            var domainPeriod *domain.BudgetPeriod
            if req.Period != nil {
                p, err := domain.ParseBudgetPeriod(*req.Period)
                if err != nil {
                    return nil, app.ParamError(errors.New("invalid period type"))
                }
                domainPeriod = &p
            }

            err := x.budgetService.UpdateBudget(ctx, domain.UpdateBudgetRequest{
                ID:         budgetID,
                Name:       req.Name,
                Period:     domainPeriod,
                StartDate:  req.StartDate,
                Amount:     req.Amount,
                CategoryID: req.CategoryID,
            }, userID)
            if err != nil {
                return nil, err
            }
            
            updatedBudget, err := x.budgetService.GetBudget(ctx, budgetID, userID)
            if err != nil {
                return nil, err 
            }
            var res jsonBudget
            res.fromDomain(updatedBudget)
            return &res, nil
        }).Param("budget_id", &budgetID).BindJSON(&req).Call(req).ResponseJSON()
    }
}

// DeleteBudget handles deleting a budget
func (x *Controller) DeleteBudget() func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        var budgetID int32
        engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*engine.Empty, error) {
            userID := ctx.GetUserID()
            if userID == 0 {
                return nil, app.Unauthorized(errors.New("user not authenticated"))
            }
            err := x.budgetService.DeleteBudget(ctx, budgetID, userID)
            return nil, err
        }).Param("budget_id", &budgetID).Call(&engine.Empty{}).ResponseJSON()
    }
}
