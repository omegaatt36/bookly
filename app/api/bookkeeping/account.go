package bookkeeping

import (
	"errors"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
)

type jsonAccount struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Currency  string `json:"currency"`
	Balance   string `json:"balance"`
}

func (r *jsonAccount) fromDomain(account *domain.Account) {
	r.ID = account.ID
	r.CreatedAt = account.CreatedAt.Format(time.RFC3339)
	r.UpdatedAt = account.UpdatedAt.Format(time.RFC3339)
	r.Name = account.Name
	r.Status = account.Status.String()
	r.Currency = account.Currency
	r.Balance = account.Balance.String()

}

// CreateAccount handles the creation of a new account
func (x *Controller) CreateAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			UserID   string `json:"user_id"`
			Name     string `json:"name"`
			Currency string `json:"currency"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			if req.UserID == "" {
				return nil, app.ParamError(errors.New("user_id is required"))
			}

			if req.Name == "" {
				return nil, app.ParamError(errors.New("name is required"))
			}

			if req.Currency == "" {
				return nil, app.ParamError(errors.New("currency is required"))
			}

			return nil, x.service.CreateAccount(domain.CreateAccountRequest{
				UserID:   req.UserID,
				Name:     req.Name,
				Currency: req.Currency,
			})
		}).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// GetAllAccounts handles the retrieval of all accounts
func (x *Controller) GetAllAccounts() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]jsonAccount, error) {
			accounts, err := x.service.GetAllAccounts()
			if err != nil {
				return nil, err
			}

			jsonAccounts := make([]jsonAccount, len(accounts))
			for index, account := range accounts {
				jsonAccounts[index].fromDomain(account)
			}

			return jsonAccounts, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetAccountByID handles the retrieval of an account by its ID
func (x *Controller) GetAccountByID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id string
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*jsonAccount, error) {
			account, err := x.service.GetAccountByID(id)
			if err != nil {
				return nil, err
			}

			var jsonAccount jsonAccount
			jsonAccount.fromDomain(account)

			return &jsonAccount, nil
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}

// UpdateAccount handles the updating of an account
func (x *Controller) UpdateAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id     string
			Name   *string `json:"name"`
			Status *string `json:"status"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {

			var accountStatus *domain.AccountStatus
			if req.Status != nil {
				status, err := domain.ParseAccountStatus(*req.Status)
				if err != nil {
					return nil, errors.New("invalid account status")
				}
				accountStatus = &status
			}

			return nil, x.service.UpdateAccount(domain.UpdateAccountRequest{
				ID:     req.id,
				Name:   req.Name,
				Status: accountStatus,
			})
		}).Param("id", &req.id).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// DeactivateAccountByID handles the deactivation of an account by its ID
func (x *Controller) DeactivateAccountByID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id string
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*engine.Empty, error) {
			return nil, x.service.DeactivateAccountByID(id)
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetUserAccounts handles the retrieval of all accounts for a specific user
func (x *Controller) GetUserAccounts() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var userID string
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]jsonAccount, error) {
			accounts, err := x.service.GetAccountsByUserID(userID)
			if err != nil {
				return nil, err
			}

			jsonAccounts := make([]jsonAccount, len(accounts))
			for index, account := range accounts {
				jsonAccounts[index].fromDomain(account)
			}

			return jsonAccounts, nil
		}).Param("user_id", &userID).Call(&engine.Empty{}).ResponseJSON()
	}
}

// CreateUserAccount handles the creation of a new account for a specific user
func (x *Controller) CreateUserAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			UserID   string `json:"-"`
			Name     string `json:"name"`
			Currency string `json:"currency"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			if req.Name == "" {
				return nil, app.ParamError(errors.New("name is required"))
			}

			if req.Currency == "" {
				return nil, app.ParamError(errors.New("currency is required"))
			}

			return nil, x.service.CreateAccount(domain.CreateAccountRequest{
				UserID:   req.UserID,
				Name:     req.Name,
				Currency: req.Currency,
			})
		}).Param("user_id", &req.UserID).BindJSON(&req).Call(req).ResponseJSON()
	}
}
