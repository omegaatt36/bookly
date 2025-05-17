package bookkeeping

import (
	"github.com/omegaatt36/bookly/domain"
)

// Service represents a bookkeeping service
type Service struct {
	accountRepo              domain.AccountRepository
	ledgerRepo               domain.LedgerRepository
	recurringTransactionRepo domain.RecurringTransactionRepository
	reminderRepo             domain.ReminderRepository
	bankAccountRepo          domain.BankAccountRepository
}

// NewServiceRequest represents the request to create a new bookkeeping service
type NewServiceRequest struct {
	AccountRepo              domain.AccountRepository
	LedgerRepo               domain.LedgerRepository
	RecurringTransactionRepo domain.RecurringTransactionRepository
	ReminderRepo             domain.ReminderRepository
	BankAccountRepo          domain.BankAccountRepository
}

// NewService creates a new bookkeeping service
func NewService(req NewServiceRequest) *Service {
	return &Service{
		accountRepo:              req.AccountRepo,
		ledgerRepo:               req.LedgerRepo,
		recurringTransactionRepo: req.RecurringTransactionRepo,
		reminderRepo:             req.ReminderRepo,
		bankAccountRepo:          req.BankAccountRepo,
	}
}
