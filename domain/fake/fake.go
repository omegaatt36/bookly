package fake

import (
	"sync"

	"github.com/omegaatt36/bookly/domain"
)

var _ domain.AccountRepository = (*Repository)(nil)
var _ domain.LedgerRepository = (*Repository)(nil)
var _ domain.UserRepository = (*Repository)(nil)

// Repository represents a fake repository
type Repository struct {
	mu sync.RWMutex

	users    map[string]*domain.User
	accounts map[string]*domain.Account
	ledgers  map[string]*domain.Ledger
}

// NewRepository creates a new fake repository
func NewRepository() *Repository {
	return &Repository{
		accounts: make(map[string]*domain.Account),
		ledgers:  make(map[string]*domain.Ledger),
	}
}
