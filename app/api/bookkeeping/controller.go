package bookkeeping

import "github.com/omegaatt36/bookly/domain"

// Controller represents a controller
type Controller struct {
	accountRepo domain.AccountRepository
	ledgerRepo  domain.LedgerRepository
}

// NewController creates a new controller
func NewController(accountRepo domain.AccountRepository, ledgerRepo domain.LedgerRepository) *Controller {
	return &Controller{
		accountRepo: accountRepo,
		ledgerRepo:  ledgerRepo,
	}
}
