package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/omegaatt36/bookly/app/web/api"
	"github.com/omegaatt36/bookly/app/web/templates/components"
	"github.com/omegaatt36/bookly/app/web/templates/layout"
	"github.com/omegaatt36/bookly/app/web/templates/pages"
	"github.com/omegaatt36/bookly/domain"
)

// parseInt32 converts a string to int32 safely, returning 0 if conversion fails
func parseInt32(s string) int32 {
	if s == "" {
		slog.Debug("empty string provided for int32 conversion")
		return 0
	}
	fmt.Printf("parsing string to int32: '%s'\n", s)
	val, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		slog.Debug("failed to parse string to int32", slog.String("value", s), slog.String("error", err.Error()))
		return 0
	}

	return int32(val)
}

// Server represents a web server
type Server struct {
	port      int
	router    http.Handler
	templates *template.Template

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

	var response struct {
		Code    int             `json:"code"`
		Data    json.RawMessage `json:"data"`
		Message string          `json:"message"`
	}

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
	var accounts []api.Account
	err := s.sendRequest(r, "GET", "/v1/accounts", nil, &accounts)
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
	var account api.Account
	err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/accounts/%d", accountID), nil, &account)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get ledgers for this account
	var ledger []api.Ledger
	err = s.sendRequest(r, "GET", fmt.Sprintf("/v1/accounts/%d/ledgers", accountID), nil, &ledger)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	component := pages.AccountDetail(user, account, ledger)
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
	var recurringTransactions []api.RecurringTransaction
	err := s.sendRequest(r, "GET", "/v1/recurring", nil, &recurringTransactions)
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
	var reminders []api.Reminder
	err := s.sendRequest(r, "GET", "/v1/recurring/reminders", nil, &reminders)
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

	var account api.Account
	err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/accounts/%d", accountID), nil, &account)
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
	var accounts []api.Account
	err := s.sendRequest(r, "GET", "/v1/accounts", nil, &accounts)
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
	var accounts []api.Account
	err := s.sendRequest(r, "GET", "/v1/accounts", nil, &accounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	component := components.LedgerForm(accounts, "", 0)
	component.Render(r.Context(), w)
}

func (s *Server) modalRecurringForm(w http.ResponseWriter, r *http.Request) {
	var accounts []api.Account
	err := s.sendRequest(r, "GET", "/v1/accounts", nil, &accounts)
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
	var accounts []api.Account
	err := s.sendRequest(r, "GET", "/v1/accounts", nil, &accounts)
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

	err := s.sendRequest(r, "POST", "/public/auth/login", loginReq, &loginResp)
	if err != nil {
		// Return error toast
		component := layout.Toast("登入失敗："+err.Error(), "error")
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
		component := layout.Toast("密碼不一致", "error")
		component.Render(r.Context(), w)
		return
	}

	registerReq := map[string]string{
		"email":    email,
		"password": password,
	}

	err := s.sendRequest(r, "POST", "/public/auth/register", registerReq, nil)
	if err != nil {
		component := layout.Toast("註冊失敗："+err.Error(), "error")
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
	var account domain.Account
	err := s.sendRequest(r, "POST", "/v1/accounts", req, &account)
	if err != nil {
		component := layout.Toast("新增帳戶失敗："+err.Error(), "error")
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
	currency := r.FormValue("currency")

	req := map[string]any{
		"id": accountID,
	}
	if name != "" {
		req["name"] = name
	}
	if status != "" {
		req["status"] = status
	}
	if currency != "" {
		req["currency"] = currency
	}

	var account domain.Account
	err := s.sendRequest(r, "PUT", "/v1/accounts/"+strconv.Itoa(int(accountID)), req, &account)
	if err != nil {
		component := layout.Toast("更新帳戶失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Redirect", "/accounts")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleActivateAccount(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.PathValue("id"))
	req := map[string]any{
		"status": "active",
	}
	var account domain.Account
	err := s.sendRequest(r, "PUT", "/v1/accounts/"+strconv.Itoa(int(accountID)), req, &account)
	if err != nil {
		component := layout.Toast("帳戶啟用失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeactivateAccount(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.PathValue("id"))
	req := map[string]any{
		"status": "closed",
	}
	var account domain.Account
	err := s.sendRequest(r, "PUT", "/v1/accounts/"+strconv.Itoa(int(accountID)), req, &account)
	if err != nil {
		component := layout.Toast("帳戶停用失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleCreateLedger(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.FormValue("account_id"))
	date := r.FormValue("date")
	tType := r.FormValue("type")
	amount := r.FormValue("amount")
	note := r.FormValue("note")

	req := map[string]any{
		"account_id": accountID,
		"date":       date,
		"type":       tType,
		"amount":     amount,
	}
	if note != "" {
		req["note"] = note
	}
	var ledger domain.Ledger
	err := s.sendRequest(r, "POST", "/v1/accounts/"+strconv.Itoa(int(accountID))+"/ledgers", req, &ledger)
	if err != nil {
		component := layout.Toast("新增交易失敗："+err.Error(), "error")
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

	req := map[string]any{
		"id": ledgerID,
	}
	if date != "" {
		req["date"] = date
	}
	if tType != "" {
		req["type"] = tType
	}
	if amount != "" {
		req["amount"] = amount
	}
	if note != "" {
		req["note"] = note
	}

	var ledger domain.Ledger
	err := s.sendRequest(r, "PUT", "/v1/ledgers/"+strconv.Itoa(int(ledgerID)), req, &ledger)
	if err != nil {
		component := layout.Toast("更新交易失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleVoidLedger(w http.ResponseWriter, r *http.Request) {
	// 作廢交易
	ledgerID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, "POST", "/v1/ledgers/"+strconv.Itoa(int(ledgerID))+"/void", nil, nil)
	if err != nil {
		component := layout.Toast("作廢失敗："+err.Error(), "error")
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
	endDate := r.FormValue("end_date")
	recurType := r.FormValue("recur_type")
	frequency := r.FormValue("frequency")
	dayOfWeek := r.FormValue("day_of_week")
	dayOfMonth := r.FormValue("day_of_month")
	monthOfYear := r.FormValue("month_of_year")

	req := map[string]any{
		"account_id": accountID,
		"name":       name,
		"type":       tType,
		"amount":     amount,
		"start_date": startDate,
		"recur_type": recurType,
		"frequency":  frequency,
	}
	if note != "" {
		req["note"] = note
	}
	if endDate != "" {
		req["end_date"] = endDate
	}
	if dayOfWeek != "" {
		req["day_of_week"] = dayOfWeek
	}
	if dayOfMonth != "" {
		req["day_of_month"] = dayOfMonth
	}
	if monthOfYear != "" {
		req["month_of_year"] = monthOfYear
	}
	var tx domain.RecurringTransaction
	err := s.sendRequest(r, "POST", "/v1/recurring", req, &tx)
	if err != nil {
		component := layout.Toast("建立定期交易失敗："+err.Error(), "error")
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
	endDate := r.FormValue("end_date")
	recurType := r.FormValue("recur_type")
	status := r.FormValue("status")
	frequency := r.FormValue("frequency")
	dayOfWeek := r.FormValue("day_of_week")
	dayOfMonth := r.FormValue("day_of_month")
	monthOfYear := r.FormValue("month_of_year")

	req := map[string]any{
		"id": recurringID,
	}
	if name != "" {
		req["name"] = name
	}
	if tType != "" {
		req["type"] = tType
	}
	if amount != "" {
		req["amount"] = amount
	}
	if note != "" {
		req["note"] = note
	}
	if endDate != "" {
		req["end_date"] = endDate
	}
	if recurType != "" {
		req["recur_type"] = recurType
	}
	if status != "" {
		req["status"] = status
	}
	if frequency != "" {
		req["frequency"] = frequency
	}
	if dayOfWeek != "" {
		req["day_of_week"] = dayOfWeek
	}
	if dayOfMonth != "" {
		req["day_of_month"] = dayOfMonth
	}
	if monthOfYear != "" {
		req["month_of_year"] = monthOfYear
	}

	var tx domain.RecurringTransaction
	err := s.sendRequest(r, "PUT", "/v1/recurring/"+strconv.Itoa(int(recurringID)), req, &tx)
	if err != nil {
		component := layout.Toast("更新定期交易失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Redirect", "/recurring")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handlePauseRecurring(w http.ResponseWriter, r *http.Request) {
	recurringID := parseInt32(r.PathValue("id"))
	req := map[string]any{
		"status": "paused",
	}
	var tx domain.RecurringTransaction
	err := s.sendRequest(r, "PUT", "/v1/recurring/"+strconv.Itoa(int(recurringID)), req, &tx)
	if err != nil {
		component := layout.Toast("暫停失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleResumeRecurring(w http.ResponseWriter, r *http.Request) {
	recurringID := parseInt32(r.PathValue("id"))
	req := map[string]any{
		"status": "active",
	}
	var tx domain.RecurringTransaction
	err := s.sendRequest(r, "PUT", "/v1/recurring/"+strconv.Itoa(int(recurringID)), req, &tx)
	if err != nil {
		component := layout.Toast("恢復失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeleteRecurring(w http.ResponseWriter, r *http.Request) {
	recurringID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, "DELETE", "/v1/recurring/"+strconv.Itoa(int(recurringID)), nil, nil)
	if err != nil {
		component := layout.Toast("刪除失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleMarkReminderRead(w http.ResponseWriter, r *http.Request) {
	reminderID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, "POST", "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID))+"/read", nil, nil)
	if err != nil {
		component := layout.Toast("標記失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleMarkReminderUnread(w http.ResponseWriter, r *http.Request) {
	reminderID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, "POST", "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID))+"/unread", nil, nil)
	if err != nil {
		component := layout.Toast("標記失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleExecuteReminder(w http.ResponseWriter, r *http.Request) {
	reminderID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, "POST", "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID))+"/execute", nil, nil)
	if err != nil {
		component := layout.Toast("執行提醒失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleSnoozeReminder(w http.ResponseWriter, r *http.Request) {
	reminderID := parseInt32(r.PathValue("id"))
	snoozeMinutes := r.FormValue("minutes")
	body := map[string]any{}
	if snoozeMinutes != "" {
		body["minutes"] = snoozeMinutes
	}
	err := s.sendRequest(r, "POST", "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID))+"/snooze", body, nil)
	if err != nil {
		component := layout.Toast("延後失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleDeleteReminder(w http.ResponseWriter, r *http.Request) {
	reminderID := parseInt32(r.PathValue("id"))
	err := s.sendRequest(r, "DELETE", "/v1/recurring/reminders/"+strconv.Itoa(int(reminderID)), nil, nil)
	if err != nil {
		component := layout.Toast("刪除失敗："+err.Error(), "error")
		component.Render(r.Context(), w)
		return
	}
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleMarkAllRemindersRead(w http.ResponseWriter, r *http.Request) {
	// err := s.sendRequest(r, "POST", "/v1/recurring/reminders/mark-all-read", nil, nil)
	// if err != nil {
	// 	component := layout.Toast("批次標記失敗："+err.Error(), "error")
	// 	component.Render(r.Context(), w)
	// 	return
	// }
	// w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusOK)
}

// API handlers for dashboard data
func (s *Server) apiAccountsSummary(w http.ResponseWriter, r *http.Request) {
	var summary any
	// err := s.sendRequest(r, "GET", "/v1/accounts/summary", nil, &summary)
	// if err != nil {
	// 	http.Error(w, "取得失敗: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (s *Server) apiRecentLedgers(w http.ResponseWriter, r *http.Request) {
	var ledgers any
	// err := s.sendRequest(r, "GET", "/v1/ledgers/recent", nil, &ledgers)
	// if err != nil {
	// 	http.Error(w, "取得失敗: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ledgers)
}

func (s *Server) apiMonthlyStatistics(w http.ResponseWriter, r *http.Request) {
	var stats any
	// err := s.sendRequest(r, "GET", "/v1/statistics/monthly", nil, &stats)
	// if err != nil {
	// 	http.Error(w, "取得失敗: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (s *Server) apiUpcomingReminders(w http.ResponseWriter, r *http.Request) {
	var reminders any
	// err := s.sendRequest(r, "GET", "/v1/recurring/reminders/upcoming", nil, &reminders)
	// if err != nil {
	// 	http.Error(w, "取得失敗: "+err.Error(), http.StatusInternalServerError)
	// 	return
	// }
	// w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reminders)
}
