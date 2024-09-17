package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/omegaatt36/bookly/domain"
)

var _ domain.AccountRepository = (*GORMRepository)(nil)

// Account represents the database model for an account
type Account struct {
	ID        string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string          `gorm:"type:varchar(255);not null"`
	Status    string          `gorm:"type:varchar(20);not null"`
	Currency  string          `gorm:"type:varchar(3);not null"`
	Balance   decimal.Decimal `gorm:"type:decimal(20,2);not null"`

	Ledgers []Ledger
}

// toDomainAccount converts repository Account to domain.Account
func (a *Account) toDomainAccount() *domain.Account {
	return &domain.Account{
		ID:        a.ID,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdatedAt,
		Name:      a.Name,
		Status:    domain.AccountStatus(a.Status),
		Currency:  a.Currency,
		Balance:   a.Balance,
	}
}

// CreateAccount creates a new account
func (r *GORMRepository) CreateAccount(req domain.CreateAccountRequest) error {
	account := Account{
		Name:     req.Name,
		Currency: req.Currency,
		Status:   domain.AccountStatusActive.String(),
		Balance:  decimal.Zero,
	}
	if err := r.db.Create(&account).Error; err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetAccountByID retrieves an account by its ID
func (r *GORMRepository) GetAccountByID(id string) (*domain.Account, error) {
	var account Account
	if err := r.db.First(&account, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("account not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return account.toDomainAccount(), nil
}

// UpdateAccount updates an existing account
func (r *GORMRepository) UpdateAccount(req domain.UpdateAccountRequest) error {
	var account Account
	if err := r.db.First(&account, "id = ?", req.ID).Error; err != nil {
		return fmt.Errorf("failed to find account: %w", err)
	}

	if req.Name != nil {
		account.Name = *req.Name
	}
	if req.Currency != nil {
		account.Currency = *req.Currency
	}
	if req.Status != nil {
		account.Status = string(*req.Status)
	}
	if err := r.db.Save(&account).Error; err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// DeactivateAccountByID deactivates an account by setting its status to closed
func (r *GORMRepository) DeactivateAccountByID(id string) error {
	result := r.db.Model(&Account{}).Where("id = ?", id).
		Update("status", domain.AccountStatusClosed.String())
	if result.Error != nil {
		return fmt.Errorf("failed to deactivate account: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("account not found: %s", id)
	}

	return nil
}

// GetAllAccounts retrieves all accounts
func (r *GORMRepository) GetAllAccounts() ([]*domain.Account, error) {
	var accounts []Account
	if err := r.db.Find(&accounts).Error; err != nil {
		return nil, fmt.Errorf("failed to get all accounts: %w", err)
	}

	domainAccounts := make([]*domain.Account, len(accounts))
	for i, account := range accounts {
		domainAccounts[i] = account.toDomainAccount()
	}

	return domainAccounts, nil
}
