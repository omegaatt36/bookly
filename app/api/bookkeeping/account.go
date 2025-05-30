package bookkeeping

import (
	"errors"
	"net/http"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/sdk/datatype"
)

func convertAccountFromDomain(account *domain.Account) datatype.Account {
	return datatype.Account{
		ID:        account.ID,
		CreatedAt: account.CreatedAt.Format(datatype.TimeFormat),
		UpdatedAt: account.UpdatedAt.Format(datatype.TimeFormat),
		Name:      account.Name,
		Status:    account.Status.String(),
		Currency:  account.Currency,
		Balance:   account.Balance.String(),
	}
}

// CreateAccount handles the creation of a new account
func (x *Controller) CreateAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var req datatype.CreateAccountRequest
		engine.Chain(r, w, func(ctx *engine.Context, req datatype.CreateAccountRequest) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			if req.Name == "" {
				return nil, app.ParamError(errors.New("name is required"))
			}

			if req.Currency == "" {
				return nil, app.ParamError(errors.New("currency is required"))
			}

			return nil, x.service.CreateAccount(domain.CreateAccountRequest{
				UserID:   userID,
				Name:     req.Name,
				Currency: req.Currency,
			})
		}).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// GetAllAccounts handles the retrieval of all accounts for the current authenticated user
func (x *Controller) GetAllAccounts() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]datatype.Account, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			accounts, err := x.service.GetAccountsByUserID(userID)
			if err != nil {
				return nil, err
			}

			jsonAccounts := make([]datatype.Account, len(accounts))
			for index, account := range accounts {
				jsonAccounts[index] = convertAccountFromDomain(account)
			}

			return jsonAccounts, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetAccountByID handles the retrieval of an account by its ID for the current authenticated user
func (x *Controller) GetAccountByID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*datatype.Account, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			account, err := x.service.GetAccountByID(id)
			if err != nil {
				return nil, err
			}

			// Verify account ownership
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			jsonAccount := convertAccountFromDomain(account)
			return &jsonAccount, nil
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}

// UpdateAccount handles the updating of an account
func (x *Controller) UpdateAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id int32
			datatype.UpdateAccountRequest
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(req.id)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

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
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(id)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			return nil, x.service.DeactivateAccountByID(id)
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetUserAccounts handles the retrieval of all accounts for a specific user
func (x *Controller) GetUserAccounts() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var userID int32
		engine.Chain(r, w, func(_ *engine.Context, _ *engine.Empty) ([]datatype.Account, error) {
			accounts, err := x.service.GetAccountsByUserID(userID)
			if err != nil {
				return nil, err
			}

			jsonAccounts := make([]datatype.Account, len(accounts))
			for index, account := range accounts {
				jsonAccounts[index] = convertAccountFromDomain(account)
			}

			return jsonAccounts, nil
		}).Param("user_id", &userID).Call(&engine.Empty{}).ResponseJSON()
	}
}

// CreateUserAccount handles the creation of a new account for a specific user (admin only)
func (x *Controller) CreateUserAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			userID int32
			datatype.CreateAccountForUserRequest
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			// Admin validation would go here
			// For now we'll just use the authenticated user's ID
			authenticatedUserID := ctx.GetUserID()
			if authenticatedUserID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			if req.Name == "" {
				return nil, app.ParamError(errors.New("name is required"))
			}

			if req.Currency == "" {
				return nil, app.ParamError(errors.New("currency is required"))
			}

			// In a real admin scenario, we'd use req.UserID, but for now validate that it matches authenticated user
			if req.userID != authenticatedUserID {
				return nil, app.Forbidden(errors.New("can only create accounts for yourself"))
			}

			return nil, x.service.CreateAccount(domain.CreateAccountRequest{
				UserID:   req.userID,
				Name:     req.Name,
				Currency: req.Currency,
			})
		}).Param("user_id", &req.userID).BindJSON(&req).Call(req).ResponseJSON()
	}
}
