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
	ID           int32      `json:"id"`
	AccountID    int32      `json:"account_id"`
	Date         time.Time  `json:"date"`
	Type         string     `json:"type"`
	Currency     string     `json:"currency"`
	Amount       string     `json:"amount"` // Using string to represent decimal
	Note         string     `json:"note"`
	Adjustable   bool       `json:"adjustable"`
	IsAdjustment bool       `json:"is_adjustment"`
	AdjustedFrom *int32     `json:"adjusted_from"`
	IsVoided     bool       `json:"is_voided"`
	VoidedAt     *time.Time `json:"voided_at"`
	CategoryID   *int32     `json:"category_id,omitempty"` // Added
	CategoryName string     `json:"category_name,omitempty"` // Added
}

func (s *Server) pageLedger(w http.ResponseWriter, r *http.Request) {
	ledgerID := parseInt32(r.PathValue("ledger_id"))

	var ledger ledger
	if err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/ledgers/%d", ledgerID), nil, &ledger); err != nil {
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
	accountIDStr := r.PathValue("account_id")
	accountID, err := s.parseInt32(accountIDStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid account ID: %w", err))
		return
	}

	// Fetch categories for the dropdown
	var categories []category // Assuming 'category' struct is defined in category.go
	err = s.sendRequest(r.Context(), http.MethodGet, "/v1/categories", nil, &categories)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch categories: %w", err))
		// We might still want to render the page, or show a more specific error.
		// For now, let's proceed and the template can handle empty categories.
	}

	data := struct {
		AccountID  int32
		Categories []category
		CSRFToken  string
	}{
		AccountID:  accountID,
		Categories: categories,
		CSRFToken:  s.getCSRFToken(r),
	}

	s.renderPage(w, r, templates.CreateLedgerPage(data.AccountID, data.Categories, data.CSRFToken))
}

func (s *Server) pageLedgerDetails(w http.ResponseWriter, r *http.Request) {
	ledgerIDStr := r.PathValue("ledger_id")
	ledgerID, err := s.parseInt32(ledgerIDStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid ledger ID: %w", err))
		return
	}

	var l ledger
	err = s.sendRequest(r.Context(), http.MethodGet, fmt.Sprintf("/v1/ledgers/%d", ledgerID), nil, &l)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch ledger %d: %w", ledgerID, err))
		return
	}

	// Fetch categories for the dropdown
	var categories []category
	err = s.sendRequest(r.Context(), http.MethodGet, "/v1/categories", nil, &categories)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch categories: %w", err))
		// Proceed without categories if it's a non-critical error for display,
		// but for an edit form, it might be better to show an error.
	}
	
	// Fetch category name if CategoryID is set
	if l.CategoryID != nil && *l.CategoryID != 0 {
		var cat category
		catErr := s.sendRequest(r.Context(), http.MethodGet, fmt.Sprintf("/v1/categories/%d", *l.CategoryID), nil, &cat)
		if catErr == nil {
			l.CategoryName = cat.Name
		} else {
			slog.Warn("failed to fetch category name for ledger", "ledgerID", l.ID, "categoryID", *l.CategoryID, "error", catErr)
		}
	}


	s.renderPage(w, r, templates.LedgerDetailsPage(&l, categories, s.getCSRFToken(r)))
}

func (s *Server) pageLedgersByAccount(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.PathValue("account_id"))

	var ledgers []ledger
	if err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/accounts/%d/ledgers", accountID), nil, &ledgers); err != nil {
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
		AccountID int32
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
		Date       string `json:"date"`
		Type       string `json:"type"`
		Amount     string `json:"amount"`
		Note       string `json:"note"`
		CategoryID *int32 `json:"category_id,omitempty"` // Added
	}

	date := r.FormValue("date")
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		slog.Error("failed to parse date", slog.String("error", err.Error()))
		http.Error(w, "Invalid date format", http.StatusBadRequest)
		return
	}
	payload.Date = t.Format("2006-01-02T15:04:05Z07:00") // API expects RFC3339
	payload.Type = r.FormValue("type")
	payload.Amount = r.FormValue("amount")
	payload.Note = r.FormValue("note")

	categoryIDStr := r.FormValue("category_id")
	if categoryIDStr != "" {
		catID, err := s.parseInt32(categoryIDStr)
		if err == nil && catID != 0 { // Assuming 0 is not a valid category ID for selection
			payload.CategoryID = &catID
		} else if err != nil {
			slog.Error("failed to parse category_id for create ledger", slog.String("category_id_str", categoryIDStr), slog.String("error", err.Error()))
			// Decide if this is a hard error or optional
		}
	}
	
	accountIDStr := r.PathValue("account_id")
	accountID, err := s.parseInt32(accountIDStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid account ID: %w", err))
		return
	}

	// API endpoint is POST /v1/accounts/{account_id}/ledgers
	// The API controller for CreateLedger expects domain.CreateLedgerRequest
	// which contains CategoryID *int32.
	// The payload for sendRequest should match the JSON request body for the API.
	
	// Convert payload struct to map[string]interface{} or directly use map for dynamic fields
    requestBody := map[string]interface{}{
        "date":   payload.Date,
        "type":   payload.Type,
        "amount": payload.Amount,
        "note":   payload.Note,
    }
    if payload.CategoryID != nil {
        requestBody["category_id"] = *payload.CategoryID
    }

	if err := s.sendRequest(r.Context(), http.MethodPost, fmt.Sprintf("/v1/accounts/%d/ledgers", accountID), requestBody, nil); err != nil {
		slog.Error("failed to create ledger", slog.String("accountID", accountIDStr), slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to create ledger", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadLedgers, reloadAccounts")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) updateLedger(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Date       *string `json:"date,omitempty"` // Pointers for optional fields
		Type       *string `json:"type,omitempty"`
		Amount     *string `json:"amount,omitempty"`
		Note       *string `json:"note,omitempty"`
		CategoryID *int32  `json:"category_id,omitempty"` // Added
	}

	ledgerIDStr := r.PathValue("ledger_id")
	ledgerID, err := s.parseInt32(ledgerIDStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid ledger ID: %w", err))
		return
	}

	// Populate payload with form values if they are provided
	if dateStr := r.FormValue("date"); dateStr != "" {
		t, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			s.handleError(w, r, fmt.Errorf("invalid date format for update: %w", err))
			return
		}
		formattedDate := t.Format("2006-01-02T15:04:05Z07:00")
		payload.Date = &formattedDate
	}
	if typeStr := r.FormValue("type"); typeStr != "" {
		payload.Type = &typeStr
	}
	if amountStr := r.FormValue("amount"); amountStr != "" {
		payload.Amount = &amountStr
	}
	if noteStr := r.FormValue("note"); noteStr != "" {
		payload.Note = &noteStr
	}
	if categoryIDStr := r.FormValue("category_id"); categoryIDStr != "" {
		if categoryIDStr == "0" || categoryIDStr == "" { // Allow unsetting category
			var zeroCatID int32 = 0 // API might expect null or a specific value to unset
			// Depending on API, sending category_id: 0 or category_id: null or not sending it.
			// Assuming API interprets missing field as "no change" and explicit null/0 to unset.
			// The API's UpdateLedger expects *int32, so sending a value means update.
			// To "unset", the API might need an explicit null or a specific sentinel.
			// For now, if "0" is chosen, we send 0. If empty, we don't send.
			// If the select has an "empty" option with value "0" or ""
            if categoryIDStr == "0" {
                 payload.CategoryID = &zeroCatID // Send 0 to potentially unset
            }
		} else {
			catID, err := s.parseInt32(categoryIDStr)
			if err == nil {
				payload.CategoryID = &catID
			} else {
				slog.Warn("failed to parse category_id for update ledger", "category_id_str", categoryIDStr, "error", err)
			}
		}
	}
	
	// API endpoint is PATCH /v1/ledgers/{id}
	// The API controller for UpdateLedger expects domain.UpdateLedgerRequest
	// which contains CategoryID *int32.
	// The payload for sendRequest should match the JSON request body for the API.

	if err := s.sendRequest(r.Context(), http.MethodPatch, fmt.Sprintf("/v1/ledgers/%d", ledgerID), payload, nil); err != nil {
		slog.Error("failed to update ledger", slog.String("ledgerID", ledgerIDStr), slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to update ledger", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadLedgers, reloadAccounts")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) voidLedger(w http.ResponseWriter, r *http.Request) {
	ledgerID := parseInt32(r.PathValue("ledger_id"))

	if err := s.sendRequest(r, "DELETE", fmt.Sprintf("/v1/ledgers/%d", ledgerID), nil, nil); err != nil {
		slog.Error("failed to delete ledger", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to delete ledger", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadLedgers, reloadAccounts")
	w.WriteHeader(http.StatusOK)
}
