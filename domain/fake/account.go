package fake

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/omegaatt36/bookly/domain"
)

// CreateAccount creates a new account with the given request details.
func (r *Repository) CreateAccount(req domain.CreateAccountRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("ACC-%d", len(r.accounts)+1)
	now := time.Now()
	account := &domain.Account{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		UserID:    req.UserID,
		Name:      req.Name,
		Status:    domain.AccountStatusActive,
		Currency:  req.Currency,
		Balance:   decimal.Zero,
	}

	r.accounts[id] = account
	return nil
}

// GetAllAccounts retrieves all accounts from the repository.
func (r *Repository) GetAllAccounts() ([]*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts := make([]*domain.Account, 0, len(r.accounts))
	for _, account := range r.accounts {
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetAccountByID retrieves an account by its ID from the repository.
func (r *Repository) GetAccountByID(id string) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, exists := r.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account not found: %s", id)
	}

	return account, nil
}

// UpdateAccount updates an existing account with the given request details.
func (r *Repository) UpdateAccount(req domain.UpdateAccountRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	account, exists := r.accounts[req.ID]
	if !exists {
		return fmt.Errorf("account not found: %s", req.ID)
	}

	if req.UserID != nil {
		account.UserID = *req.UserID
	}
	if req.Name != nil {
		account.Name = *req.Name
	}
	if req.Currency != nil {
		account.Currency = *req.Currency
	}
	if req.Status != nil {
		account.Status = *req.Status
	}
	account.UpdatedAt = time.Now()

	return nil
}

// DeactivateAccountByID deactivates an account by setting its status to closed.
func (r *Repository) DeactivateAccountByID(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	account, exists := r.accounts[id]
	if !exists {
		return fmt.Errorf("account not found: %s", id)
	}

	account.Status = domain.AccountStatusClosed
	account.UpdatedAt = time.Now()

	return nil
}

// GetAccountsByUserID retrieves all accounts associated with a given user ID.
func (r *Repository) GetAccountsByUserID(userID string) ([]*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts := make([]*domain.Account, 0, len(r.accounts))
	for _, account := range r.accounts {
		if account.UserID == userID {
			accounts = append(accounts, account)
		}
	}

	return accounts, nil
}
