package bookkeeping

import "github.com/omegaatt36/bookly/domain"

// Service represents a bookkeeping service
type Service struct {
	accountRepo domain.AccountRepository
	ledgerRepo  domain.LedgerRepository
}

// NewService creates a new bookkeeping service
func NewService(accountRepo domain.AccountRepository, ledgerRepo domain.LedgerRepository) *Service {
	return &Service{
		accountRepo: accountRepo,
		ledgerRepo:  ledgerRepo,
	}
}
