package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/omegaatt36/bookly/app"
)

type recurringTransaction struct {
	ID           int32      `json:"id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Amount       string     `json:"amount"`
	Note         string     `json:"note"`
	StartDate    time.Time  `json:"start_date"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	RecurType    string     `json:"recur_type"`
	Status       string     `json:"status"`
	Frequency    int        `json:"frequency"`
	DayOfWeek    *int       `json:"day_of_week,omitempty"`
	DayOfMonth   *int       `json:"day_of_month,omitempty"`
	MonthOfYear  *int       `json:"month_of_year,omitempty"`
	LastExecuted *time.Time `json:"last_executed,omitempty"`
	NextDue      time.Time  `json:"next_due"`
}

type reminder struct {
	ID                     int32      `json:"id"`
	CreatedAt              time.Time  `json:"created_at"`
	RecurringTransactionID int32      `json:"recurring_transaction_id"`
	ReminderDate           time.Time  `json:"reminder_date"`
	IsRead                 bool       `json:"is_read"`
	ReadAt                 *time.Time `json:"read_at,omitempty"`
}

func (s *Server) pageRecurringList(w http.ResponseWriter, r *http.Request) {
	var recurring []recurringTransaction
	err := s.sendRequest(r, "GET", "/v1/recurring", nil, &recurring)
	if err != nil {
		slog.Error("failed to get recurring transactions", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get recurring transactions", http.StatusInternalServerError)
		return
	}

	// Also fetch accounts for the account selector in the create form
	var accounts []account
	err = s.sendRequest(r, "GET", "/v1/accounts", nil, &accounts)
	if err != nil {
		slog.Error("failed to get accounts", slog.String("error", err.Error()))
		// Continue with empty accounts list
	}

	result := struct {
		RecurringTransactions []recurringTransaction
		Accounts              []account
	}{
		RecurringTransactions: recurring,
		Accounts:              accounts,
	}

	if err := s.templates.ExecuteTemplate(w, "recurring_list.html", result); err != nil {
		slog.Error("failed to render recurring_list.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) pageRecurringDetails(w http.ResponseWriter, r *http.Request) {
	id := parseInt32(r.PathValue("recurring_id"))

	var recurring recurringTransaction
	err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/recurring/%d", id), nil, &recurring)
	if err != nil {
		slog.Error("failed to get recurring transaction", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get recurring transaction", http.StatusInternalServerError)
		return
	}

	// Get accounts for the edit form
	var accounts []account
	err = s.sendRequest(r, "GET", "/v1/accounts", nil, &accounts)
	if err != nil {
		slog.Error("failed to get accounts", slog.String("error", err.Error()))
		// Continue with empty accounts list
	}

	result := struct {
		RecurringTransaction recurringTransaction
		Accounts             []account
	}{
		RecurringTransaction: recurring,
		Accounts:             accounts,
	}

	if err := s.templates.ExecuteTemplate(w, "recurring_details.html", result); err != nil {
		slog.Error("failed to render recurring_details.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) pageCreateRecurring(w http.ResponseWriter, r *http.Request) {
	// Get accounts for the create form
	var accounts []account
	err := s.sendRequest(r, "GET", "/v1/accounts", nil, &accounts)
	if err != nil {
		slog.Error("failed to get accounts", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get accounts", http.StatusInternalServerError)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "create_recurring.html", accounts); err != nil {
		slog.Error("failed to render create_recurring.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) pageReminders(w http.ResponseWriter, r *http.Request) {
	var reminders []reminder
	err := s.sendRequest(r, "GET", "/v1/recurring/reminders", nil, &reminders)
	if err != nil {
		slog.Error("failed to get reminders", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get reminders", http.StatusInternalServerError)
		return
	}

	if err := s.templates.ExecuteTemplate(w, "reminders.html", reminders); err != nil {
		slog.Error("failed to render reminders.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) createRecurring(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		AccountID   int32    `json:"account_id"`
		Name        string   `json:"name"`
		Type        string   `json:"type"`
		Amount      string   `json:"amount"`
		Note        string   `json:"note"`
		StartDate   string   `json:"start_date"`
		EndDate     *string  `json:"end_date,omitempty"`
		RecurType   string   `json:"recur_type"`
		Frequency   int      `json:"frequency"`
		DayOfWeek   *int     `json:"day_of_week,omitempty"`
		DayOfMonth  *int     `json:"day_of_month,omitempty"`
		MonthOfYear *int     `json:"month_of_year,omitempty"`
	}

	accountIDStr := r.FormValue("account_id")
	accountID, _ := strconv.ParseInt(accountIDStr, 10, 32)
	payload.AccountID = int32(accountID)
	payload.Name = r.FormValue("name")
	payload.Type = r.FormValue("type")
	payload.Amount = r.FormValue("amount")
	payload.Note = r.FormValue("note")

	startDate, err := time.Parse("2006-01-02", r.FormValue("start_date"))
	if err != nil {
		slog.Error("failed to parse start date", slog.String("error", err.Error()))
		http.Error(w, "Invalid start date format", http.StatusBadRequest)
		return
	}
	payload.StartDate = startDate.Format(time.RFC3339)

	if r.FormValue("end_date") != "" {
		endDate, err := time.Parse("2006-01-02", r.FormValue("end_date"))
		if err != nil {
			slog.Error("failed to parse end date", slog.String("error", err.Error()))
			http.Error(w, "Invalid end date format", http.StatusBadRequest)
			return
		}
		endDateStr := endDate.Format(time.RFC3339)
		payload.EndDate = &endDateStr
	}

	payload.RecurType = r.FormValue("recur_type")
	frequency, err := strconv.Atoi(r.FormValue("frequency"))
	if err != nil {
		slog.Error("failed to parse frequency", slog.String("error", err.Error()))
		http.Error(w, "Invalid frequency", http.StatusBadRequest)
		return
	}
	payload.Frequency = frequency

	switch payload.RecurType {
	case "weekly":
		if dow := r.FormValue("day_of_week"); dow != "" {
			dayOfWeek, err := strconv.Atoi(dow)
			if err == nil && dayOfWeek >= 0 && dayOfWeek <= 6 {
				payload.DayOfWeek = &dayOfWeek
			}
		}
	case "monthly":
		if dom := r.FormValue("day_of_month"); dom != "" {
			dayOfMonth, err := strconv.Atoi(dom)
			if err == nil && dayOfMonth >= 1 && dayOfMonth <= 31 {
				payload.DayOfMonth = &dayOfMonth
			}
		}
	case "yearly":
		if moy := r.FormValue("month_of_year"); moy != "" {
			monthOfYear, err := strconv.Atoi(moy)
			if err == nil && monthOfYear >= 1 && monthOfYear <= 12 {
				payload.MonthOfYear = &monthOfYear
			}
		}
	}

	if err := s.sendRequest(r, "POST", "/v1/recurring", payload, nil); err != nil {
		slog.Error("failed to create recurring transaction", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to create recurring transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadRecurring")
	w.Header().Set("HX-Redirect", "/page/recurring")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) updateRecurring(w http.ResponseWriter, r *http.Request) {
	id := parseInt32(r.PathValue("recurring_id"))

	var payload struct {
		Name        *string  `json:"name,omitempty"`
		Type        *string  `json:"type,omitempty"`
		Amount      *string  `json:"amount,omitempty"`
		Note        *string  `json:"note,omitempty"`
		EndDate     *string  `json:"end_date,omitempty"`
		RecurType   *string  `json:"recur_type,omitempty"`
		Status      *string  `json:"status,omitempty"`
		Frequency   *int     `json:"frequency,omitempty"`
		DayOfWeek   *int     `json:"day_of_week,omitempty"`
		DayOfMonth  *int     `json:"day_of_month,omitempty"`
		MonthOfYear *int     `json:"month_of_year,omitempty"`
	}

	if name := r.FormValue("name"); name != "" {
		payload.Name = &name
	}

	if typ := r.FormValue("type"); typ != "" {
		payload.Type = &typ
	}

	if amount := r.FormValue("amount"); amount != "" {
		payload.Amount = &amount
	}

	if note := r.FormValue("note"); note != "" {
		payload.Note = &note
	}

	if endDate := r.FormValue("end_date"); endDate != "" {
		date, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			slog.Error("failed to parse end date", slog.String("error", err.Error()))
			http.Error(w, "Invalid end date format", http.StatusBadRequest)
			return
		}
		formattedDate := date.Format(time.RFC3339)
		payload.EndDate = &formattedDate
	}

	if status := r.FormValue("status"); status != "" {
		payload.Status = &status
	}

	if recurType := r.FormValue("recur_type"); recurType != "" {
		payload.RecurType = &recurType
	}

	if frequency := r.FormValue("frequency"); frequency != "" {
		freq, err := strconv.Atoi(frequency)
		if err != nil {
			slog.Error("failed to parse frequency", slog.String("error", err.Error()))
			http.Error(w, "Invalid frequency", http.StatusBadRequest)
			return
		}
		payload.Frequency = &freq
	}

	if dayOfWeek := r.FormValue("day_of_week"); dayOfWeek != "" {
		dow, err := strconv.Atoi(dayOfWeek)
		if err == nil && dow >= 0 && dow <= 6 {
			payload.DayOfWeek = &dow
		}
	}

	if dayOfMonth := r.FormValue("day_of_month"); dayOfMonth != "" {
		dom, err := strconv.Atoi(dayOfMonth)
		if err == nil && dom >= 1 && dom <= 31 {
			payload.DayOfMonth = &dom
		}
	}

	if monthOfYear := r.FormValue("month_of_year"); monthOfYear != "" {
		moy, err := strconv.Atoi(monthOfYear)
		if err == nil && moy >= 1 && moy <= 12 {
			payload.MonthOfYear = &moy
		}
	}

	if err := s.sendRequest(r, "PUT", fmt.Sprintf("/v1/recurring/%d", id), payload, nil); err != nil {
		slog.Error("failed to update recurring transaction", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to update recurring transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadRecurring")
	w.Header().Set("HX-Redirect", "/page/recurring")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) deleteRecurring(w http.ResponseWriter, r *http.Request) {
	id := parseInt32(r.PathValue("recurring_id"))

	if err := s.sendRequest(r, "DELETE", fmt.Sprintf("/v1/recurring/%d", id), nil, nil); err != nil {
		slog.Error("failed to delete recurring transaction", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to delete recurring transaction", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadRecurring")
	w.Header().Set("HX-Redirect", "/page/recurring")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) markReminderAsRead(w http.ResponseWriter, r *http.Request) {
	id := parseInt32(r.PathValue("reminder_id"))

	if err := s.sendRequest(r, "POST", fmt.Sprintf("/v1/recurring/reminders/%d/read", id), nil, nil); err != nil {
		slog.Error("failed to mark reminder as read", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to mark reminder as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Trigger", "reloadReminders")
	w.WriteHeader(http.StatusOK)
}