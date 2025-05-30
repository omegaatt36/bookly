package bookkeeping

import (
	"errors"
	"net/http"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/sdk/datatype"
)

func convertBankAccountFromDomain(bankAccount *domain.BankAccount) datatype.BankAccount {
	return datatype.BankAccount{
		ID:            bankAccount.ID,
		CreatedAt:     bankAccount.CreatedAt.Format(datatype.TimeFormat),
		UpdatedAt:     bankAccount.UpdatedAt.Format(datatype.TimeFormat),
		AccountID:     bankAccount.AccountID,
		AccountNumber: bankAccount.AccountNumber,
		BankName:      bankAccount.BankName,
		BranchName:    bankAccount.BranchName,
		SwiftCode:     bankAccount.SwiftCode,
	}
}

// CreateBankAccount handles the creation of a new bank account
func (x *Controller) CreateBankAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var req datatype.CreateBankAccountRequest
		engine.Chain(r, w, func(ctx *engine.Context, req datatype.CreateBankAccountRequest) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(req.AccountID)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			if req.AccountNumber == "" {
				return nil, app.ParamError(errors.New("account number is required"))
			}

			if req.BankName == "" {
				return nil, app.ParamError(errors.New("bank name is required"))
			}

			return nil, x.service.CreateBankAccount(domain.CreateBankAccountRequest{
				AccountID:     req.AccountID,
				AccountNumber: req.AccountNumber,
				BankName:      req.BankName,
				BranchName:    req.BranchName,
				SwiftCode:     req.SwiftCode,
			})
		}).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// GetBankAccountByID handles the retrieval of a bank account by its ID
func (x *Controller) GetBankAccountByID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*datatype.BankAccount, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			bankAccount, err := x.service.GetBankAccountByID(id)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				return nil, err
			}

			// Verify ownership by checking the associated account
			account, err := x.service.GetAccountByID(bankAccount.AccountID)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			jsonBankAccount := convertBankAccountFromDomain(bankAccount)

			return &jsonBankAccount, nil
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}

// GetBankAccountByAccountID handles the retrieval of a bank account by its associated account ID
func (x *Controller) GetBankAccountByAccountID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var accountID int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*datatype.BankAccount, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(accountID)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			bankAccount, err := x.service.GetBankAccountByAccountID(accountID)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				return nil, err
			}

			jsonBankAccount := convertBankAccountFromDomain(bankAccount)
			return &jsonBankAccount, nil
		}).Param("account_id", &accountID).Call(nil).ResponseJSON()
	}
}

// UpdateBankAccount handles the updating of a bank account
func (x *Controller) UpdateBankAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id int32
			datatype.UpdateBankAccountRequest
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Get the bank account
			bankAccount, err := x.service.GetBankAccountByID(req.id)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				return nil, err
			}

			// Verify ownership by checking the associated account
			account, err := x.service.GetAccountByID(bankAccount.AccountID)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			return nil, x.service.UpdateBankAccount(domain.UpdateBankAccountRequest{
				ID:            req.id,
				AccountNumber: req.AccountNumber,
				BankName:      req.BankName,
				BranchName:    req.BranchName,
				SwiftCode:     req.SwiftCode,
			})
		}).Param("id", &req.id).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// DeleteBankAccount handles the deletion of a bank account by its ID
func (x *Controller) DeleteBankAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Get the bank account
			bankAccount, err := x.service.GetBankAccountByID(id)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return nil, app.NotFoundError()
				}
				return nil, err
			}

			// Verify ownership by checking the associated account
			account, err := x.service.GetAccountByID(bankAccount.AccountID)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			return nil, x.service.DeleteBankAccount(id)
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}
