package bookkeeping

import (
	"errors"
	"fmt"
	"time"

	"github.com/omegaatt36/bookly/domain"
)

// CreateLedger creates a new ledger based on the provided CreateLedgerRequest.
func (s *Service) CreateLedger(ctx context.Context, req domain.CreateLedgerRequest) (int32, error) {
	account, err := s.accountRepo.GetAccountByID(req.AccountID)
	if err != nil {
		return 0, fmt.Errorf("account not found: %d, %w", req.AccountID, err)
	}

	// Validate CategoryID if provided
	if req.CategoryID != nil && *req.CategoryID != 0 {
		if s.categoryService == nil { // Guard against nil categoryService if not initialized properly
			return 0, errors.New("category service not initialized")
		}
		_, err := s.categoryService.GetCategory(ctx, *req.CategoryID, account.UserID)
		if err != nil {
			return 0, err // Category not found or forbidden
		}
	}

	// The CreateLedger method in the repository layer was updated to take domain.CreateLedgerRequest.
	// The domain.CreateLedgerRequest now includes *int32 for CategoryID.
	// The repository implementation sqlc.CreateLedger handles the conversion to sql.NullInt32.
	// However, the repository's CreateLedger method signature in the interface might still be CreateLedger(domain.CreateLedgerRequest) (int32, error)
	// Let's assume the context needs to be passed down.
	// The previous version of repo.CreateLedger did not take context.
	// Let's check the domain.LedgerRepository interface. It's CreateLedger(CreateLedgerRequest) (int32, error).
	// This needs to be consistent or the call adapted.
	// For now, proceeding with the assumption that the repo method CreateLedger does not require context directly,
	// or that s.ledgerRepo is context-aware.
	// The sqlc repository methods do take context, but the domain.LedgerRepository interface method might not.
	// The `persistence/repository/sqlc/ledger.go` CreateLedger does not take context.
	// This is an inconsistency that might need fixing across layers.
	// Sticking to the current interface:
	return s.ledgerRepo.CreateLedger(req)

}

// GetLedgerByID retrieves a ledger by its ID.
func (s *Service) GetLedgerByID(id int32) (*domain.Ledger, error) {
	return s.ledgerRepo.GetLedgerByID(id)
}

// GetLedgersByAccountID retrieves ledgers by account ID.
func (s *Service) GetLedgersByAccountID(accountID int32) ([]*domain.Ledger, error) {
	return s.ledgerRepo.GetLedgersByAccountID(accountID)
}

// UpdateLedger updates an existing ledger based on the provided UpdateLedgerRequest.
func (s *Service) UpdateLedger(ctx context.Context, req domain.UpdateLedgerRequest) error {
	ledger, err := s.ledgerRepo.GetLedgerByID(req.ID)
	if err != nil {
		return err
	}

	if time.Since(ledger.UpdatedAt) > domain.EditableDuration {
		return errors.New("ledger is too old to be edited")
	}

	// Validate CategoryID if provided
	if req.CategoryID != nil && *req.CategoryID != 0 {
		if s.categoryService == nil { // Guard against nil categoryService
			return errors.New("category service not initialized")
		}
		// We need UserID to check category ownership. Get it from the account associated with the ledger.
		account, err := s.accountRepo.GetAccountByID(ledger.AccountID)
		if err != nil {
			return fmt.Errorf("could not retrieve account for ledger: %w", err)
		}
		_, err = s.categoryService.GetCategory(ctx, *req.CategoryID, account.UserID)
		if err != nil {
			return err // Category not found or forbidden
		}
	}
	// Similar to CreateLedger, UpdateLedger in the repository layer takes domain.UpdateLedgerRequest.
	// The domain.UpdateLedgerRequest now includes *int32 for CategoryID.
	// The repository implementation sqlc.UpdateLedger handles the conversion to sql.NullInt32.
	// The domain.LedgerRepository interface method is UpdateLedger(UpdateLedgerRequest) error.
	// This interface method does not take context.
	return s.ledgerRepo.UpdateLedger(req)
}

// VoidLedger voids a ledger by its ID.
func (s *Service) VoidLedger(id int32) error {
	return s.ledgerRepo.VoidLedger(id)
}

// AdjustLedger adjusts a ledger by its original ID.
func (s *Service) AdjustLedger(originalID int32, adjustment domain.CreateLedgerRequest) error {
	return s.ledgerRepo.AdjustLedger(originalID, adjustment)
}
