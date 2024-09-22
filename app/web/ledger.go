package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

type ledger struct {
	ID     string    `json:"id"`
	Date   time.Time `json:"date"`
	Type   string    `json:"type"`
	Amount string    `json:"amount"` // Using string to represent decimal
	Note   string    `json:"note"`
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
		http.Error(w, "Failed to create ledger", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/")
}

func (s *Server) getLedgers(r *http.Request, accountID string) ([]ledger, error) {
	var ledgers []ledger
	err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/accounts/%s/ledgers", accountID), nil, &ledgers)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledgers: %w", err)
	}
	return ledgers, nil
}

func (s *Server) getLedgersByAccount(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("account_id")
	ledgers, err := s.getLedgers(r, accountID)
	if err != nil {
		slog.Error("failed to get ledgers", slog.String("error", err.Error()))
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