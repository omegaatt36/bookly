package bookkeeping

import (
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/service/bookkeeping"
)

// Controller represents a controller
type Controller struct {
	service *bookkeeping.Service
}

// NewController creates a new controller
func NewController(accountRepo domain.AccountRepository, ledgerRepo domain.LedgerRepository) *Controller {
	return &Controller{
		service: bookkeeping.NewService(accountRepo, ledgerRepo),
	}
}
