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
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/repository"
)

type testLedgerSuite struct {
	suite.Suite

	router *http.ServeMux

	repo     *repository.GORMRepository
	finalize func()
}

func (s *testLedgerSuite) SetupTest() {
	s.finalize = database.TestingInitialize(database.PostgresOpt)
	s.repo = repository.NewGORMRepository(database.GetDB())
	s.router = http.NewServeMux()
	controller := bookkeeping.NewController(s.repo, s.repo)
	s.router.HandleFunc("POST /accounts/{account_id}/ledgers", controller.CreateLedger())
	s.router.HandleFunc("GET /accounts/{account_id}/ledgers", controller.GetLedgersByAccount())
	s.router.HandleFunc("GET /ledgers/{id}", controller.GetLedgerByID())
	s.router.HandleFunc("PATCH /ledgers/{id}", controller.UpdateLedger())
	s.router.HandleFunc("DELETE /ledgers/{id}", controller.VoidLedger())
	s.router.HandleFunc("POST /ledgers/{id}/adjust", controller.AdjustLedger())

	s.NoError(s.repo.AutoMigrate())
}

func (s *testLedgerSuite) TearDownTest() {
	s.finalize()
	s.router = nil
	s.repo = nil
}

func TestLedgerSuite(t *testing.T) {
	suite.Run(t, new(testLedgerSuite))
}

func (s *testLedgerSuite) TestCreateLedger() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	reqBody := []byte(`{
  "date": "2023-05-01T00:00:00Z",
  "type": "income",
  "amount": "100.00",
  "note": "Test Ledger"
 }`)

	req := httptest.NewRequest(http.MethodPost, "/accounts/"+accountID+"/ledgers", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type createLedgerResponse struct {
		Code int `json:"code"`
		Data any `json:"data"`
	}

	var resp createLedgerResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	ledgers, err := s.repo.GetLedgersByAccountID(accountID)
	s.NoError(err)
	s.Len(ledgers, 1)

	ledger := ledgers[0]
	s.Equal(accountID, ledger.AccountID)
	s.Equal(domain.LedgerTypeIncome, ledger.Type)
	s.Equal("Test Ledger", ledger.Note)
	s.Equal(decimal.NewFromFloat(100.00).String(), ledger.Amount.String())
	s.NotNil(ledger.CreatedAt)
	s.NotNil(ledger.UpdatedAt)
	s.Equal("2023-05-01T00:00:00Z", ledger.Date.Format(time.RFC3339))

	account, err := s.repo.GetAccountByID(accountID)
	s.NoError(err)
	s.Equal(decimal.NewFromFloat(100.00).String(), account.Balance.String())
}

func (s *testLedgerSuite) TestGetAllLedgers() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	// Create a test ledger
	s.NoError(s.repo.CreateLedger(domain.CreateLedgerRequest{
		AccountID: accountID,
		Date:      time.Now(),
		Type:      domain.LedgerTypeExpense,
		Amount:    decimal.NewFromFloat(50.00),
		Note:      "Test Expense",
	}))

	req := httptest.NewRequest(http.MethodGet, "/accounts/"+accountID+"/ledgers", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type getAllLedgersResponse struct {
		Code int `json:"code"`
		Data []struct {
			ID        string          `json:"id"`
			AccountID string          `json:"account_id"`
			Type      string          `json:"type"`
			Amount    decimal.Decimal `json:"amount"`
			Note      string          `json:"note"`
		} `json:"data"`
	}

	var resp getAllLedgersResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)
	s.Len(resp.Data, 1)

	s.Equal(accountID, resp.Data[0].AccountID)
	s.Equal(domain.LedgerTypeExpense.String(), resp.Data[0].Type)
	s.Equal(decimal.NewFromFloat(50.00).String(), resp.Data[0].Amount.String())
	s.Equal("Test Expense", resp.Data[0].Note)
}

func (s *testLedgerSuite) TestGetLedgerByID() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	// Create a test ledger
	s.NoError(s.repo.CreateLedger(domain.CreateLedgerRequest{
		AccountID: accountID,
		Date:      time.Now(),
		Type:      domain.LedgerTypeIncome,
		Amount:    decimal.NewFromFloat(75.00),
		Note:      "Test Income",
	}))

	ledgers, _ := s.repo.GetLedgersByAccountID(accountID)
	ledgerID := ledgers[0].ID

	req := httptest.NewRequest(http.MethodGet, "/ledgers/"+ledgerID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type getLedgerByIDResponse struct {
		Code int `json:"code"`
		Data struct {
			ID        string          `json:"id"`
			AccountID string          `json:"account_id"`
			Type      string          `json:"type"`
			Amount    decimal.Decimal `json:"amount"`
			Note      string          `json:"note"`
		} `json:"data"`
	}

	var resp getLedgerByIDResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)
	s.Equal(ledgerID, resp.Data.ID)
	s.Equal(accountID, resp.Data.AccountID)
	s.Equal(domain.LedgerTypeIncome.String(), resp.Data.Type)
	s.Equal(decimal.NewFromFloat(75.00), resp.Data.Amount)
	s.Equal("Test Income", resp.Data.Note)
}

func (s *testLedgerSuite) TestUpdateLedger() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	// Create a test ledger
	s.NoError(s.repo.CreateLedger(domain.CreateLedgerRequest{
		AccountID: accountID,
		Date:      time.Now(),
		Type:      domain.LedgerTypeExpense,
		Amount:    decimal.NewFromFloat(100.00),
		Note:      "Original Expense",
	}))

	ledgers, _ := s.repo.GetLedgersByAccountID(accountID)
	ledgerID := ledgers[0].ID

	reqBody := []byte(`{
  "amount": "120.00",
  "note": "Updated Expense"
 }`)

	req := httptest.NewRequest(http.MethodPatch, "/ledgers/"+ledgerID, bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type updateLedgerResponse struct {
		Code int `json:"code"`
		Data any `json:"data"`
	}

	var resp updateLedgerResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	updatedLedger, err := s.repo.GetLedgerByID(ledgerID)
	s.NoError(err)
	s.Equal(decimal.NewFromFloat(120.00).String(), updatedLedger.Amount.String())
	s.Equal("Updated Expense", updatedLedger.Note)

	account, err := s.repo.GetAccountByID(accountID)
	s.NoError(err)
	s.Equal(decimal.NewFromFloat(120.00).String(), account.Balance.String())
}

func (s *testLedgerSuite) TestVoidLedger() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	// Create a test ledger
	s.NoError(s.repo.CreateLedger(domain.CreateLedgerRequest{
		AccountID: accountID,
		Date:      time.Now(),
		Type:      domain.LedgerTypeIncome,
		Amount:    decimal.NewFromFloat(200.00),
		Note:      "Income to be voided",
	}))

	ledgers, _ := s.repo.GetLedgersByAccountID(accountID)
	ledgerID := ledgers[0].ID

	req := httptest.NewRequest(http.MethodDelete, "/ledgers/"+ledgerID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type voidLedgerResponse struct {
		Code int `json:"code"`
		Data any `json:"data"`
	}

	var resp voidLedgerResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	voidedLedger, err := s.repo.GetLedgerByID(ledgerID)
	s.NoError(err)
	s.True(voidedLedger.IsVoided)
	s.NotNil(voidedLedger.VoidedAt)

	account, err := s.repo.GetAccountByID(accountID)
	s.NoError(err)
	s.Equal(decimal.NewFromFloat(0.00).String(), account.Balance.String())
}

func (s *testLedgerSuite) TestAdjustLedger() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	// Create a test ledger
	s.NoError(s.repo.CreateLedger(domain.CreateLedgerRequest{
		AccountID: accountID,
		Date:      time.Now(),
		Type:      domain.LedgerTypeExpense,
		Amount:    decimal.NewFromFloat(150.00),
		Note:      "Original Expense",
	}))

	ledgers, _ := s.repo.GetLedgersByAccountID(accountID)
	originalLedgerID := ledgers[0].ID

	reqBody := []byte(fmt.Sprintf(`{
  "account_id": "%s",
  "date": "2023-05-02T00:00:00Z",
  "type": "expense",
  "amount": "170.00",
  "note": "Adjusted Expense"
 }`, accountID))

	req := httptest.NewRequest(http.MethodPost, "/ledgers/"+originalLedgerID+"/adjust", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type adjustLedgerResponse struct {
		Code int `json:"code"`
		Data any `json:"data"`
	}

	var resp adjustLedgerResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	// Get all ledgers after adjustment
	updatedLedgers, err := s.repo.GetLedgersByAccountID(accountID)
	s.NoError(err)
	s.Len(updatedLedgers, 2)

	for _, ledger := range updatedLedgers {
		if ledger.ID == originalLedgerID {
			s.Equal(decimal.NewFromFloat(150.00).String(), ledger.Amount.String())
			s.Equal("Original Expense", ledger.Note)
			s.False(ledger.IsAdjustment)
			s.Nil(ledger.AdjustedFrom)
		} else {
			s.Equal(decimal.NewFromFloat(170.00).String(), ledger.Amount.String())
			s.True(ledger.IsAdjustment)
			s.NotNil(ledger.AdjustedFrom)
			s.Equal(*ledger.AdjustedFrom, originalLedgerID)
			s.Equal("Adjusted Expense", ledger.Note)
		}
	}

	account, err := s.repo.GetAccountByID(accountID)
	s.NoError(err)
	s.Equal(decimal.NewFromFloat(320.00).String(), account.Balance.String())
}

func (s *testLedgerSuite) createSeedUser() (string, error) {
	userID, err := s.repo.CreateUser(domain.CreateUserRequest{
		Name: seedUser.Name,
	})
	s.NoError(err)

	return userID, nil
}

func (s *testLedgerSuite) createSeedAccount() (accountID string, err error) {
	userID, err := s.createSeedUser()
	s.NoError(err)

	s.NoError(s.repo.CreateAccount(domain.CreateAccountRequest{
		UserID:   userID,
		Name:     seedAccount.Name,
		Currency: seedAccount.Currency,
	}))

	accounts, err := s.repo.GetAllAccounts()
	if err != nil {
		return
	}

	accountID = accounts[0].ID

	return
}
