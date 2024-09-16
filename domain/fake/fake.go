package fake

import (
	"sync"

	"github.com/omegaatt36/bookly/domain"
)

var _ domain.AccountRepository = (*FakeRepository)(nil)
var _ domain.LedgerRepository = (*FakeRepository)(nil)

// FakeRepository represents a fake repository
type FakeRepository struct {
	accounts map[string]*domain.Account
	ledgers  map[string]*domain.Ledger
	mu       sync.RWMutex
}

// NewFakeRepository creates a new fake repository
func NewFakeRepository() *FakeRepository {
	return &FakeRepository{
		accounts: make(map[string]*domain.Account),
		ledgers:  make(map[string]*domain.Ledger),
	}
}
