package sqlc

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
)

// CreateBankAccount implements the domain.BankAccountRepository interface
func (r *Repository) CreateBankAccount(req domain.CreateBankAccountRequest) error {
	params := sqlcgen.CreateBankAccountParams{
		AccountID:     req.AccountID,
		AccountNumber: req.AccountNumber,
		BankName:      req.BankName,
		BranchName: pgtype.Text{
			String: req.BranchName,
			Valid:  true,
		},
		SwiftCode: pgtype.Text{
			String: req.SwiftCode,
			Valid:  true,
		},
	}

	_, err := r.querier.CreateBankAccount(r.ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create bank account: %w", err)
	}
	return nil
}

// GetBankAccountByID implements the domain.BankAccountRepository interface
func (r *Repository) GetBankAccountByID(id int32) (*domain.BankAccount, error) {
	bankAccount, err := r.querier.GetBankAccountByID(r.ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get bank account by ID: %w", err)
	}

	return mapToDomainBankAccount(bankAccount)
}

// GetBankAccountByAccountID implements the domain.BankAccountRepository interface
func (r *Repository) GetBankAccountByAccountID(accountID int32) (*domain.BankAccount, error) {
	bankAccount, err := r.querier.GetBankAccountByAccountID(r.ctx, accountID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get bank account by account ID: %w", err)
	}

	return mapToDomainBankAccount(bankAccount)
}

// UpdateBankAccount implements the domain.BankAccountRepository interface
func (r *Repository) UpdateBankAccount(req domain.UpdateBankAccountRequest) error {
	params := sqlcgen.UpdateBankAccountParams{
		ID: req.ID,
	}

	if req.AccountNumber != nil {
		params.AccountNumber = pgtype.Text{
			String: *req.AccountNumber,
			Valid:  true,
		}
	}

	if req.BankName != nil {
		params.BankName = pgtype.Text{
			String: *req.BankName,
			Valid:  true,
		}
	}

	if req.BranchName != nil {
		params.BranchName = pgtype.Text{
			String: *req.BranchName,
			Valid:  true,
		}
	}

	if req.SwiftCode != nil {
		params.SwiftCode = pgtype.Text{
			String: *req.SwiftCode,
			Valid:  true,
		}
	}

	_, err := r.querier.UpdateBankAccount(r.ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update bank account: %w", err)
	}

	return nil
}

// DeleteBankAccount implements the domain.BankAccountRepository interface
func (r *Repository) DeleteBankAccount(id int32) error {
	_, err := r.querier.DeleteBankAccount(r.ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete bank account: %w", err)
	}
	return nil
}

// Helper function to map database model to domain model
func mapToDomainBankAccount(bankAccount sqlcgen.BankAccount) (*domain.BankAccount, error) {
	var createdAt, updatedAt time.Time
	if bankAccount.CreatedAt.Valid {
		createdAt = bankAccount.CreatedAt.Time
	}
	if bankAccount.UpdatedAt.Valid {
		updatedAt = bankAccount.UpdatedAt.Time
	}

	var deletedAt *time.Time
	if bankAccount.DeletedAt.Valid {
		deletedAt = &bankAccount.DeletedAt.Time
	}

	return &domain.BankAccount{
		ID:            bankAccount.ID,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		DeletedAt:     deletedAt,
		AccountID:     bankAccount.AccountID,
		AccountNumber: bankAccount.AccountNumber,
		BankName:      bankAccount.BankName,
		BranchName:    bankAccount.BranchName.String,
		SwiftCode:     bankAccount.SwiftCode.String,
	}, nil
}
