package bookkeeping

import (
	"errors"
	"fmt"
	"time"

	"github.com/omegaatt36/bookly/domain"
)

// CreateLedger creates a new ledger based on the provided CreateLedgerRequest.
func (s *Service) CreateLedger(req domain.CreateLedgerRequest) (string, error) {
	if _, err := s.accountRepo.GetAccountByID(req.AccountID); err != nil {
		return "", fmt.Errorf("account not found: %s, %w", req.AccountID, err)
	}

	return s.ledgerRepo.CreateLedger(req)

}

// GetLedgerByID retrieves a ledger by its ID.
func (s *Service) GetLedgerByID(id string) (*domain.Ledger, error) {
	return s.ledgerRepo.GetLedgerByID(id)
}

// GetLedgersByAccountID retrieves ledgers by account ID.
func (s *Service) GetLedgersByAccountID(accountID string) ([]*domain.Ledger, error) {
	return s.ledgerRepo.GetLedgersByAccountID(accountID)
}

// UpdateLedger updates an existing ledger based on the provided UpdateLedgerRequest.
func (s *Service) UpdateLedger(req domain.UpdateLedgerRequest) error {
	ledger, err := s.ledgerRepo.GetLedgerByID(req.ID)
	if err != nil {
		return err
	}

	if time.Since(ledger.UpdatedAt) > domain.EditableDuration {
		return errors.New("ledger is too old to be edited")
	}

	return s.ledgerRepo.UpdateLedger(req)
}

// VoidLedger voids a ledger by its ID.
func (s *Service) VoidLedger(id string) error {
	return s.ledgerRepo.VoidLedger(id)
}

// AdjustLedger adjusts a ledger by its original ID.
func (s *Service) AdjustLedger(originalID string, adjustment domain.CreateLedgerRequest) error {
	return s.ledgerRepo.AdjustLedger(originalID, adjustment)
}
