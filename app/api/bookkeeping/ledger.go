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

type jsonLedger struct {
	ID           int32           `json:"id"`
	AccountID    int32           `json:"account_id"`
	Date         time.Time       `json:"date"`
	Type         string          `json:"type"`
	Currency     string          `json:"currency"`
	Amount       decimal.Decimal `json:"amount"`
	Note         string          `json:"note"`
	Category     string          `json:"category"`
	Adjustable   bool            `json:"adjustable"`
	IsAdjustment bool            `json:"is_adjustment"`
	AdjustedFrom *int32          `json:"adjusted_from"`
	IsVoided     bool            `json:"is_voided"`
	VoidedAt     *time.Time      `json:"voided_at"`
}

func (l *jsonLedger) fromDomain(ledger *domain.Ledger) {
	l.ID = ledger.ID
	l.AccountID = ledger.AccountID
	l.Date = ledger.Date
	l.Type = ledger.Type.String()
	l.Currency = ledger.Currency
	l.Amount = ledger.Amount
	l.Note = ledger.Note
	l.Category = ledger.Category
	l.Adjustable = time.Since(ledger.CreatedAt) <= domain.EditableDuration
	l.IsAdjustment = ledger.IsAdjustment
	l.AdjustedFrom = ledger.AdjustedFrom
	l.IsVoided = ledger.IsVoided
	l.VoidedAt = ledger.VoidedAt
}

// CreateLedger handles the creation of a new ledger entry
func (x *Controller) CreateLedger() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			accountID int32
			Date      time.Time       `json:"date"`
			Type      string          `json:"type"`
			Amount    decimal.Decimal `json:"amount"`
			Note      string          `json:"note"`
			Category  string          `json:"category"`
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

			if req.Date.IsZero() {
				req.Date = time.Now()
			}

			if req.Amount.IsZero() {
				return nil, app.ParamError(errors.New("amount is required"))
			}

			_, err = x.service.CreateLedger(domain.CreateLedgerRequest{
				AccountID: req.accountID,
				Date:      req.Date,
				Type:      ledgerType,
				Amount:    req.Amount,
				Note:      req.Note,
				Category:  req.Category,
			})

			return nil, err
		}).Param("account_id", &req.accountID).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// GetLedgersByAccount retrieves all ledger entries for a given account
func (x *Controller) GetLedgersByAccount() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var accountID int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]jsonLedger, error) {
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

			jsonLedgers := make([]jsonLedger, len(ledgers))
			for index, ledger := range ledgers {
				jsonLedgers[index].fromDomain(ledger)
			}

			return jsonLedgers, nil
		}).Param("account_id", &accountID).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetLedgerByID retrieves a specific ledger entry by its ID
func (x *Controller) GetLedgerByID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*jsonLedger, error) {
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

			var jsonLedger jsonLedger
			jsonLedger.fromDomain(ledger)

			return &jsonLedger, nil
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}

// UpdateLedger handles the update of an existing ledger entry
func (x *Controller) UpdateLedger() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id       int32
			Date     *time.Time       `json:"date"`
			Type     *string          `json:"type"`
			Amount   *decimal.Decimal `json:"amount"`
			Note     *string          `json:"note"`
			Category *string          `json:"category"`
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

			var ledgerType *domain.LedgerType
			if req.Type != nil {
				t, err := domain.ParseLedgerType(*req.Type)
				if err != nil {
					return nil, err
				}
				ledgerType = &t
			}

			err = x.service.UpdateLedger(domain.UpdateLedgerRequest{
				ID:       req.id,
				Date:     req.Date,
				Type:     ledgerType,
				Amount:   req.Amount,
				Note:     req.Note,
				Category: req.Category,
			})
			return nil, err
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
			id        int32
			AccountID int32           `json:"account_id"`
			Date      time.Time       `json:"date"`
			Type      string          `json:"type"`
			Amount    decimal.Decimal `json:"amount"`
			Note      string          `json:"note"`
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

			if req.Amount.IsZero() {
				return nil, app.ParamError(errors.New("amount is required"))
			}

			return nil, x.service.AdjustLedger(req.id, domain.CreateLedgerRequest{
				AccountID: req.AccountID,
				Date:      req.Date,
				Type:      ledgerType,
				Amount:    req.Amount,
				Note:      req.Note,
			})
		}).Param("id", &req.id).BindJSON(&req).Call(req).ResponseJSON()
	}
}
