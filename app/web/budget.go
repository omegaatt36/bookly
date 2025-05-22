package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/app/web/templates"
	"github.com/omegaatt36/bookly/domain" // For domain.BudgetPeriod
	"github.com/shopspring/decimal"
)

// budget represents a budget for web display
type budget struct {
	ID         int32           `json:"id"`
	Name       string          `json:"name"`
	Period     string          `json:"period"` // domain.BudgetPeriod.String()
	StartDate  time.Time       `json:"start_date"`
	EndDate    time.Time       `json:"end_date"`
	Amount     decimal.Decimal `json:"amount"`
	CategoryID int32           `json:"category_id"`
	// For display purposes, we might want category name
	CategoryName string    `json:"category_name,omitempty"`
	UserID       int32     `json:"user_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// budgetUsage represents budget usage for web display
type budgetUsage struct {
	BudgetID        int32           `json:"budget_id"`
	BudgetName      string          `json:"budget_name"`
	BudgetAmount    decimal.Decimal `json:"budget_amount"`
	SpentAmount     decimal.Decimal `json:"spent_amount"`
	RemainingAmount decimal.Decimal `json:"remaining_amount"`
	Period          string          `json:"period"`
	StartDate       time.Time       `json:"start_date"`
	EndDate         time.Time       `json:"end_date"`
	CategoryID      int32           `json:"category_id"`
	CategoryName    string          `json:"category_name,omitempty"`
}

// pageBudgets fetches budgets via API client and renders budget_list.html.
func (s *Server) pageBudgets(w http.ResponseWriter, r *http.Request) {
	var budgets []budget
	err := s.sendRequest(r.Context(), http.MethodGet, "/v1/budgets", nil, &budgets)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch budgets: %w", err))
		return
	}

	// Optionally, fetch category names for each budget for better display
	// This would involve an additional API call per budget or modifying the /v1/budgets API
	// For simplicity here, we might omit CategoryName or make it part of a later enhancement.

	s.renderPage(w, r, templates.BudgetsPage(budgets, s.getCSRFToken(r)))
}

// pageCreateBudget fetches categories and renders create_budget.html.
func (s *Server) pageCreateBudget(w http.ResponseWriter, r *http.Request) {
	var categories []category // Assuming 'category' struct is defined in category.go
	err := s.sendRequest(r.Context(), http.MethodGet, "/v1/categories", nil, &categories)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch categories for budget creation: %w", err))
		return
	}

	// Prepare data for template (e.g. for period dropdown)
	periods := []struct {
		Value string
		Label string
	}{
		{Value: "monthly", Label: "Monthly"},
		{Value: "yearly", Label: "Yearly"},
	}

	s.renderPage(w, r, templates.CreateBudgetPage(categories, periods, s.getCSRFToken(r)))
}

// createBudget handles form submission from create_budget.html
func (s *Server) createBudget(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.handleError(w, r, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	name := r.FormValue("name")
	periodStr := r.FormValue("period")
	startDateStr := r.FormValue("start_date") // Expects YYYY-MM-DD
	amountStr := r.FormValue("amount")
	categoryIDStr := r.FormValue("category_id")

	// Validations
	if name == "" || periodStr == "" || amountStr == "" || categoryIDStr == "" {
		s.sessionManager.Put(r.Context(), "error", "All fields (Name, Period, Amount, Category) are required.")
		http.Redirect(w, r, "/budgets/create", http.StatusSeeOther) // Consider re-populating form
		return
	}

	var startDate time.Time
	if startDateStr != "" { // StartDate is optional in API, service defaults it. If provided, parse it.
		parsedDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			s.sessionManager.Put(r.Context(), "error", "Invalid start date format. Use YYYY-MM-DD.")
			http.Redirect(w, r, "/budgets/create", http.StatusSeeOther)
			return
		}
		startDate = parsedDate
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		s.sessionManager.Put(r.Context(), "error", "Invalid amount format.")
		http.Redirect(w, r, "/budgets/create", http.StatusSeeOther)
		return
	}

	categoryID, err := s.parseInt32(categoryIDStr)
	if err != nil {
		s.sessionManager.Put(r.Context(), "error", "Invalid category ID.")
		http.Redirect(w, r, "/budgets/create", http.StatusSeeOther)
		return
	}

	payload := map[string]interface{}{
		"name":        name,
		"period":      periodStr,
		"amount":      amount.String(), // API expects string for decimal
		"category_id": categoryID,
	}
	if !startDate.IsZero() {
		payload["start_date"] = startDate.Format(time.RFC3339) // API expects RFC3339
	}

	payloadBytes, _ := json.Marshal(payload)

	var createdBudget budget
	err = s.sendRequest(r.Context(), http.MethodPost, "/v1/budgets", bytes.NewBuffer(payloadBytes), &createdBudget)
	if err != nil {
		s.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Failed to create budget: %s", err.Error()))
		http.Redirect(w, r, "/budgets/create", http.StatusSeeOther)
		return
	}

	s.sessionManager.Put(r.Context(), "success", "Budget created successfully.")
	http.Redirect(w, r, "/budgets", http.StatusSeeOther)
}

// pageBudgetDetails fetches budget details and usage, renders budget_details.html.
func (s *Server) pageBudgetDetails(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := s.parseInt32(idStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid budget ID: %w", err))
		return
	}

	var bud budget
	err = s.sendRequest(r.Context(), http.MethodGet, fmt.Sprintf("/v1/budgets/%d", budgetID), nil, &bud)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch budget %d: %w", budgetID, err))
		return
	}

	var usage budgetUsage
	err = s.sendRequest(r.Context(), http.MethodGet, fmt.Sprintf("/v1/budgets/%d/usage", budgetID), nil, &usage)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch budget usage for budget %d: %w", budgetID, err))
		// Still might want to render the page with partial data if budget fetch succeeded
	}
	
	// Optionally fetch category name for budget and usage
	var cat category
	if bud.CategoryID != 0 {
	    s.sendRequest(r.Context(), http.MethodGet, fmt.Sprintf("/v1/categories/%d", bud.CategoryID), nil, &cat)
	    bud.CategoryName = cat.Name
	    if usage.BudgetID != 0 { // if usage was fetched successfully
	        usage.CategoryName = cat.Name
	    }
	}


	s.renderPage(w, r, templates.BudgetDetailsPage(&bud, &usage))
}

// pageEditBudget fetches budget and categories, renders edit_budget.html.
func (s *Server) pageEditBudget(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := s.parseInt32(idStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid budget ID: %w", err))
		return
	}

	var bud budget
	err = s.sendRequest(r.Context(), http.MethodGet, fmt.Sprintf("/v1/budgets/%d", budgetID), nil, &bud)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch budget %d for editing: %w", budgetID, err))
		return
	}

	var categories []category
	err = s.sendRequest(r.Context(), http.MethodGet, "/v1/categories", nil, &categories)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("failed to fetch categories for budget editing: %w", err))
		return
	}
	
	periods := []struct {
		Value string
		Label string
	}{
		{Value: "monthly", Label: "Monthly"},
		{Value: "yearly", Label: "Yearly"},
	}

	s.renderPage(w, r, templates.EditBudgetPage(&bud, categories, periods, s.getCSRFToken(r)))
}

// updateBudget handles form submission from edit_budget.html
func (s *Server) updateBudget(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := s.parseInt32(idStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid budget ID: %w", err))
		return
	}

	if err := r.ParseForm(); err != nil {
		s.handleError(w, r, fmt.Errorf("failed to parse form: %w", err))
		return
	}

	name := r.FormValue("name")
	periodStr := r.FormValue("period")
	startDateStr := r.FormValue("start_date")
	amountStr := r.FormValue("amount")
	categoryIDStr := r.FormValue("category_id")

	// Construct payload with only non-empty fields for partial update
	payload := make(map[string]interface{})
	if name != "" {
		payload["name"] = name
	}
	if periodStr != "" {
		// Validate period string
		_, err := domain.ParseBudgetPeriod(periodStr)
		if err != nil {
			s.sessionManager.Put(r.Context(), "error", "Invalid period value.")
			http.Redirect(w, r, fmt.Sprintf("/budgets/%d/edit", budgetID), http.StatusSeeOther)
			return
		}
		payload["period"] = periodStr
	}
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			s.sessionManager.Put(r.Context(), "error", "Invalid start date format.")
			http.Redirect(w, r, fmt.Sprintf("/budgets/%d/edit", budgetID), http.StatusSeeOther)
			return
		}
		payload["start_date"] = startDate.Format(time.RFC3339)
	}
	if amountStr != "" {
		amount, err := decimal.NewFromString(amountStr)
		if err != nil {
			s.sessionManager.Put(r.Context(), "error", "Invalid amount format.")
			http.Redirect(w, r, fmt.Sprintf("/budgets/%d/edit", budgetID), http.StatusSeeOther)
			return
		}
		payload["amount"] = amount.String()
	}
	if categoryIDStr != "" {
		categoryID, err := s.parseInt32(categoryIDStr)
		if err != nil {
			s.sessionManager.Put(r.Context(), "error", "Invalid category ID.")
			http.Redirect(w, r, fmt.Sprintf("/budgets/%d/edit", budgetID), http.StatusSeeOther)
			return
		}
		payload["category_id"] = categoryID
	}
	
	if len(payload) == 0 {
		s.sessionManager.Put(r.Context(), "info", "No changes submitted.")
		http.Redirect(w, r, fmt.Sprintf("/budgets/%d/edit", budgetID), http.StatusSeeOther)
        return
	}

	payloadBytes, _ := json.Marshal(payload)

	err = s.sendRequest(r.Context(), http.MethodPut, fmt.Sprintf("/v1/budgets/%d", budgetID), bytes.NewBuffer(payloadBytes), nil)
	if err != nil {
		s.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Failed to update budget: %s", err.Error()))
		http.Redirect(w, r, fmt.Sprintf("/budgets/%d/edit", budgetID), http.StatusSeeOther)
		return
	}

	s.sessionManager.Put(r.Context(), "success", "Budget updated successfully.")
	http.Redirect(w, r, "/budgets", http.StatusSeeOther)
}

// deleteBudget handles deletion of a budget
func (s *Server) deleteBudget(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := s.parseInt32(idStr)
	if err != nil {
		s.handleError(w, r, fmt.Errorf("invalid budget ID: %w", err))
		return
	}

	err = s.sendRequest(r.Context(), http.MethodDelete, fmt.Sprintf("/v1/budgets/%d", budgetID), nil, nil)
	if err != nil {
		s.sessionManager.Put(r.Context(), "error", fmt.Sprintf("Failed to delete budget: %s", err.Error()))
		http.Redirect(w, r, "/budgets", http.StatusSeeOther)
		return
	}

	s.sessionManager.Put(r.Context(), "success", "Budget deleted successfully.")
	http.Redirect(w, r, "/budgets", http.StatusSeeOther)
}
