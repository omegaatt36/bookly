package sqlc

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/jackc/pgx/v5"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
)

// CreateLedger implements the domain.LedgerRepository interface
func (r *Repository) CreateLedger(req domain.CreateLedgerRequest) (int32, error) {
	var ledgerID int32
	err := r.ExecuteTx(r.ctx, func(repo *Repository) error {
		// Create the ledger entry
		var err error
		ledger, err := repo.querier.CreateLedger(repo.ctx, sqlcgen.CreateLedgerParams{
			AccountID: req.AccountID,
			Date:      pgtype.Timestamptz{Time: req.Date, Valid: true},
			Type:      string(req.Type),
			Amount:    req.Amount,
			Note:      pgtype.Text{String: req.Note, Valid: true},
			Category:  pgtype.Text{String: req.Category, Valid: req.Category != ""},
		})
		if err != nil {
			return fmt.Errorf("failed to create ledger: %w", err)
		}

		ledgerID = ledger.ID

		// Update account balance
		_, err = repo.querier.IncreaseAccountBalance(repo.ctx, sqlcgen.IncreaseAccountBalanceParams{
			Balance: req.Amount,
			ID:      req.AccountID,
		})
		if err != nil {
			return fmt.Errorf("failed to update account balance: %w", err)
		}

		return nil
	})

	return ledgerID, err
}

// GetLedgerByID implements the domain.LedgerRepository interface
func (r *Repository) GetLedgerByID(id int32) (*domain.Ledger, error) {
	ledger, err := r.querier.GetLedgerByID(r.ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get ledger: %w", err)
	}

	var voidedAt *time.Time
	if ledger.VoidedAt.Valid {
		voidedAt = &ledger.VoidedAt.Time
	}

	return &domain.Ledger{
		ID:           ledger.ID,
		CreatedAt:    ledger.CreatedAt.Time,
		UpdatedAt:    ledger.UpdatedAt.Time,
		AccountID:    ledger.AccountID,
		Date:         ledger.Date.Time,
		Type:         domain.LedgerType(ledger.Type),
		Currency:     ledger.Currency,
		Amount:       ledger.Amount,
		Note:         ledger.Note.String,
		Category:     ledger.Category.String,
		IsAdjustment: ledger.IsAdjustment,
		AdjustedFrom: func() *int32 {
			if ledger.AdjustedFrom.Valid {
				return &ledger.AdjustedFrom.Int32
			}
			return nil
		}(),
		IsVoided: ledger.IsVoided,
		VoidedAt: voidedAt,
	}, nil
}

// GetLedgersByAccountID implements the domain.LedgerRepository interface
func (r *Repository) GetLedgersByAccountID(accountID int32) ([]*domain.Ledger, error) {
	ledgers, err := r.querier.GetLedgersByAccountID(r.ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get ledgers for account: %w", err)
	}

	domainLedgers := make([]*domain.Ledger, len(ledgers))
	for i, ledger := range ledgers {
		var voidedAt *time.Time
		if ledger.VoidedAt.Valid {
			voidedAt = &ledger.VoidedAt.Time
		}

		domainLedgers[i] = &domain.Ledger{
			ID:           ledger.ID,
			CreatedAt:    ledger.CreatedAt.Time,
			UpdatedAt:    ledger.UpdatedAt.Time,
			AccountID:    ledger.AccountID,
			Date:         ledger.Date.Time,
			Type:         domain.LedgerType(ledger.Type),
			Category:     ledger.Category.String,
			Currency:     ledger.Currency,
			Amount:       ledger.Amount,
			Note:         ledger.Note.String,
			IsAdjustment: ledger.IsAdjustment,
			AdjustedFrom: func() *int32 {
				if ledger.AdjustedFrom.Valid {
					return &ledger.AdjustedFrom.Int32
				}
				return nil
			}(),
			IsVoided: ledger.IsVoided,
			VoidedAt: voidedAt,
		}
	}

	return domainLedgers, nil
}

// UpdateLedger implements the domain.LedgerRepository interface
func (r *Repository) UpdateLedger(req domain.UpdateLedgerRequest) error {
	return r.ExecuteTx(r.ctx, func(repo *Repository) error {
		// Get the old amount for balance adjustment
		var oldAmount decimal.Decimal
		if req.Amount != nil {
			amount, err := repo.querier.GetLedgerAmount(repo.ctx, req.ID)
			if err != nil {
				return fmt.Errorf("failed to get ledger amount: %w", err)
			}
			oldAmount = amount
		}

		// Prepare update params
		updateParams := sqlcgen.UpdateLedgerParams{
			ID: req.ID,
		}

		// Set optional fields if provided
		if req.Date != nil {
			updateParams.Date = pgtype.Timestamptz{
				Time:  *req.Date,
				Valid: true,
			}
		}

		if req.Type != nil {
			updateParams.Type = pgtype.Text{
				String: string(*req.Type),
				Valid:  true,
			}
		}

		if req.Amount != nil {
			updateParams.Amount = pgtype.Numeric{Valid: true}
			updateParams.Amount.InfinityModifier = pgtype.Finite
			updateParams.Amount.NaN = false
			updateParams.Amount.Int = req.Amount.Coefficient()
			updateParams.Amount.Exp = req.Amount.Exponent()
		}

		if req.Note != nil {
			updateParams.Note = pgtype.Text{
				String: *req.Note,
				Valid:  true,
			}
		}

		if req.Category != nil {
			updateParams.Category = pgtype.Text{
				String: *req.Category,
				Valid:  true,
			}
		}

		// Update the ledger
		if _, err := repo.querier.UpdateLedger(repo.ctx, updateParams); err != nil {
			return fmt.Errorf("failed to update ledger: %w", err)
		}

		// If amount has changed, update account balance
		if req.Amount != nil {
			// Get the ledger to find the account ID
			ledger, err := repo.querier.GetLedgerByID(repo.ctx, req.ID)
			if err != nil {
				return fmt.Errorf("failed to get ledger account for balance update: %w", err)
			}

			// Calculate balance adjustment
			adjustment := req.Amount.Sub(oldAmount)
			if !adjustment.IsZero() {
				if _, err := repo.querier.IncreaseAccountBalance(repo.ctx, sqlcgen.IncreaseAccountBalanceParams{
					Balance: adjustment,
					ID:      ledger.AccountID,
				}); err != nil {
					return fmt.Errorf("failed to adjust account balance: %w", err)
				}
			}
		}

		return nil
	})
}

// VoidLedger implements the domain.LedgerRepository interface
// This method now also performs a soft delete by setting the deleted_at timestamp.
func (r *Repository) VoidLedger(id int32) error {
	return r.ExecuteTx(r.ctx, func(repo *Repository) error {
		// Get the ledger to find the amount and account ID
		// The GetLedgerByID query already filters out soft-deleted records.
		ledger, err := repo.querier.GetLedgerByID(repo.ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get ledger for voiding: %w", err)
		}

		// Void the ledger (SQL query now also sets deleted_at)
		if _, err := repo.querier.VoidLedger(repo.ctx, id); err != nil {
			return fmt.Errorf("failed to void ledger: %w", err)
		}

		// Reverse the amount in the account balance
		reverseAmount := ledger.Amount.Neg()
		if _, err := repo.querier.IncreaseAccountBalance(repo.ctx, sqlcgen.IncreaseAccountBalanceParams{
			Balance: reverseAmount,
			ID:      ledger.AccountID,
		}); err != nil {
			return fmt.Errorf("failed to update account balance for voided ledger: %w", err)
		}

		return nil
	})
}

// DeleteLedger implements the domain.LedgerRepository interface for soft delete
func (r *Repository) DeleteLedger(id int32) error {
	// We need to get the ledger first to reverse the balance impact before soft deleting.
	return r.ExecuteTx(r.ctx, func(repo *Repository) error {
		// Get the ledger to find the amount and account ID
		// The GetLedgerByID query already filters out soft-deleted records.
		ledger, err := repo.querier.GetLedgerByID(repo.ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get ledger for soft deletion: %w", err)
		}

		// Soft delete the ledger
		if _, err := repo.querier.DeleteLedger(repo.ctx, id); err != nil {
			return fmt.Errorf("failed to soft delete ledger: %w", err)
		}

		// Reverse the amount in the account balance
		reverseAmount := ledger.Amount.Neg()
		if _, err := repo.querier.IncreaseAccountBalance(repo.ctx, sqlcgen.IncreaseAccountBalanceParams{
			Balance: reverseAmount,
			ID:      ledger.AccountID,
		}); err != nil {
			return fmt.Errorf("failed to update account balance for soft deleted ledger: %w", err)
		}

		return nil
	})
}

// AdjustLedger adjusts a ledger by its original ID.
func (r *Repository) AdjustLedger(originalID int32, adjustment domain.CreateLedgerRequest) error {
	return r.ExecuteTx(r.ctx, func(repo *Repository) error {
		// Create a new ledger entry marked as an adjustment
		_, err := repo.querier.AdjustLedger(repo.ctx, sqlcgen.AdjustLedgerParams{
			AccountID:    adjustment.AccountID,
			Date:         pgtype.Timestamptz{Time: adjustment.Date, Valid: true},
			Type:         string(adjustment.Type),
			Amount:       adjustment.Amount,
			Note:         pgtype.Text{String: adjustment.Note, Valid: true},
			Category:     pgtype.Text{String: adjustment.Category, Valid: adjustment.Category != ""},
			AdjustedFrom: pgtype.Int4{Int32: originalID, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to create adjustment ledger: %w", err)
		}

		// Update account balance
		if _, err := repo.querier.IncreaseAccountBalance(repo.ctx, sqlcgen.IncreaseAccountBalanceParams{
			Balance: adjustment.Amount,
			ID:      adjustment.AccountID,
		}); err != nil {
			return fmt.Errorf("failed to update account balance for adjustment: %w", err)
		}

		return nil
	})
}
