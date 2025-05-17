package bookkeeping

import (
	"errors"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
)

type jsonBankAccount struct {
	ID            int32  `json:"id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	AccountID     int32  `json:"account_id"`
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
	BranchName    string `json:"branch_name,omitempty"`
	SwiftCode     string `json:"swift_code,omitempty"`
}

func (r *jsonBankAccount) fromDomain(bankAccount *domain.BankAccount) {
	r.ID = bankAccount.ID
	r.CreatedAt = bankAccount.CreatedAt.Format(time.RFC3339)
	r.UpdatedAt = bankAccount.UpdatedAt.Format(time.RFC3339)
	r.AccountID = bankAccount.AccountID
	r.AccountNumber = bankAccount.AccountNumber
	r.BankName = bankAccount.BankName
	r.BranchName = bankAccount.BranchName
	r.SwiftCode = bankAccount.SwiftCode
}

// CreateBankAccount handles the creation of a new bank account
func (x *Controller) CreateBankAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			AccountID     int32  `json:"account_id"`
			AccountNumber string `json:"account_number"`
			BankName      string `json:"bank_name"`
			BranchName    string `json:"branch_name,omitempty"`
			SwiftCode     string `json:"swift_code,omitempty"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
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
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*jsonBankAccount, error) {
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

			var jsonBankAccount jsonBankAccount
			jsonBankAccount.fromDomain(bankAccount)

			return &jsonBankAccount, nil
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}

// GetBankAccountByAccountID handles the retrieval of a bank account by its associated account ID
func (x *Controller) GetBankAccountByAccountID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var accountID int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*jsonBankAccount, error) {
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

			var jsonBankAccount jsonBankAccount
			jsonBankAccount.fromDomain(bankAccount)

			return &jsonBankAccount, nil
		}).Param("account_id", &accountID).Call(nil).ResponseJSON()
	}
}

// UpdateBankAccount handles the updating of a bank account
func (x *Controller) UpdateBankAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id            int32
			AccountNumber *string `json:"account_number,omitempty"`
			BankName      *string `json:"bank_name,omitempty"`
			BranchName    *string `json:"branch_name,omitempty"`
			SwiftCode     *string `json:"swift_code,omitempty"`
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
