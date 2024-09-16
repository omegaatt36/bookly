package fake

import (
	"sync"

	"github.com/omegaatt36/bookly/domain"
)

var _ domain.AccountRepository = (*Repository)(nil)
var _ domain.LedgerRepository = (*Repository)(nil)

// Repository represents a fake repository
type Repository struct {
	accounts map[string]*domain.Account
	ledgers  map[string]*domain.Ledger
	mu       sync.RWMutex
}

// NewRepository creates a new fake repository
func NewRepository() *Repository {
	return &Repository{
		accounts: make(map[string]*domain.Account),
		ledgers:  make(map[string]*domain.Ledger),
	}
}
