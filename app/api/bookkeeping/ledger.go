package bookkeeping

import (
	"errors"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/sdk/datatype"
)

func convertLedgerFromDomain(ledger *domain.Ledger) datatype.Ledger {
	return datatype.Ledger{
		ID:           ledger.ID,
		AccountID:    ledger.AccountID,
		Date:         ledger.Date.Format(datatype.TimeFormat),
		Type:         ledger.Type.String(),
		Currency:     ledger.Currency,
		Amount:       ledger.Amount.String(),
		Note:         ledger.Note,
		Adjustable:   time.Since(ledger.CreatedAt) <= domain.EditableDuration,
		IsAdjustment: ledger.IsAdjustment,
		AdjustedFrom: ledger.AdjustedFrom,
		IsVoided:     ledger.IsVoided,
		VoidedAt: func() *string {
			if ledger.VoidedAt != nil {
				t := ledger.VoidedAt.Format(datatype.TimeFormat)
				return &t
			}
			return nil
		}(),
	}
}

// CreateLedger handles the creation of a new ledger entry
func (x *Controller) CreateLedger() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			accountID int32
			datatype.CreateLedgerRequest
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(req.accountID)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			ledgerType, err := domain.ParseLedgerType(req.Type)
			if err != nil {
				return nil, err
			}

			date := time.Now()
			if req.Date != "" {
				date, err = time.Parse(datatype.TimeFormat, req.Date)
				if err != nil {
					return nil, app.ParamError(errors.New("invalid date format"))
				}
			}

			amount, err := decimal.NewFromString(req.Amount)
			if err != nil {
				return nil, app.ParamError(errors.New("amount is required"))
			}

			_, err = x.service.CreateLedger(domain.CreateLedgerRequest{
				AccountID: req.accountID,
				Date:      date,
				Type:      ledgerType,
				Amount:    amount,
				Note:      req.Note,
			})

			return nil, err
		}).Param("account_id", &req.accountID).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// GetLedgersByAccount retrieves all ledger entries for a given account
func (x *Controller) GetLedgersByAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var accountID int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]datatype.Ledger, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(accountID)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: account does not belong to user"))
			}

			ledgers, err := x.service.GetLedgersByAccountID(accountID)
			if err != nil {
				return nil, err
			}

			jsonLedgers := make([]datatype.Ledger, len(ledgers))
			for index, ledger := range ledgers {
				jsonLedgers[index] = convertLedgerFromDomain(ledger)
			}

			return jsonLedgers, nil
		}).Param("account_id", &accountID).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetLedgerByID retrieves a specific ledger entry by its ID
func (x *Controller) GetLedgerByID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*datatype.Ledger, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			ledger, err := x.service.GetLedgerByID(id)
			if err != nil {
				return nil, err
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(ledger.AccountID)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: ledger does not belong to user"))
			}

			jsonLedger := convertLedgerFromDomain(ledger)

			return &jsonLedger, nil
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}

// UpdateLedger handles the update of an existing ledger entry
func (x *Controller) UpdateLedger() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id int32
			datatype.UpdateLedgerRequest
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify ledger ownership
			ledger, err := x.service.GetLedgerByID(req.id)
			if err != nil {
				return nil, err
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(ledger.AccountID)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: ledger does not belong to user"))
			}

			updateLedgerReq := domain.UpdateLedgerRequest{
				ID:   req.id,
				Note: req.Note,
			}

			if req.Type != nil {
				t, err := domain.ParseLedgerType(*req.Type)
				if err != nil {
					return nil, err
				}
				updateLedgerReq.Type = &t
			}

			if req.Date != nil {
				date, err := time.Parse(datatype.TimeFormat, *req.Date)
				if err != nil {
					return nil, app.ParamError(errors.New("invalid date format"))
				}
				updateLedgerReq.Date = &date
			}

			if req.Amount != nil {
				a, err := decimal.NewFromString(*req.Amount)
				if err != nil {
					return nil, app.ParamError(errors.New("amount is required"))
				}
				updateLedgerReq.Amount = &a
			}

			return nil, x.service.UpdateLedger(updateLedgerReq)
		}).Param("id", &req.id).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// VoidLedger handles the voiding of a ledger entry
func (x *Controller) VoidLedger() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify ledger ownership
			ledger, err := x.service.GetLedgerByID(id)
			if err != nil {
				return nil, err
			}

			// Verify account ownership
			account, err := x.service.GetAccountByID(ledger.AccountID)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: ledger does not belong to user"))
			}

			return nil, x.service.VoidLedger(id)
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}

// AdjustLedger handles the adjustment of an existing ledger entry
func (x *Controller) AdjustLedger() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id int32
			datatype.CreateLedgerRequest
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// Verify ledger ownership
			ledger, err := x.service.GetLedgerByID(req.id)
			if err != nil {
				return nil, err
			}

			// Verify original account ownership
			account, err := x.service.GetAccountByID(ledger.AccountID)
			if err != nil {
				return nil, err
			}
			if account.UserID != userID {
				return nil, app.Forbidden(errors.New("access denied: ledger does not belong to user"))
			}

			// Verify new account ownership if changing account
			if req.AccountID != ledger.AccountID {
				newAccount, err := x.service.GetAccountByID(req.AccountID)
				if err != nil {
					return nil, err
				}
				if newAccount.UserID != userID {
					return nil, app.Forbidden(errors.New("access denied: new account does not belong to user"))
				}
			}

			ledgerType, err := domain.ParseLedgerType(req.Type)
			if err != nil {
				return nil, err
			}

			date := time.Now()
			if req.Date != "" {
				date, err = time.Parse(datatype.TimeFormat, req.Date)
				if err != nil {
					return nil, app.ParamError(errors.New("invalid date format"))
				}
			}

			amount, err := decimal.NewFromString(req.Amount)
			if err != nil {
				return nil, app.ParamError(errors.New("amount is required"))
			}

			return nil, x.service.AdjustLedger(req.id, domain.CreateLedgerRequest{
				AccountID: req.AccountID,
				Date:      date,
				Type:      ledgerType,
				Amount:    amount,
				Note:      req.Note,
			})
		}).Param("id", &req.id).BindJSON(&req).Call(req).ResponseJSON()
	}
}
