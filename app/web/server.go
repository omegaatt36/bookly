package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/omegaatt36/bookly/app/web/templates/components"
	"github.com/omegaatt36/bookly/app/web/templates/layout"
	"github.com/omegaatt36/bookly/app/web/templates/pages"
	"github.com/omegaatt36/bookly/sdk/datatype"
)

// parseInt32 converts a string to int32 safely, returning 0 if conversion fails
func parseInt32(s string) int32 {
	if s == "" {
		slog.Warn("empty string provided for int32 conversion")
		return 0
	}

	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		slog.Debug("failed to parse string to int32", slog.String("value", s), slog.String("error", err.Error()))
		return 0
	}

	return int32(val)
}

// Server represents a web server
type Server struct {
	port   int
	router http.Handler

	serverURL string
}

// NewServer creates a new web server
func NewServer(options ...Option) *Server {
	server := &Server{
		serverURL: "http://localhost:8080",
		port:      3000,
	}

	for _, option := range options {
		option.apply(server)
	}

	server.initTemplates()
	server.registerRoutes()

	return server
}

func (s *Server) initTemplates() {
	// Templates are now handled by templ - no manual template initialization needed
}

// Run starts the server.
func (s *Server) Run(ctx context.Context) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	go func() {
		<-ctx.Done()
		if err := srv.Shutdown(ctx); err != nil {
			slog.Error("server shutdown error", slog.String("error", err.Error()))
		}
	}()

	slog.Info("starting web server", slog.String("addr", srv.Addr))

	if err := srv.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", slog.String("error", err.Error()))
	}
}

type sendRequestError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *sendRequestError) Error() string {
	return fmt.Sprintf("failed to send request: %s", e.Message)
}

func (s *Server) sendRequest(r *http.Request, method, path string, body any, result any) error {
	url := fmt.Sprintf("%s%s", s.serverURL, path)
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if !strings.HasPrefix(path, "/public") {
		token, err := r.Cookie("token")
		if err != nil {
			return fmt.Errorf("failed to get token from cookie: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token.Value)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var response datatype.Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Code != 0 {
		return &sendRequestError{
			Code:    response.Code,
			Message: fmt.Sprintf("failed to send request: %s", response.Message),
		}
	}

	if result == nil {
		return nil
	}

	if err := json.Unmarshal(response.Data, result); err != nil {
		return fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return nil
}

// Page handlers
func (s *Server) pageLogin(w http.ResponseWriter, r *http.Request) {
	component := pages.Login()
	component.Render(r.Context(), w)
}

func (s *Server) pageRegister(w http.ResponseWriter, r *http.Request) {
	component := pages.Register()
	component.Render(r.Context(), w)
}

func (s *Server) pageHome(w http.ResponseWriter, r *http.Request) {
	user := s.getUserFromContext(r)
	component := pages.Home(user)
	component.Render(r.Context(), w)
}

func (s *Server) pageAccounts(w http.ResponseWriter, r *http.Request) {
	user := s.getUserFromContext(r)
	// Get accounts from API
	var accounts []datatype.Account
	err := s.sendRequest(r, http.MethodGet, "/v1/accounts", nil, &accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	component := pages.AccountsList(user, accounts)
	component.Render(r.Context(), w)
}

func (s *Server) pageAccountDetail(w http.ResponseWriter, r *http.Request) {
	user := s.getUserFromContext(r)
	accountID := parseInt32(r.PathValue("id"))

	// Get account details
	var account datatype.Account
	err := s.sendRequest(r, http.MethodGet, fmt.Sprintf("/v1/accounts/%d", accountID), nil, &account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get ledgers for this account
	var ledgers []datatype.Ledger
	err = s.sendRequest(r, http.MethodGet, fmt.Sprintf("/v1/accounts/%d/ledgers", accountID), nil, &ledgers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	component := pages.AccountDetail(user, account, ledgers)
	component.Render(r.Context(), w)
}

func (s *Server) pageAccountLedgers(w http.ResponseWriter, r *http.Request) {
	// Redirect to account detail page
	accountID := r.PathValue("id")
	http.Redirect(w, r, "/accounts/"+accountID, http.StatusMovedPermanently)
}

func (s *Server) pageRecurring(w http.ResponseWriter, r *http.Request) {
	user := s.getUserFromContext(r)

	// Get recurring transactions
	var recurringTransactions []datatype.RecurringTransaction
	err := s.sendRequest(r, http.MethodGet, "/v1/recurring", nil, &recurringTransactions)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	component := pages.RecurringTransactions(user, recurringTransactions)
	component.Render(r.Context(), w)
}

func (s *Server) pageReminders(w http.ResponseWriter, r *http.Request) {
	user := s.getUserFromContext(r)

	// Get reminders
	var reminders []datatype.Reminder
	err := s.sendRequest(r, http.MethodGet, "/v1/recurring/reminders", nil, &reminders)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	component := pages.Reminders(user, reminders)
	component.Render(r.Context(), w)
}

// Modal handlers
func (s *Server) modalAccountForm(w http.ResponseWriter, r *http.Request) {
	currencies := []string{"TWD", "USD", "EUR", "JPY"}
	component := components.AccountForm(nil, currencies)
	component.Render(r.Context(), w)
}

func (s *Server) modalAccountEdit(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.PathValue("id"))

	var account datatype.Account
	err := s.sendRequest(r, http.MethodGet, fmt.Sprintf("/v1/accounts/%d", accountID), nil, &account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	currencies := []string{"TWD", "USD", "EUR", "JPY"}
	component := components.AccountForm(&account, currencies)
	component.Render(r.Context(), w)
}

func (s *Server) modalLedgerForm(w http.ResponseWriter, r *http.Request) {
	// Get accounts for dropdown
	var accounts []datatype.Account
	err := s.sendRequest(r, http.MethodGet, "/v1/accounts", nil, &accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ledgerType := r.URL.Query().Get("type")
	accountID := parseInt32(r.URL.Query().Get("account_id"))

	component := components.LedgerForm(accounts, ledgerType, accountID)
	component.Render(r.Context(), w)
}

func (s *Server) modalLedgerEdit(w http.ResponseWriter, r *http.Request) {
	// Similar to modalLedgerForm but with existing ledger data
	var accounts []datatype.Account
	err := s.sendRequest(r, http.MethodGet, "/v1/accounts", nil, &accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	component := components.LedgerForm(accounts, "", 0) // Placeholder, actual data loading needed
	component.Render(r.Context(), w)
}

func (s *Server) modalRecurringForm(w http.ResponseWriter, r *http.Request) {
	var accounts []datatype.Account
	err := s.sendRequest(r, http.MethodGet, "/v1/accounts", nil, &accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recurrenceTypes := []string{
		"daily",
		"weekly",
		"monthly",
		"yearly",
	}

	component := components.RecurringForm(accounts, recurrenceTypes)
	component.Render(r.Context(), w)
}

func (s *Server) modalRecurringEdit(w http.ResponseWriter, r *http.Request) {
	var accounts []datatype.Account
	err := s.sendRequest(r, http.MethodGet, "/v1/accounts", nil, &accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	recurrenceTypes := []string{
		"daily",
		"weekly",
		"monthly",
		"yearly",
	}

	component := components.RecurringForm(accounts, recurrenceTypes) // Placeholder, actual data loading needed
	component.Render(r.Context(), w)
}

// Auth handlers
func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")

	loginReq := map[string]string{
		"email":    email,
		"password": password,
	}

	var loginResp struct {
		Token string `json:"token"`
	}

	err := s.sendRequest(r, http.MethodPost, "/public/auth/login", loginReq, &loginResp)
	if err != nil {
		// Return error toast
		component := layout.Toast("Login failed: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}

	// Set cookie and redirect
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    loginResp.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("HX-Redirect", "/")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if password != confirmPassword {
		component := layout.Toast("Passwords do not match", "error")
		component.Render(r.Context(), w)
		return
	}

	registerReq := map[string]string{
		"email":    email,
		"password": password,
	}

	err := s.sendRequest(r, http.MethodPost, "/public/auth/register", registerReq, nil)
	if err != nil {
		component := layout.Toast("Registration failed: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}

	w.Header().Set("HX-Redirect", "/auth/login")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.Header().Set("HX-Redirect", "/auth/login")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	// user := s.getUserFromContext(r)
	name := r.FormValue("name")
	currency := r.FormValue("currency")
	req := map[string]any{
		"name":     name,
		"currency": currency,
	}
	var account datatype.Account
	err := s.sendRequest(r, http.MethodPost, "/v1/accounts", req, &account)
	if err != nil {
		component := layout.Toast("Failed to create account: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Redirect", "/accounts")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUpdateAccount(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.PathValue("id"))
	name := r.FormValue("name")
	status := r.FormValue("status")

	req := datatype.UpdateAccountRequest{}

	if name != "" {
		req.Name = &name
	}
	if status != "" {
		req.Status = &status
	}

	var account datatype.Account
	err := s.sendRequest(r, http.MethodPut, fmt.Sprintf("/v1/accounts/%d", accountID), req, &account)
	if err != nil {
		component := layout.Toast("Failed to update account: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Redirect", "/accounts")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleActivateAccount(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.PathValue("id"))

	req := datatype.UpdateAccountRequest{
		Status: func() *string {
			status := "active"
			return &status
		}(),
	}

	var account datatype.Account
	err := s.sendRequest(r, http.MethodPut, "/v1/accounts/"+strconv.Itoa(int(accountID)), req, &account)
	if err != nil {
		component := layout.Toast("Failed to activate account: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeactivateAccount(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.PathValue("id"))
	req := datatype.UpdateAccountRequest{
		Status: func() *string {
			status := "closed"
			return &status
		}(),
	}

	var account datatype.Account
	err := s.sendRequest(r, http.MethodPut, "/v1/accounts/"+strconv.Itoa(int(accountID)), req, &account)
	if err != nil {
		component := layout.Toast("Failed to deactivate account: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCreateLedger(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.FormValue("account_id"))

	req := datatype.CreateLedgerRequest{
		AccountID: accountID,
		Date:      r.FormValue("date"),
		Type:      r.FormValue("type"),
		Amount:    r.FormValue("amount"),
		Note:      r.FormValue("note"),
	}

	var ledger datatype.Ledger
	err := s.sendRequest(r, http.MethodPost, "/v1/accounts/"+strconv.Itoa(int(accountID))+"/ledgers", req, &ledger)
	if err != nil {
		component := layout.Toast("Failed to create transaction: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Redirect", "/accounts/"+strconv.Itoa(int(accountID)))
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUpdateLedger(w http.ResponseWriter, r *http.Request) {
	ledgerID := parseInt32(r.PathValue("id"))
	date := r.FormValue("date")
	tType := r.FormValue("type")
	amount := r.FormValue("amount")
	note := r.FormValue("note")

	req := datatype.UpdateLedgerRequest{
		ID: ledgerID,
	}

	if date != "" {
		req.Date = &date
	}
	if tType != "" {
		req.Type = &tType
	}
	if amount != "" {
		req.Amount = &amount
	}
	if note != "" {
		req.Note = &note
	}

	var ledger datatype.Ledger
	err := s.sendRequest(r, http.MethodPut, "/v1/ledgers/"+strconv.Itoa(int(ledgerID)), req, &ledger)
	if err != nil {
		component := layout.Toast("Failed to update transaction: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleVoidLedger(w http.ResponseWriter, r *http.Request) {
	// Void transaction
	ledgerID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, http.MethodDelete, "/v1/ledgers/"+strconv.Itoa(int(ledgerID)), nil, nil)
	if err != nil {
		component := layout.Toast("Failed to void transaction: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCreateRecurring(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.FormValue("account_id"))
	name := r.FormValue("name")
	tType := r.FormValue("type")
	amount := r.FormValue("amount")
	note := r.FormValue("note")
	startDate := r.FormValue("start_date")
	endDateStr := r.FormValue("end_date")
	recurType := r.FormValue("recur_type")
	frequencyStr := r.FormValue("frequency")
	dayOfWeekStr := r.FormValue("day_of_week")
	dayOfMonthStr := r.FormValue("day_of_month")
	monthOfYearStr := r.FormValue("month_of_year")

	// Parse optional fields into pointers as needed
	var (
		endDate     *string
		dayOfWeek   *int
		dayOfMonth  *int
		monthOfYear *int
	)

	if endDateStr != "" {
		endDate = &endDateStr
	}
	if dayOfWeekStr != "" {
		val, err := strconv.Atoi(dayOfWeekStr)
		if err == nil {
			dayOfWeek = &val
		}
	}
	if dayOfMonthStr != "" {
		val, err := strconv.Atoi(dayOfMonthStr)
		if err == nil {
			dayOfMonth = &val
		}
	}
	if monthOfYearStr != "" {
		val, err := strconv.Atoi(monthOfYearStr)
		if err == nil {
			monthOfYear = &val
		}
	}

	frequency := 1
	if frequencyStr != "" {
		if freqInt, err := strconv.Atoi(frequencyStr); err == nil {
			frequency = freqInt
		}
	}

	req := datatype.CreateRecurringTransactionRequest{
		AccountID:   accountID,
		Name:        name,
		Type:        tType,
		Amount:      amount,
		Note:        note,
		StartDate:   startDate,
		EndDate:     endDate,
		RecurType:   recurType,
		Frequency:   frequency,
		DayOfWeek:   dayOfWeek,
		DayOfMonth:  dayOfMonth,
		MonthOfYear: monthOfYear,
	}

	var tx datatype.RecurringTransaction
	err := s.sendRequest(r, http.MethodPost, "/v1/recurring", req, &tx)
	if err != nil {
		component := layout.Toast("Failed to create recurring transaction: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Redirect", "/recurring")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUpdateRecurring(w http.ResponseWriter, r *http.Request) {
	recurringID := parseInt32(r.PathValue("id"))
	name := r.FormValue("name")
	tType := r.FormValue("type")
	amount := r.FormValue("amount")
	note := r.FormValue("note")
	endDateStr := r.FormValue("end_date")
	recurType := r.FormValue("recur_type")
	status := r.FormValue("status")
	frequencyStr := r.FormValue("frequency")
	dayOfWeekStr := r.FormValue("day_of_week")
	dayOfMonthStr := r.FormValue("day_of_month")
	monthOfYearStr := r.FormValue("month_of_year")

	var (
		pName        *string
		pType        *string
		pAmount      *string
		pNote        *string
		pEndDate     *string
		pRecurType   *string
		pStatus      *string
		pFrequency   *int
		pDayOfWeek   *int
		pDayOfMonth  *int
		pMonthOfYear *int
	)

	if name != "" {
		pName = &name
	}
	if tType != "" {
		pType = &tType
	}
	if amount != "" {
		pAmount = &amount
	}
	if note != "" {
		pNote = &note
	}
	if endDateStr != "" {
		pEndDate = &endDateStr
	}
	if recurType != "" {
		pRecurType = &recurType
	}
	if status != "" {
		pStatus = &status
	}
	if frequencyStr != "" {
		if freqInt, err := strconv.Atoi(frequencyStr); err == nil {
			pFrequency = &freqInt
		}
	}
	if dayOfWeekStr != "" {
		if val, err := strconv.Atoi(dayOfWeekStr); err == nil {
			pDayOfWeek = &val
		}
	}
	if dayOfMonthStr != "" {
		if val, err := strconv.Atoi(dayOfMonthStr); err == nil {
			pDayOfMonth = &val
		}
	}
	if monthOfYearStr != "" {
		if val, err := strconv.Atoi(monthOfYearStr); err == nil {
			pMonthOfYear = &val
		}
	}

	req := datatype.UpdateRecurringTransactionRequest{
		ID:          recurringID,
		Name:        pName,
		Type:        pType,
		Amount:      pAmount,
		Note:        pNote,
		EndDate:     pEndDate,
		RecurType:   pRecurType,
		Status:      pStatus,
		Frequency:   pFrequency,
		DayOfWeek:   pDayOfWeek,
		DayOfMonth:  pDayOfMonth,
		MonthOfYear: pMonthOfYear,
	}

	var tx datatype.RecurringTransaction
	err := s.sendRequest(r, http.MethodPut, "/v1/recurring/"+strconv.Itoa(int(recurringID)), req, &tx)
	if err != nil {
		component := layout.Toast("Failed to update recurring transaction: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Redirect", "/recurring")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePauseRecurring(w http.ResponseWriter, r *http.Request) {
	recurringID := parseInt32(r.PathValue("id"))

	req := datatype.UpdateRecurringTransactionRequest{
		Status: func() *string {
			status := "paused"
			return &status
		}(),
	}

	var tx datatype.RecurringTransaction
	err := s.sendRequest(r, http.MethodPut, "/v1/recurring/"+strconv.Itoa(int(recurringID)), req, &tx)
	if err != nil {
		component := layout.Toast("Failed to pause recurring transaction: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleResumeRecurring(w http.ResponseWriter, r *http.Request) {
	recurringID := parseInt32(r.PathValue("id"))

	req := datatype.UpdateRecurringTransactionRequest{
		Status: func() *string {
			status := "active"
			return &status
		}(),
	}

	var tx datatype.RecurringTransaction
	err := s.sendRequest(r, http.MethodPut, "/v1/recurring/"+strconv.Itoa(int(recurringID)), req, &tx)
	if err != nil {
		component := layout.Toast("Failed to resume recurring transaction: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeleteRecurring(w http.ResponseWriter, r *http.Request) {
	recurringID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, http.MethodDelete, "/v1/recurring/"+strconv.Itoa(int(recurringID)), nil, nil)
	if err != nil {
		component := layout.Toast("Failed to delete recurring transaction: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleMarkReminderRead(w http.ResponseWriter, r *http.Request) {
	reminderID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, http.MethodPost, "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID))+"/read", nil, nil)
	if err != nil {
		component := layout.Toast("Failed to mark reminder as read: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleMarkReminderUnread(w http.ResponseWriter, r *http.Request) {
	reminderID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, http.MethodPost, "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID))+"/unread", nil, nil)
	if err != nil {
		component := layout.Toast("Failed to mark reminder as unread: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleExecuteReminder(w http.ResponseWriter, r *http.Request) {
	reminderID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, http.MethodPost, "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID))+"/execute", nil, nil)
	if err != nil {
		component := layout.Toast("Failed to execute reminder: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSnoozeReminder(w http.ResponseWriter, r *http.Request) {
	// reminderID := parseInt32(r.PathValue("id"))
	// snoozeMinutes := r.FormValue("minutes")
	// body := map[string]any{}
	// if snoozeMinutes != "" {
	// 	body["minutes"] = snoozeMinutes
	// }
	// err := s.sendRequest(r, http.MethodPost, "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID))+"/snooze", body, nil)
	// if err != nil {
	// 	component := layout.Toast("Failed to snooze reminder: "+err.Error(), "error")
	// 	component.Render(r.Context(), w)
	// 	return
	// }
	// w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeleteReminder(w http.ResponseWriter, r *http.Request) {
	reminderID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, http.MethodDelete, "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID)), nil, nil)
	if err != nil {
		component := layout.Toast("Failed to delete reminder: "+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleMarkAllRemindersRead(w http.ResponseWriter, r *http.Request) {
	// err := s.sendRequest(r, http.MethodPost, "/v1/recurring/reminders/mark-all-read", nil, nil)
	// if err != nil {
	// component := layout.Toast("Failed to mark all reminders as read: "+err.Error(), "error")
	// component.Render(r.Context(), w)
	// return
	// }
	// w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

// API handlers (for HTMX partial updates)
func (s *Server) apiAccountsSummary(w http.ResponseWriter, r *http.Request) {
	// Fetch account summary data and render as HTML partial
	// Example:
	// accounts := []struct{ Name string; Balance string }{{\"Bank\", \"$1,234.56\"}, {\"Credit Card\", \"$567.89\"}}
	// component := components.AccountSummary(accounts)
	// component.Render(r.Context(), w)
	fmt.Fprint(w, "<p>Account summary loaded</p>")
}

func (s *Server) apiRecentLedgers(w http.ResponseWriter, r *http.Request) {
	// Fetch recent ledgers and render as HTML partial
	// Example:
	// ledgers := []struct{ Description string; Amount string; Date string }{{\"Salary\", \"+$2,000.00\", \"2024-07-15\"}, {\"Groceries\", \"-$50.00\", \"2024-07-14\"}}
	// component := components.RecentLedgers(ledgers)
	// component.Render(r.Context(), w)
	fmt.Fprint(w, "<p>Recent transactions loaded</p>")
}

func (s *Server) apiMonthlyStatistics(w http.ResponseWriter, r *http.Request) {
	// Fetch monthly statistics and render as HTML partial
	// Example:
	// stats := struct{ Income string; Expenses string; Net string }{Income: \"$3,000.00\", Expenses: \"$1,500.00\", Net: \"$1,500.00\"}
	// component := components.MonthlyStats(stats)
	// component.Render(r.Context(), w)
	fmt.Fprint(w, "<p>Monthly statistics loaded</p>")
}

func (s *Server) apiUpcomingReminders(w http.ResponseWriter, r *http.Request) {
	// Fetch upcoming reminders and render as HTML partial
	// Example:
	// reminders := []struct{ Title string; DueDate string }{{\"Rent Payment\", \"2024-08-01\"}, {\"Subscription Renewal\", \"2024-07-20\"}}
	// component := components.UpcomingReminders(reminders)
	// component.Render(r.Context(), w)
	fmt.Fprint(w, "<p>Upcoming reminders loaded</p>")
}
