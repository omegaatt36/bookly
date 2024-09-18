package bookkeeping

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/service/bookkeeping"
)

// RegisterAccountRouters registers account-related routes on the provided router.
func (x *Controller) RegisterAccountRouters(router *http.ServeMux) {
	router.HandleFunc("POST /accounts", x.createAccount)
	router.HandleFunc("GET /accounts", x.getAllAccounts)
	router.HandleFunc("GET /accounts/{id}", x.getAccountByID)
	router.HandleFunc("PATCH /accounts/{id}", x.updateAccount)
	router.HandleFunc("DELETE /accounts/{id}", x.deactivateAccountByID)

	router.HandleFunc("GET /users/{userID}/accounts", x.getUserAccounts)
	router.HandleFunc("POST /users/{userID}/accounts", x.createUserAccount)
}

type jsonAccount struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Currency  string `json:"currency"`
	Balance   string `json:"balance"`
}

func (r *jsonAccount) fromDomain(account *domain.Account) {
	r.ID = account.ID
	r.CreatedAt = account.CreatedAt.Format(time.RFC3339)
	r.UpdatedAt = account.UpdatedAt.Format(time.RFC3339)
	r.Name = account.Name
	r.Status = account.Status.String()
	r.Currency = account.Currency
	r.Balance = account.Balance.String()

}

func (x *Controller) createAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		Name     string `json:"name"`
		Currency string `json:"currency"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		http.Error(w, "currency is required", http.StatusBadRequest)
		return
	}

	if err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).CreateAccount(domain.CreateAccountRequest{
		UserID:   req.UserID,
		Name:     req.Name,
		Currency: req.Currency,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (x *Controller) getAllAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).GetAllAccounts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonAccounts := make([]jsonAccount, len(accounts))
	for index, account := range accounts {
		jsonAccounts[index].fromDomain(account)
	}

	bs, err := json.Marshal(jsonAccounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)

}

func (x *Controller) getAccountByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	account, err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).GetAccountByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var jsonAcocunt jsonAccount
	jsonAcocunt.fromDomain(account)

	bs, err := json.Marshal(jsonAcocunt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)
}

func (x *Controller) updateAccount(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Name   *string `json:"name"`
		Status *string `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var accountStatus *domain.AccountStatus
	if req.Status != nil {
		status, err := domain.ParseAccountStatus(*req.Status)
		if err != nil {
			http.Error(w, "invalid account status", http.StatusBadRequest)
			return
		}

		accountStatus = &status
	}

	if err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).UpdateAccount(domain.UpdateAccountRequest{
		ID:     id,
		Name:   req.Name,
		Status: accountStatus,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (x *Controller) deactivateAccountByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).DeactivateAccountByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (x *Controller) createUserAccount(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if userID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Name     string `json:"name"`
		Currency string `json:"currency"`
	}

	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		http.Error(w, "currency is required", http.StatusBadRequest)
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).CreateAccount(domain.CreateAccountRequest{
		UserID:   userID,
		Name:     req.Name,
		Currency: req.Currency,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func (x *Controller) getUserAccounts(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if userID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	accounts, err := bookkeeping.NewService(x.accountRepo, x.ledgerRepo).GetAccountsByUserID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonAccounts := make([]jsonAccount, len(accounts))
	for index, account := range accounts {
		jsonAccounts[index].fromDomain(account)
	}

	bs, err := json.Marshal(jsonAccounts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(bs)
}
