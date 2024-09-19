package bookkeeping

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/domain"
)

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

// CreateAccount handles the creation of a new account
func (x *Controller) CreateAccount(w http.ResponseWriter, r *http.Request) {
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

	if err := x.service.CreateAccount(domain.CreateAccountRequest{
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

// GetAllAccounts handles the retrieval of all accounts
func (x *Controller) GetAllAccounts(w http.ResponseWriter, r *http.Request) {
	accounts, err := x.service.GetAllAccounts()
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

// GetAccountByID handles the retrieval of an account by its ID
func (x *Controller) GetAccountByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	account, err := x.service.GetAccountByID(id)
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

// UpdateAccount handles the updating of an account
func (x *Controller) UpdateAccount(w http.ResponseWriter, r *http.Request) {
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

	if err := x.service.UpdateAccount(domain.UpdateAccountRequest{
		ID:     id,
		Name:   req.Name,
		Status: accountStatus,
	}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeactivateAccountByID handles the deactivation of an account by its ID
func (x *Controller) DeactivateAccountByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "parameter 'id' is required", http.StatusBadRequest)
		return
	}

	err := x.service.DeactivateAccountByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// CreateUserAccount handles the creation of a new account for a specific user
func (x *Controller) CreateUserAccount(w http.ResponseWriter, r *http.Request) {
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

	if err := x.service.CreateAccount(domain.CreateAccountRequest{
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

// GetUserAccounts handles the retrieval of all accounts for a specific user
func (x *Controller) GetUserAccounts(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("userID")
	if userID == "" {
		http.Error(w, "userID is required", http.StatusBadRequest)
		return
	}

	accounts, err := x.service.GetAccountsByUserID(userID)
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
