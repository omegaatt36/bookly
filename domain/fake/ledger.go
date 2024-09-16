package fake

import (
	"fmt"
	"time"

	"github.com/omegaatt36/bookly/domain"
)

func (r *FakeRepository) CreateLedger(req domain.CreateLedgerRequest) error {
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

	r.ledgers[id] = ledger
	return nil
}

func (r *FakeRepository) GetLedgerByID(id string) (*domain.Ledger, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ledger, exists := r.ledgers[id]
	if !exists {
		return nil, fmt.Errorf("ledger not found: %s", id)
	}

	return ledger, nil
}

func (r *FakeRepository) GetLedgersByAccountID(accountID string) ([]*domain.Ledger, error) {
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

func (r *FakeRepository) UpdateLedger(req domain.UpdateLedgerRequest) error {
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
	}
	if req.Note != nil {
		ledger.Note = *req.Note
	}

	ledger.UpdatedAt = time.Now()

	return nil
}

func (r *FakeRepository) VoidLedger(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	ledger, exists := r.ledgers[id]
	if !exists {
		return fmt.Errorf("ledger not found: %s", id)
	}

	ledger.IsVoided = true
	ledger.VoidedAt = func(t time.Time) *time.Time { return &t }(time.Now())
	ledger.UpdatedAt = time.Now()

	return nil
}

func (r *FakeRepository) AdjustLedger(originalID string, adjustment domain.CreateLedgerRequest) error {
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
	return nil
}
