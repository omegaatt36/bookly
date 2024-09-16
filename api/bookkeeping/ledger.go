package bookkeeping

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/service/bookkeeping"
)

// RegisterLedgerRouters registers ledger-related routes on the provided router.
func (x *Controller) RegisterLedgerRouters(router *http.ServeMux) {
	router.HandleFunc("POST /accounts/{account_id}/ledgers", x.createLedger)
	router.HandleFunc("GET /accounts/{account_id}/ledgers", x.getLedgers)

	router.HandleFunc("GET /ledgers/{id}", x.getLedgerByID)
	router.HandleFunc("PATCH /ledgers/{id}", x.updateLedger)
	router.HandleFunc("DELETE /ledgers/{id}", x.voidLedger)
	router.HandleFunc("POST /ledgers/{id}/adjust", x.adjustLedger)

}

type jsonLedger struct {
	ID           string          `json:"id"`
	AccountID    string          `json:"account_id"`
	Date         time.Time       `json:"date"`
	Type         string          `json:"type"`
	Amount       decimal.Decimal `json:"amount"`
	Note         string          `json:"note"`
	IsAdjustment bool            `json:"is_adjustment"`
	AdjustedFrom *string         `json:"adjusted_from"`
	IsVoided     bool            `json:"is_voided"`
	VoidedAt     *time.Time      `json:"voided_at"`
}

func (l *jsonLedger) fromDomain(ledger *domain.Ledger) {
	l.ID = ledger.ID
	l.AccountID = ledger.AccountID
	l.Date = ledger.Date
	l.Type = ledger.Type.String()
	l.Amount = ledger.Amount
	l.Note = ledger.Note
	l.IsAdjustment = ledger.IsAdjustment
	l.AdjustedFrom = ledger.AdjustedFrom
	l.IsVoided = ledger.IsVoided
	l.VoidedAt = ledger.VoidedAt
}

func (x *Controller) createLedger(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("account_id")
	if accountID == "" {
		http.Error(w, "parameter 'account_id' is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Date   time.Time       `json:"date"`
		Type   string          `json:"type"`
		Amount decimal.Decimal `json:"amount"`
		Note   string          `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ledgerType, err := domain.ParseLedgerType(req.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).CreateLedger(domain.CreateLedgerRequest{
		AccountID: accountID,
		Date:      req.Date,
		Type:      ledgerType,
		Amount:    req.Amount,
		Note:      req.Note,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
}

func (x *Controller) getLedgers(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("account_id")
	if accountID == "" {
		http.Error(w, "parameter 'account_id' is required", http.StatusBadRequest)
		return
	}

	ledgers, err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).GetLedgersByAccountID(accountID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonLedgers := make([]jsonLedger, len(ledgers))
	for index, ledger := range ledgers {
		jsonLedgers[index].fromDomain(ledger)
	}

	bs, err := json.Marshal(jsonLedgers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)
}

func (x *Controller) getLedgerByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	ledger, err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).GetLedgerByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var jsonLedger jsonLedger
	jsonLedger.fromDomain(ledger)

	bs, err := json.Marshal(jsonLedger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)
}

func (x *Controller) updateLedger(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Date   *time.Time `json:"date"`
		Type   *string    `json:"type"`
		Amount *decimal.Decimal
		Note   *string `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var ledgerType *domain.LedgerType
	if req.Type != nil {
		t, err := domain.ParseLedgerType(*req.Type)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ledgerType = &t
	}

	if err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).UpdateLedger(domain.UpdateLedgerRequest{
		ID:     id,
		Date:   req.Date,
		Type:   ledgerType,
		Amount: req.Amount,
		Note:   req.Note,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (x *Controller) voidLedger(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	if err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).VoidLedger(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (x *Controller) adjustLedger(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	var req struct {
		AccountID string          `json:"account_id"`
		Date      time.Time       `json:"date"`
		Type      string          `json:"type"`
		Amount    decimal.Decimal `json:"amount"`
		Note      string          `json:"note"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ledgerType, err := domain.ParseLedgerType(req.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).AdjustLedger(id, domain.CreateLedgerRequest{
		AccountID: req.AccountID,
		Date:      req.Date,
		Type:      ledgerType,
		Amount:    req.Amount,
		Note:      req.Note,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
