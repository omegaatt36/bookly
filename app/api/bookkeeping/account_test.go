package bookkeeping_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/omegaatt36/bookly/app/api/bookkeeping"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/domain/fake"
)

type testAccountSuite struct {
	suite.Suite

	router *http.ServeMux
	repo   *fake.Repository
}

func (s *testAccountSuite) SetupTest() {
	s.router = http.NewServeMux()
	s.repo = fake.NewRepository()
	controller := bookkeeping.NewController(s.repo, s.repo)
	controller.RegisterAccountRouters(s.router)
}

func (s *testAccountSuite) TearDownTest() {
	s.router = nil
	s.repo = nil
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, new(testAccountSuite))
}

func (s *testAccountSuite) TestCreateAccount() {
	reqBody := []byte(`{"name": "Test Account", "currency": "NTD"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	accounts, err := s.repo.GetAllAccounts()
	s.NoError(err)
	s.Len(accounts, 1)

	account := accounts[0]
	s.Equal("Test Account", account.Name)
	s.Equal("NTD", account.Currency)
	s.Equal(domain.AccountStatusActive, account.Status)
}

func (s *testAccountSuite) TestGetAllAccounts() {
	// Create a test account
	s.repo.CreateAccount(domain.CreateAccountRequest{
		Name:     "Test Account",
		Currency: "NTD",
	})

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var accounts []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Currency string `json:"currency"`
		Status   string `json:"status"`
	}

	s.NoError(json.Unmarshal(w.Body.Bytes(), &accounts))
	s.Len(accounts, 1)
	s.Equal("Test Account", accounts[0].Name)
	s.Equal("NTD", accounts[0].Currency)
	s.Equal(domain.AccountStatusActive.String(), accounts[0].Status)
}

func (s *testAccountSuite) TestGetAccountByID() {
	// Create a test account
	s.repo.CreateAccount(domain.CreateAccountRequest{Name: "Test Account", Currency: "NTD"})
	accounts, _ := s.repo.GetAllAccounts()
	accountID := accounts[0].ID

	req := httptest.NewRequest(http.MethodGet, "/accounts/"+accountID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var account struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Currency string `json:"currency"`
		Status   string `json:"status"`
	}

	s.NoError(json.Unmarshal(w.Body.Bytes(), &account))
	s.Equal(accountID, account.ID)
	s.Equal("Test Account", account.Name)
	s.Equal("NTD", account.Currency)
	s.Equal(domain.AccountStatusActive.String(), account.Status)
}

func (s *testAccountSuite) TestUpdateAccount() {
	// Create a test account
	s.repo.CreateAccount(domain.CreateAccountRequest{Name: "Test Account", Currency: "NTD"})
	accounts, _ := s.repo.GetAllAccounts()
	accountID := accounts[0].ID

	reqBody := []byte(`{"name": "Updated Account"}`)
	req := httptest.NewRequest(http.MethodPatch, "/accounts/"+accountID, bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	updatedAccount, err := s.repo.GetAccountByID(accountID)
	s.NoError(err)
	s.Equal("Updated Account", updatedAccount.Name)
}

func (s *testAccountSuite) TestDeactivateAccountByID() {
	// Create a test account
	s.repo.CreateAccount(domain.CreateAccountRequest{Name: "Test Account", Currency: "NTD"})
	accounts, _ := s.repo.GetAllAccounts()
	accountID := accounts[0].ID

	req := httptest.NewRequest(http.MethodDelete, "/accounts/"+accountID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	deactivatedAccount, err := s.repo.GetAccountByID(accountID)
	s.NoError(err)
	s.Equal(domain.AccountStatusClosed, deactivatedAccount.Status)
}
