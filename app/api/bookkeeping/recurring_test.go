package bookkeeping_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/suite"

	"github.com/omegaatt36/bookly/app/api/bookkeeping"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/migration"
	"github.com/omegaatt36/bookly/persistence/repository" // Assuming SQLCRepository is here
)

type testRecurringSuite struct {
	suite.Suite

	router *http.ServeMux

	// Assuming SQLCRepository implements all necessary interfaces (Account, Ledger, RecurringTransaction, Reminder)
	repo      *repository.SQLCRepository
	finalize  func()
	userID    string
	accountID string
}

func (s *testRecurringSuite) SetupTest() {
	s.finalize = database.TestingInitialize(database.PostgresOpt)
	db := database.GetDB()
	s.repo = repository.NewSQLCRepository(db) // Assuming SQLCRepository implements all necessary interfaces
	s.router = http.NewServeMux()

	// Pass all repositories to the controller
	controller := bookkeeping.NewController(bookkeeping.NewControllerRequest{
		AccountRepository:              s.repo,
		LedgerRepository:               s.repo,
		RecurringTransactionRepository: s.repo, // Use s.repo for recurring
		ReminderRepository:             s.repo, // Use s.repo for reminders
	})

	// Authentication middleware
	authMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// In a real auth middleware, you would validate the token/session
			// For testing, we just set a dummy user ID from the suite
			ctx := engine.WithUserID(r.Context(), s.userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	// Helper to register handlers with authentication
	registerWithAuth := func(pattern string, handler http.Handler) {
		s.router.Handle(pattern, authMiddleware(handler))
	}

	// Register recurring transaction and reminder routes
	registerWithAuth("POST /recurring", http.HandlerFunc(controller.CreateRecurringTransaction()))
	registerWithAuth("GET /recurring", http.HandlerFunc(controller.GetRecurringTransactions()))
	registerWithAuth("GET /recurring/{id}", http.HandlerFunc(controller.GetRecurringTransaction()))
	registerWithAuth("PUT /recurring/{id}", http.HandlerFunc(controller.UpdateRecurringTransaction()))
	registerWithAuth("DELETE /recurring/{id}", http.HandlerFunc(controller.DeleteRecurringTransaction()))
	registerWithAuth("GET /recurring/reminders", http.HandlerFunc(controller.GetReminders()))
	registerWithAuth("POST /recurring/reminders/{id}/read", http.HandlerFunc(controller.MarkReminderAsRead()))

	s.NoError(migration.NewMigrator(db).Upgrade())

	// Create a seed user and account for tests
	userID, err := s.createSeedUser()
	s.NoError(err)
	s.userID = userID

	accountID, err := s.createSeedAccount(userID)
	s.NoError(err)
	s.accountID = accountID
}

func (s *testRecurringSuite) TearDownTest() {
	s.finalize()
	s.router = nil
	s.repo = nil
	s.userID = ""
	s.accountID = ""
}

func TestRecurringSuite(t *testing.T) {
	suite.Run(t, new(testRecurringSuite))
}

func (s *testRecurringSuite) createSeedUser() (string, error) {
	return s.repo.CreateUser(domain.CreateUserRequest{
		Name: "Recurring Test User",
	})
}

func (s *testRecurringSuite) createSeedAccount(userID string) (string, error) {
	s.NoError(s.repo.CreateAccount(domain.CreateAccountRequest{
		UserID:   userID,
		Name:     "Recurring Test Account",
		Currency: "USD",
	}))

	accounts, err := s.repo.GetAccountsByUserID(userID)
	if err != nil {
		return "", err
	}
	if len(accounts) == 0 {
		return "", fmt.Errorf("failed to create seed account")
	}
	return accounts[0].ID, nil
}

func (s *testRecurringSuite) createSeedRecurringTransaction(accountID string, recurType domain.RecurrenceType, amount decimal.Decimal) (*domain.RecurringTransaction, error) {
	// Set start date in the past to ensure NextDue calculation
	startDate := time.Now().Add(-24 * time.Hour).Truncate(24 * time.Hour)

	return s.repo.CreateRecurringTransaction(s.T().Context(), domain.CreateRecurringTransactionRequest{
		UserID:    s.userID,
		AccountID: accountID,
		Name:      "Test Recurring Transaction",
		Type:      domain.LedgerTypeIncome,
		Amount:    amount,
		Note:      "Recurring Note",
		StartDate: startDate,
		RecurType: recurType,
		Frequency: 1,
		// Adding optional fields for specific tests
		DayOfWeek:   nil,
		DayOfMonth:  nil,
		MonthOfYear: nil,
	})
}

// Define response structs matching the expected {"code": 0, "data": ...} structure
type recurringSingleResponse struct {
	Code int                                      `json:"code"`
	Data bookkeeping.RecurringTransactionResponse `json:"data"`
}

type recurringListResponse struct {
	Code int                                        `json:"code"`
	Data []bookkeeping.RecurringTransactionResponse `json:"data"`
}

type reminderListResponse struct {
	Code int                            `json:"code"`
	Data []bookkeeping.ReminderResponse `json:"data"`
}

type reminderSingleResponse struct {
	Code int                          `json:"code"`
	Data bookkeeping.ReminderResponse `json:"data"`
}

type emptyResponse struct {
	Code int `json:"code"`
	Data any `json:"data"`
}

func (s *testRecurringSuite) TestCreateRecurringTransaction() {
	// Ensure the date is in RFC3339 format and includes optional fields if needed
	now := time.Now()
	reqBody := fmt.Appendf(nil, `{
		"account_id": "%s",
		"name": "Monthly Income",
		"type": "income",
		"amount": "1000.00",
		"note": "Monthly salary",
		"start_date": "%s",
		"recur_type": "monthly",
		"frequency": 1,
		"day_of_month": 1
	}`, s.accountID, now.Format(time.RFC3339))

	req := httptest.NewRequest(http.MethodPost, "/recurring", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Use the correct response struct
	var resp recurringSingleResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code) // Verify the code field

	s.Equal("Monthly Income", resp.Data.Name)
	s.Equal("income", resp.Data.Type)
	s.Equal(decimal.NewFromFloat(1000.00).String(), resp.Data.Amount.String())
	s.Equal("monthly", resp.Data.RecurType)
	s.Equal(1, resp.Data.Frequency)
	s.NotNil(resp.Data.DayOfMonth) // DayOfMonth should not be nil if set in request
	s.Equal(1, *resp.Data.DayOfMonth)
}

func (s *testRecurringSuite) TestGetRecurringTransactions() {
	// Create a recurring transaction first
	_, err := s.createSeedRecurringTransaction(s.accountID, domain.RecurrenceTypeMonthly, decimal.NewFromFloat(500.00))
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/recurring", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Use the correct response struct
	var resp recurringListResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code) // Verify the code field

	s.Len(resp.Data, 1) // Check the length of the Data slice
	s.Equal("Test Recurring Transaction", resp.Data[0].Name)
}

func (s *testRecurringSuite) TestGetRecurringTransaction() {
	// Create a recurring transaction first
	transaction, err := s.createSeedRecurringTransaction(s.accountID, domain.RecurrenceTypeWeekly, decimal.NewFromFloat(50.00))
	s.NoError(err)
	s.NotNil(transaction)

	req := httptest.NewRequest(http.MethodGet, "/recurring/"+transaction.ID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Use the correct response struct
	var resp recurringSingleResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code) // Verify the code field

	s.Equal(transaction.ID, resp.Data.ID)
	s.Equal("Test Recurring Transaction", resp.Data.Name)
}

func (s *testRecurringSuite) TestUpdateRecurringTransaction() {
	// Create a recurring transaction first
	transaction, err := s.createSeedRecurringTransaction(s.accountID, domain.RecurrenceTypeDaily, decimal.NewFromFloat(10.00))
	s.NoError(err)
	s.NotNil(transaction)

	reqBody := []byte(`{
		"name": "Updated Daily Expense",
		"amount": "15.00",
		"status": "paused"
	}`)

	req := httptest.NewRequest(http.MethodPut, "/recurring/"+transaction.ID, bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Use the correct response struct
	var resp recurringSingleResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code) // Verify the code field

	s.Equal("Updated Daily Expense", resp.Data.Name)
	s.Equal(decimal.NewFromFloat(15.00).String(), resp.Data.Amount.String())
	s.Equal("paused", resp.Data.Status)
}

func (s *testRecurringSuite) TestDeleteRecurringTransaction() {
	// Create a recurring transaction first
	transaction, err := s.createSeedRecurringTransaction(s.accountID, domain.RecurrenceTypeYearly, decimal.NewFromFloat(1200.00))
	s.NoError(err)
	s.NotNil(transaction)

	req := httptest.NewRequest(http.MethodDelete, "/recurring/"+transaction.ID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Use the correct response struct for an empty data response
	var resp emptyResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code) // Verify the code field
	s.Nil(resp.Data)      // Expect data to be null or absent for empty response

	// Verify the transaction is deleted (or status is updated to cancelled based on implementation)
	// Assuming delete truly removes it for now or GetRecurringTransactionByID returns error for "deleted" items
	_, err = s.repo.GetRecurringTransactionByID(s.T().Context(), transaction.ID)
	s.Error(err) // Expecting an error because it should be deleted or not found
}

func (s *testRecurringSuite) TestGetReminders() {
	// This test requires a recurring transaction that generates a reminder
	// We need to ensure the reminder is created by the service layer or directly in the test setup

	// Create a recurring transaction that is due
	now := time.Now().Truncate(time.Second) // Truncate to avoid potential timestamp differences

	transaction, err := s.repo.CreateRecurringTransaction(s.T().Context(), domain.CreateRecurringTransactionRequest{
		UserID:    s.userID,
		AccountID: s.accountID,
		Name:      "Reminder Transaction",
		Type:      domain.LedgerTypeIncome,
		Amount:    decimal.NewFromFloat(100.00),
		Note:      "Reminder Note",
		StartDate: now.Add(-time.Hour), // Started in the past
		RecurType: domain.RecurrenceTypeDaily,
		Frequency: 1,
	})
	s.NoError(err)
	s.NotNil(transaction)

	// Manually create a reminder that is due (in a real scenario, this would be done by a scheduler)
	// Ensure reminder date is also truncated
	reminderDate := now.Add(-time.Minute).Truncate(time.Second)
	reminder, err := s.repo.CreateReminder(s.T().Context(), transaction.ID, reminderDate) // Reminder is due now
	s.NoError(err)
	s.NotNil(reminder)
	s.False(reminder.IsRead)

	req := httptest.NewRequest(http.MethodGet, "/recurring/reminders", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Use the correct response struct
	var resp reminderListResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code) // Verify the code field

	s.Len(resp.Data, 1) // Check the length of the Data slice
	s.Equal(reminder.ID, resp.Data[0].ID)
	s.Equal(transaction.ID, resp.Data[0].RecurringTransactionID)
	s.False(resp.Data[0].IsRead)
}

func (s *testRecurringSuite) TestMarkReminderAsRead() {
	// This test requires a recurring transaction and a reminder
	now := time.Now().Truncate(time.Second) // Truncate to avoid potential timestamp differences
	transaction, err := s.repo.CreateRecurringTransaction(s.T().Context(), domain.CreateRecurringTransactionRequest{
		UserID:    s.userID,
		AccountID: s.accountID,
		Name:      "Mark Read Transaction",
		Type:      domain.LedgerTypeExpense,
		Amount:    decimal.NewFromFloat(20.00),
		Note:      "Mark Read Note",
		StartDate: now.Add(-time.Hour),
		RecurType: domain.RecurrenceTypeDaily,
		Frequency: 1,
	})
	s.NoError(err)
	s.NotNil(transaction)

	reminderDate := now.Add(-time.Minute).Truncate(time.Second)
	reminder, err := s.repo.CreateReminder(s.T().Context(), transaction.ID, reminderDate)
	s.NoError(err)
	s.NotNil(reminder)
	s.False(reminder.IsRead)

	req := httptest.NewRequest(http.MethodPost, "/recurring/reminders/"+reminder.ID+"/read", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Use the correct response struct
	var resp reminderSingleResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code) // Verify the code field

	s.True(resp.Data.IsRead)
	s.NotNil(resp.Data.ReadAt)

	// Verify in the repository as well
	updatedReminder, err := s.repo.GetReminderByID(s.T().Context(), reminder.ID)
	s.NoError(err)
	s.True(updatedReminder.IsRead)
	s.NotNil(updatedReminder.ReadAt)
}
