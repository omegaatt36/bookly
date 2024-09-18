package fake

import (
	"fmt"
	"time"

	"github.com/omegaatt36/bookly/domain"
)

// CreateLedger creates a new ledger entry based on the provided request
func (r *Repository) CreateLedger(req domain.CreateLedgerRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("LED-%d", len(r.ledgers)+1)
	now := time.Now()
	ledger := &domain.Ledger{
		ID:           id,
		CreatedAt:    now,
		UpdatedAt:    now,
		AccountID:    req.AccountID,
		Date:         req.Date,
		Type:         req.Type,
		Amount:       req.Amount,
		Note:         req.Note,
		IsAdjustment: false,
		AdjustedFrom: nil,
		IsVoided:     false,
		VoidedAt:     nil,
	}

	r.accounts[req.AccountID].Balance = r.accounts[req.AccountID].Balance.Add(req.Amount)

	r.ledgers[id] = ledger
	return nil
}

// GetLedgerByID retrieves a ledger entry by its ID
func (r *Repository) GetLedgerByID(id string) (*domain.Ledger, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ledger, exists := r.ledgers[id]
	if !exists {
		return nil, fmt.Errorf("ledger not found: %s", id)
	}

	return ledger, nil
}

// GetLedgersByAccountID retrieves all ledger entries for a given account ID
func (r *Repository) GetLedgersByAccountID(accountID string) ([]*domain.Ledger, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ledgers := make([]*domain.Ledger, 0, len(r.ledgers))
	for _, ledger := range r.ledgers {
		if ledger.AccountID == accountID {
			ledgers = append(ledgers, ledger)
		}
	}

	return ledgers, nil
}

// UpdateLedger updates an existing ledger entry based on the provided request
func (r *Repository) UpdateLedger(req domain.UpdateLedgerRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ledger, exists := r.ledgers[req.ID]
	if !exists {
		return fmt.Errorf("ledger not found: %s", req.ID)
	}

	if req.Date != nil {
		ledger.Date = *req.Date
	}
	if req.Type != nil {
		ledger.Type = *req.Type
	}
	if req.Amount != nil {
		ledger.Amount = *req.Amount
		r.accounts[ledger.AccountID].Balance = r.accounts[ledger.AccountID].Balance.Sub(ledger.Amount)
	}
	if req.Note != nil {
		ledger.Note = *req.Note
	}

	ledger.UpdatedAt = time.Now()

	return nil
}

// VoidLedger marks a ledger entry as voided
func (r *Repository) VoidLedger(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ledger, exists := r.ledgers[id]
	if !exists {
		return fmt.Errorf("ledger not found: %s", id)
	}

	ledger.IsVoided = true
	ledger.VoidedAt = func(t time.Time) *time.Time { return &t }(time.Now())
	ledger.UpdatedAt = time.Now()

	r.accounts[ledger.AccountID].Balance = r.accounts[ledger.AccountID].Balance.Sub(ledger.Amount)

	return nil
}

// AdjustLedger creates a new adjusted ledger entry based on an existing one
func (r *Repository) AdjustLedger(originalID string, adjustment domain.CreateLedgerRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	originalLedger, exists := r.ledgers[originalID]
	if !exists {
		return fmt.Errorf("ledger not found: %s", originalID)
	}

	id := fmt.Sprintf("LED-%d", len(r.ledgers)+1)
	now := time.Now()
	adjustedLedger := &domain.Ledger{
		ID:           id,
		CreatedAt:    now,
		UpdatedAt:    now,
		AccountID:    originalLedger.AccountID,
		Date:         adjustment.Date,
		Type:         adjustment.Type,
		Amount:       adjustment.Amount,
		Note:         adjustment.Note,
		IsAdjustment: true,
		AdjustedFrom: &originalID,
		IsVoided:     false,
		VoidedAt:     nil,
	}

	r.ledgers[id] = adjustedLedger

	r.accounts[adjustment.AccountID].Balance = r.accounts[adjustment.AccountID].Balance.Add(adjustment.Amount)

	return nil
}
