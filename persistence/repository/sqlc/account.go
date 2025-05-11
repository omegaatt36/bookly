package sqlc

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
)

// CreateAccount implements the domain.AccountRepository interface
func (r *Repository) CreateAccount(req domain.CreateAccountRequest) error {
	params := sqlcgen.CreateAccountParams{
		UserID:   req.UserID,
		Name:     req.Name,
		Currency: req.Currency,
		Status:   domain.AccountStatusActive.String(),
		Balance:  decimal.Zero,
	}

	err := r.querier.CreateAccount(r.ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

// GetAccountByID implements the domain.AccountRepository interface
func (r *Repository) GetAccountByID(id string) (*domain.Account, error) {
	account, err := r.querier.GetAccountByID(r.ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	var createdAt, updatedAt time.Time
	if account.CreatedAt.Valid {
		createdAt = account.CreatedAt.Time
	}
	if account.UpdatedAt.Valid {
		updatedAt = account.UpdatedAt.Time
	}

	var deletedAt *time.Time
	if account.DeletedAt.Valid {
		deletedAt = &account.DeletedAt.Time
	}

	return &domain.Account{
		ID:        account.ID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		DeletedAt: deletedAt,
		UserID:    account.UserID,
		Name:      account.Name,
		Status:    domain.AccountStatus(account.Status),
		Currency:  account.Currency,
		Balance:   account.Balance,
	}, nil
}

// UpdateAccount implements the domain.AccountRepository interface
func (r *Repository) UpdateAccount(req domain.UpdateAccountRequest) error {
	params := sqlcgen.UpdateAccountParams{
		ID: req.ID,
	}

	if req.UserID != nil {
		params.UserID = pgtype.UUID{}
		if err := params.UserID.Scan(*req.UserID); err != nil {
			return fmt.Errorf("failed to scan user ID: %w", err)
		}
	}

	if req.Name != nil {
		params.Name = pgtype.Text{
			String: *req.Name,
			Valid:  true,
		}
	}

	if req.Currency != nil {
		params.Currency = pgtype.Text{
			String: *req.Currency,
			Valid:  true,
		}
	}

	if req.Status != nil {
		params.Status = pgtype.Text{
			String: string(*req.Status),
			Valid:  true,
		}
	}

	err := r.querier.UpdateAccount(r.ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	return nil
}

// DeactivateAccountByID implements the domain.AccountRepository interface
// This method now performs a soft delete by setting the deleted_at timestamp.
func (r *Repository) DeactivateAccountByID(id string) error {
	// The DeactivateAccountByID query in SQL now also sets deleted_at.
	// We keep the method name for backward compatibility in the service layer.
	params := sqlcgen.DeactivateAccountByIDParams{
		Status: domain.AccountStatusClosed.String(),
		ID:     id,
	}

	err := r.querier.DeactivateAccountByID(r.ctx, params)
	if err != nil {
		return fmt.Errorf("failed to deactivate account: %w", err)
	}
	return nil
}

// DeleteAccount implements the domain.AccountRepository interface
// This method performs a soft delete by setting the deleted_at timestamp.
func (r *Repository) DeleteAccount(id string) error {
	err := r.querier.DeleteAccount(r.ctx, id)
	if err != nil {
		return fmt.Errorf("failed to soft delete account: %w", err)
	}
	return nil
}

// GetAllAccounts implements the domain.AccountRepository interface
func (r *Repository) GetAllAccounts() ([]*domain.Account, error) {
	accounts, err := r.querier.GetAllAccounts(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all accounts: %w", err)
	}

	domainAccounts := make([]*domain.Account, len(accounts))
	for i, account := range accounts {
		var createdAt, updatedAt time.Time
		if account.CreatedAt.Valid {
			createdAt = account.CreatedAt.Time
		}
		if account.UpdatedAt.Valid {
			updatedAt = account.UpdatedAt.Time
		}

		var deletedAt *time.Time
		if account.DeletedAt.Valid {
			deletedAt = &account.DeletedAt.Time
		}

		domainAccounts[i] = &domain.Account{
			ID:        account.ID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			UserID:    account.UserID,
			Name:      account.Name,
			Status:    domain.AccountStatus(account.Status),
			Currency:  account.Currency,
			Balance:   account.Balance,
			DeletedAt: deletedAt,
		}
	}

	return domainAccounts, nil
}

// GetAccountsByUserID implements the domain.AccountRepository interface
func (r *Repository) GetAccountsByUserID(userID string) ([]*domain.Account, error) {
	accounts, err := r.querier.GetAccountsByUserID(r.ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts by user: %w", err)
	}

	domainAccounts := make([]*domain.Account, len(accounts))
	for i, account := range accounts {
		var createdAt, updatedAt time.Time
		if account.CreatedAt.Valid {
			createdAt = account.CreatedAt.Time
		}
		if account.UpdatedAt.Valid {
			updatedAt = account.UpdatedAt.Time
		}

		domainAccounts[i] = &domain.Account{
			ID:        account.ID,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			UserID:    account.UserID,
			Name:      account.Name,
			Status:    domain.AccountStatus(account.Status),
			Currency:  account.Currency,
			Balance:   account.Balance,
		}
	}

	return domainAccounts, nil
}
