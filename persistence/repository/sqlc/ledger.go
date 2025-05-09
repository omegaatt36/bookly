package sqlc

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
)

// CreateLedger implements the domain.LedgerRepository interface
func (r *Repository) CreateLedger(req domain.CreateLedgerRequest) error {
	return r.ExecuteTx(r.ctx, func(repo *Repository) error {
		// Create the ledger entry
		err := repo.querier.CreateLedger(repo.ctx, sqlcgen.CreateLedgerParams{
			AccountID:    req.AccountID,
			Date:         pgtype.Timestamptz{Time: req.Date, Valid: true},
			Type:         string(req.Type),
			Amount:       req.Amount,
			Note:         pgtype.Text{String: req.Note, Valid: true},
			IsAdjustment: false, // isAdjustment
			AdjustedFrom: pgtype.UUID{},
		})
		if err != nil {
			return fmt.Errorf("failed to create ledger: %w", err)
		}

		// Update account balance
		err = repo.querier.IncreaseAccountBalance(repo.ctx, sqlcgen.IncreaseAccountBalanceParams{
			Balance: req.Amount,
			ID:      req.AccountID,
		})
		if err != nil {
			return fmt.Errorf("failed to update account balance: %w", err)
		}

		return nil
	})
}

// GetLedgerByID implements the domain.LedgerRepository interface
func (r *Repository) GetLedgerByID(id string) (*domain.Ledger, error) {
	ledger, err := r.querier.GetLedgerByID(r.ctx, id)
	if err != nil {
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
		IsAdjustment: ledger.IsAdjustment,
		AdjustedFrom: func() *string {
			if ledger.AdjustedFrom.Valid {
				val, err := ledger.AdjustedFrom.Value()
				if err == nil && val != nil {
					if str, ok := val.(string); ok {
						return &str
					}
				}
			}
			return nil
		}(),
		IsVoided: ledger.IsVoided,
		VoidedAt: voidedAt,
	}, nil
}

// GetLedgersByAccountID implements the domain.LedgerRepository interface
func (r *Repository) GetLedgersByAccountID(accountID string) ([]*domain.Ledger, error) {
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
			Currency:     ledger.Currency,
			Amount:       ledger.Amount,
			Note:         ledger.Note.String,
			IsAdjustment: ledger.IsAdjustment,
			AdjustedFrom: func() *string {
				if ledger.AdjustedFrom.Valid {
					val, err := ledger.AdjustedFrom.Value()
					if err == nil && val != nil {
						if str, ok := val.(string); ok {
							return &str
						}
					}
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

		// Update the ledger
		if err := repo.querier.UpdateLedger(repo.ctx, updateParams); err != nil {
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
				if err := repo.querier.IncreaseAccountBalance(repo.ctx, sqlcgen.IncreaseAccountBalanceParams{
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
func (r *Repository) VoidLedger(id string) error {
	return r.ExecuteTx(r.ctx, func(repo *Repository) error {
		// Get the ledger to find the amount and account ID
		ledger, err := repo.querier.GetLedgerByID(repo.ctx, id)
		if err != nil {
			return fmt.Errorf("failed to get ledger for voiding: %w", err)
		}

		// Void the ledger
		if err := repo.querier.VoidLedger(repo.ctx, id); err != nil {
			return fmt.Errorf("failed to void ledger: %w", err)
		}

		// Reverse the amount in the account balance
		reverseAmount := ledger.Amount.Neg()
		if err := repo.querier.IncreaseAccountBalance(repo.ctx, sqlcgen.IncreaseAccountBalanceParams{
			Balance: reverseAmount,
			ID:      ledger.AccountID,
		}); err != nil {
			return fmt.Errorf("failed to update account balance for voided ledger: %w", err)
		}

		return nil
	})
}

// AdjustLedger implements the domain.LedgerRepository interface
func (r *Repository) AdjustLedger(originalID string, adjustment domain.CreateLedgerRequest) error {
	return r.ExecuteTx(r.ctx, func(repo *Repository) error {
		// Parse UUID from originalID
		var adjustedFrom pgtype.UUID
		if err := adjustedFrom.Scan(originalID); err != nil {
			return fmt.Errorf("failed to parse original ID: %w", err)
		}

		// Create a new ledger entry marked as an adjustment
		err := repo.querier.CreateLedger(repo.ctx, sqlcgen.CreateLedgerParams{
			AccountID:    adjustment.AccountID,
			Date:         pgtype.Timestamptz{Time: adjustment.Date, Valid: true},
			Type:         string(adjustment.Type),
			Amount:       adjustment.Amount,
			Note:         pgtype.Text{String: adjustment.Note, Valid: true},
			IsAdjustment: true,
			AdjustedFrom: adjustedFrom,
		})
		if err != nil {
			return fmt.Errorf("failed to create adjustment ledger: %w", err)
		}

		// Update account balance
		if err := repo.querier.IncreaseAccountBalance(repo.ctx, sqlcgen.IncreaseAccountBalanceParams{
			Balance: adjustment.Amount,
			ID:      adjustment.AccountID,
		}); err != nil {
			return fmt.Errorf("failed to update account balance for adjustment: %w", err)
		}

		return nil
	})
}
