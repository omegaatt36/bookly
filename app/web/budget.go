package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/app"
)

type budget struct {
	ID         int32  `json:"id"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
	UserID     int32  `json:"user_id"`
	Name       string `json:"name"`
	Category   string `json:"category"`
	Amount     string `json:"amount"`
	PeriodType string `json:"period_type"`
	StartDate  string `json:"start_date"`
	EndDate    string `json:"end_date,omitempty"`
	IsActive   bool   `json:"is_active"`
}

type budgetSummary struct {
	Budget      budget `json:"budget"`
	UsedAmount  string `json:"used_amount"`
	Percentage  string `json:"percentage"`
	PeriodStart string `json:"period_start"`
	PeriodEnd   string `json:"period_end"`
}

func (s *Server) pageCreateBudget(w http.ResponseWriter, _ *http.Request) {
	if err := s.templates.ExecuteTemplate(w, "create_budget.html", nil); err != nil {
		slog.Error("failed to render create_budget.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) pageBudgets(w http.ResponseWriter, r *http.Request) {
	var budgets []*budget
	err := s.sendRequest(r, "GET", "/v1/budgets", nil, &budgets)
	if err != nil {
		slog.Error("failed to get budgets", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get budgets", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Budgets": budgets,
	}

	if err := s.templates.ExecuteTemplate(w, "budgets_page.html", data); err != nil {
		slog.Error("failed to render budgets_page.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) pageBudgetList(w http.ResponseWriter, r *http.Request) {
	var budgets []*budget
	err := s.sendRequest(r, "GET", "/v1/budgets", nil, &budgets)
	if err != nil {
		slog.Error("failed to get budgets", slog.String("error", err.Error()))
		http.Error(w, "Failed to get budgets", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Budgets": budgets,
	}

	if err := s.templates.ExecuteTemplate(w, "budget_list.html", data); err != nil {
		slog.Error("failed to render budget_list.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) pageBudget(w http.ResponseWriter, r *http.Request) {
	budgetID := parseInt32(r.PathValue("budget_id"))

	var summary budgetSummary
	err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/budgets/%d/summary", budgetID), nil, &summary)
	if err != nil {
		slog.Error("failed to get budget summary", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get budget", http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Budget":      summary.Budget,
		"UsedAmount":  summary.UsedAmount,
		"Percentage":  summary.Percentage,
		"PeriodStart": summary.PeriodStart,
		"PeriodEnd":   summary.PeriodEnd,
	}

	if err := s.templates.ExecuteTemplate(w, "budget_details.html", data); err != nil {
		slog.Error("failed to render budget_details.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) createBudget(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	category := r.FormValue("category")
	amount := r.FormValue("amount")
	periodType := r.FormValue("period_type")
	startDateStr := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")

	// Parse and format start date
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		slog.Error("failed to parse start date", slog.String("error", err.Error()))
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}

	payload := map[string]any{
		"name":        name,
		"category":    category,
		"amount":      amount,
		"period_type": periodType,
		"start_date":  startDate.Format(time.RFC3339),
	}

	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			slog.Error("failed to parse end date", slog.String("error", err.Error()))
			http.Error(w, "Invalid end date format", http.StatusBadRequest)
			return
		}
		payload["end_date"] = endDate.Format(time.RFC3339)
	}

	var result budget
	err = s.sendRequest(r, "POST", "/v1/budgets", payload, &result)
	if err != nil {
		slog.Error("failed to create budget", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to create budget", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadBudgets")
	w.Header().Set("HX-Redirect", fmt.Sprintf("/page/budgets/%d", result.ID))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) updateBudget(w http.ResponseWriter, r *http.Request) {
	budgetID := parseInt32(r.PathValue("budget_id"))

	payload := make(map[string]any)
	var err error

	if name := r.FormValue("name"); name != "" {
		payload["name"] = name
	}
	if category := r.FormValue("category"); category != "" {
		payload["category"] = category
	}
	if amount := r.FormValue("amount"); amount != "" {
		payload["amount"] = amount
	}
	if periodType := r.FormValue("period_type"); periodType != "" {
		payload["period_type"] = periodType
	}
	if startDateStr := r.FormValue("start_date"); startDateStr != "" {
		var startDate time.Time
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			slog.Error("failed to parse start date", slog.String("error", err.Error()))
			http.Error(w, "Invalid start date format", http.StatusBadRequest)
			return
		}
		payload["start_date"] = startDate.Format(time.RFC3339)
	}
	if endDateStr := r.FormValue("end_date"); endDateStr != "" {
		var endDate time.Time
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			slog.Error("failed to parse end date", slog.String("error", err.Error()))
			http.Error(w, "Invalid end date format", http.StatusBadRequest)
			return
		}
		payload["end_date"] = endDate.Format(time.RFC3339)
	}
	if isActive := r.FormValue("is_active"); isActive != "" {
		payload["is_active"] = isActive == "true"
	}

	var result budget
	err = s.sendRequest(r, "PATCH", fmt.Sprintf("/v1/budgets/%d", budgetID), payload, &result)
	if err != nil {
		slog.Error("failed to update budget", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to update budget", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadBudgets")
	w.Header().Set("HX-Redirect", fmt.Sprintf("/page/budgets/%d", budgetID))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteBudget(w http.ResponseWriter, r *http.Request) {
	budgetID := parseInt32(r.PathValue("budget_id"))

	err := s.sendRequest(r, "DELETE", fmt.Sprintf("/v1/budgets/%d", budgetID), nil, nil)
	if err != nil {
		slog.Error("failed to delete budget", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to delete budget", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadBudgets")
	w.Header().Set("HX-Redirect", "/page/budgets")
	w.WriteHeader(http.StatusOK)
}
