package fake

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"

	"github.com/omegaatt36/bookly/domain"
)

func (r *FakeRepository) CreateAccount(req domain.CreateAccountRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("ACC-%d", len(r.accounts)+1)
	now := time.Now()
	account := &domain.Account{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Name:      req.Name,
		Status:    domain.AccountStatusActive,
		Currency:  req.Currency,
		Balance:   decimal.Zero,
	}

	r.accounts[id] = account
	return nil
}

func (r *FakeRepository) GetAllAccounts() ([]*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	accounts := make([]*domain.Account, 0, len(r.accounts))
	for _, account := range r.accounts {
		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (r *FakeRepository) GetAccountByID(id string) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, exists := r.accounts[id]
	if !exists {
		return nil, fmt.Errorf("account not found: %s", id)
	}

	return account, nil
}

func (r *FakeRepository) UpdateAccount(req domain.UpdateAccountRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	account, exists := r.accounts[req.ID]
	if !exists {
		return fmt.Errorf("account not found: %s", req.ID)
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

func (r *FakeRepository) DeactivateAccountByID(id string) error {
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
