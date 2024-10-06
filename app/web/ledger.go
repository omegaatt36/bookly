package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/app"
)

type ledger struct {
	ID           string     `json:"id"`
	AccountID    string     `json:"account_id"`
	Date         time.Time  `json:"date"`
	Type         string     `json:"type"`
	Amount       string     `json:"amount"` // Using string to represent decimal
	Note         string     `json:"note"`
	Adjustable   bool       `json:"adjustable"`
	IsAdjustment bool       `json:"is_adjustment"`
	AdjustedFrom *string    `json:"adjusted_from"`
	IsVoided     bool       `json:"is_voided"`
	VoidedAt     *time.Time `json:"voided_at"`
}

func (s *Server) pageLedger(w http.ResponseWriter, r *http.Request) {
	ledgerID := r.PathValue("ledger_id")

	var ledger ledger
	if err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/ledgers/%s", ledgerID), nil, &ledger); err != nil {
		slog.Error("failed to get ledgers", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get ledgers", http.StatusInternalServerError)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "ledger.html", ledger); err != nil {
		slog.Error("failed to render ledger.html", slog.String("error", err.Error()))
	}
}

func (s *Server) pageCreateLedger(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("account_id")

	result := struct {
		ID string `json:"id"`
	}{
		ID: accountID,
	}

	if err := s.templates.ExecuteTemplate(w, "create_ledger.html", result); err != nil {
		slog.Error("failed to render new_account.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) pageLedgerDetails(w http.ResponseWriter, r *http.Request) {
	ledgerID := r.PathValue("ledger_id")

	var ledgerDetails ledger
	if err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/ledgers/%s", ledgerID), nil, &ledgerDetails); err != nil {
		slog.Error("failed to get ledger details", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get ledger details", http.StatusInternalServerError)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "ledger_details.html", ledgerDetails); err != nil {
		slog.Error("failed to render ledger_details.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) pageLedgersByAccount(w http.ResponseWriter, r *http.Request) {

	accountID := r.PathValue("account_id")

	var ledgers []ledger
	if err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/accounts/%s/ledgers", accountID), nil, &ledgers); err != nil {
		slog.Error("failed to get ledgers", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get ledgers", http.StatusInternalServerError)
		return
	}

	result := struct {
		AccountID string
		Ledgers   []ledger
	}{
		AccountID: accountID,
		Ledgers:   ledgers,
	}

	if err := s.templates.ExecuteTemplate(w, "ledger_list.html", result); err != nil {
		slog.Error("failed to render ledger_list.html", slog.String("error", err.Error()))
	}
}

func (s *Server) createLedger(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Date   string `json:"date"`
		Type   string `json:"type"`
		Amount string `json:"amount"`
		Note   string `json:"note"`
	}

	date := r.FormValue("date")
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		slog.Error("failed to parse date", slog.String("error", err.Error()))
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}
	payload.Date = t.Format("2006-01-02T15:04:05Z07:00")
	payload.Type = r.FormValue("type")
	payload.Amount = r.FormValue("amount")
	payload.Note = r.FormValue("note")

	if err := s.sendRequest(r, "POST", "/v1/accounts/"+r.PathValue("account_id")+"/ledgers", payload, nil); err != nil {
		slog.Error("failed to create ledger", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to create ledger", http.StatusInternalServerError)
		return
	}

	s.pageLedgersByAccount(w, r)
}

func (s *Server) updateLedger(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Date   string `json:"date,omitempty"`
		Type   string `json:"type,omitempty"`
		Amount string `json:"amount,omitempty"`
		Note   string `json:"note,omitempty"`
	}

	ledgerID := r.PathValue("ledger_id")

	date := r.FormValue("date")
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		slog.Error("failed to parse date", slog.String("error", err.Error()))
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}
	payload.Date = t.Format("2006-01-02T15:04:05Z07:00")
	payload.Type = r.FormValue("type")
	payload.Amount = r.FormValue("amount")
	payload.Note = r.FormValue("note")

	if err := s.sendRequest(r, "PATCH", fmt.Sprintf("/v1/ledgers/%s", ledgerID), payload, nil); err != nil {
		slog.Error("failed to update ledger", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to update ledger", http.StatusInternalServerError)
		return
	}

	s.pageLedger(w, r)
}
