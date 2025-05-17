package bookkeeping

import (
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/service/bookkeeping"
)

// Controller represents a controller
type Controller struct {
	service *bookkeeping.Service
}

// NewControllerRequest represents a request to create a new controller
type NewControllerRequest struct {
	AccountRepository              domain.AccountRepository
	LedgerRepository               domain.LedgerRepository
	RecurringTransactionRepository domain.RecurringTransactionRepository
	ReminderRepository             domain.ReminderRepository
	BankAccountRepository          domain.BankAccountRepository
}

// NewController creates a new controller
func NewController(req NewControllerRequest) *Controller {
	return &Controller{
		service: bookkeeping.NewService(bookkeeping.NewServiceRequest{
			AccountRepo:              req.AccountRepository,
			LedgerRepo:               req.LedgerRepository,
			RecurringTransactionRepo: req.RecurringTransactionRepository,
			ReminderRepo:             req.ReminderRepository,
			BankAccountRepo:          req.BankAccountRepository,
		}),
	}
}
