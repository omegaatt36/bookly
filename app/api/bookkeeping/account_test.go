package bookkeeping_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/omegaatt36/bookly/app/api/bookkeeping"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/migration"
	"github.com/omegaatt36/bookly/persistence/repository"
)

var seedUser = domain.User{
	Name: "Tester",
}

var seedAccount = domain.Account{
	Name:     "Test Account",
	Currency: "NTD",
}

type testAccountSuite struct {
	suite.Suite

	router *http.ServeMux

	repo     *repository.SQLCRepository
	finalize func()
}

func (s *testAccountSuite) SetupTest() {
	s.finalize = database.TestingInitialize(database.PostgresOpt)
	db := database.GetDB()
	s.repo = repository.NewSQLCRepository(db)
	s.router = http.NewServeMux()
	controller := bookkeeping.NewController(s.repo, s.repo)

	s.router.HandleFunc("POST /accounts", controller.CreateAccount())
	s.router.HandleFunc("GET /accounts", controller.GetAllAccounts())
	s.router.HandleFunc("GET /accounts/{id}", controller.GetAccountByID())
	s.router.HandleFunc("PATCH /accounts/{id}", controller.UpdateAccount())
	s.router.HandleFunc("DELETE /accounts/{id}", controller.DeactivateAccountByID())
	s.router.HandleFunc("GET /users/{user_id}/accounts", controller.GetUserAccounts())
	s.router.HandleFunc("POST /users/{user_id}/accounts", controller.CreateUserAccount())

	s.NoError(migration.NewMigrator(db).Upgrade())
}

func (s *testAccountSuite) TearDownTest() {
	s.finalize()
	s.router = nil
	s.repo = nil
}

func TestAccountSuite(t *testing.T) {
	suite.Run(t, new(testAccountSuite))
}

func (s *testAccountSuite) TestCreateAccount() {
	userID, err := s.createSeedUser()
	s.NoError(err)

	reqBody := []byte(fmt.Sprintf(`{"user_id": "%s", "name": "Test Account", "currency": "NTD"}`, userID))
	req := httptest.NewRequest(http.MethodPost, "/accounts", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type createAccountResponse struct {
		Code int `json:"code"`
		Data any `json:"data"`
	}

	var resp createAccountResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	accounts, err := s.repo.GetAllAccounts()
	s.NoError(err)
	s.Len(accounts, 1)

	account := accounts[0]
	s.Equal("Test Account", account.Name)
	s.Equal("NTD", account.Currency)
	s.Equal(domain.AccountStatusActive, account.Status)
}

func (s *testAccountSuite) TestGetAllAccounts() {
	_, err := s.createSeedAccount()
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type getAllAccountsResponse struct {
		Code int `json:"code"`
		Data []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Currency string `json:"currency"`
			Status   string `json:"status"`
		} `json:"data"`
	}

	var resp getAllAccountsResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)
	s.Len(resp.Data, 1)
	s.Equal("Test Account", resp.Data[0].Name)
	s.Equal("NTD", resp.Data[0].Currency)
	s.Equal(domain.AccountStatusActive.String(), resp.Data[0].Status)
}

func (s *testAccountSuite) TestGetAccountByID() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	req := httptest.NewRequest(http.MethodGet, "/accounts/"+accountID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type getAccountByIDResponse struct {
		Code int `json:"code"`
		Data struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Currency string `json:"currency"`
			Status   string `json:"status"`
		} `json:"data"`
	}

	var resp getAccountByIDResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)
	s.Equal(accountID, resp.Data.ID)
	s.Equal("Test Account", resp.Data.Name)
	s.Equal("NTD", resp.Data.Currency)
	s.Equal(domain.AccountStatusActive.String(), resp.Data.Status)
}

func (s *testAccountSuite) TestUpdateAccount() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	reqBody := []byte(`{"name": "Updated Account"}`)
	req := httptest.NewRequest(http.MethodPatch, "/accounts/"+accountID, bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type updateAccountResponse struct {
		Code int `json:"code"`
		Data any `json:"data"`
	}

	var resp updateAccountResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	updatedAccount, err := s.repo.GetAccountByID(accountID)
	s.NoError(err)
	s.Equal("Updated Account", updatedAccount.Name)
}

func (s *testAccountSuite) TestDeactivateAccountByID() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	req := httptest.NewRequest(http.MethodDelete, "/accounts/"+accountID, nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type deactivateAccountResponse struct {
		Code int `json:"code"`
		Data any `json:"data"`
	}

	var resp deactivateAccountResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)

	deactivatedAccount, err := s.repo.GetAccountByID(accountID)
	s.NoError(err)
	s.Equal(domain.AccountStatusClosed, deactivatedAccount.Status)
}

func (s *testAccountSuite) TestGetAccountsByUserID() {
	accountID, err := s.createSeedAccount()
	s.NoError(err)

	acc, err := s.repo.GetAccountByID(accountID)
	s.NoError(err)
	s.NotNil(acc)

	req := httptest.NewRequest(http.MethodGet, "/users/"+acc.UserID+"/accounts", nil)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	type getAccountsByUserIDResponse struct {
		Code int `json:"code"`
		Data []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Currency string `json:"currency"`
			Status   string `json:"status"`
		} `json:"data"`
	}

	var resp getAccountsByUserIDResponse
	s.NoError(json.NewDecoder(w.Body).Decode(&resp))
	s.Equal(0, resp.Code)
	s.Len(resp.Data, 1)
	s.Equal(seedAccount.Name, resp.Data[0].Name)
	s.Equal(seedAccount.Currency, resp.Data[0].Currency)
	s.Equal(domain.AccountStatusActive.String(), resp.Data[0].Status)
}

func (s *testAccountSuite) createSeedUser() (string, error) {
	userID, err := s.repo.CreateUser(domain.CreateUserRequest{
		Name: seedUser.Name,
	})
	s.NoError(err)

	return userID, nil
}

func (s *testAccountSuite) createSeedAccount() (accountID string, err error) {
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
