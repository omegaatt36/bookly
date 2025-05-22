package bookkeeping

import (
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/service/bookkeeping"
)

// Controller represents a controller
type Controller struct {
	service         *bookkeeping.Service
	categoryService *bookkeeping.CategoryService // Added
	budgetService   *bookkeeping.BudgetService   // Added
}

// NewControllerRequest represents a request to create a new controller
// This request should also include CategoryRepo and BudgetRepo as per previous subtask's changes to service layer.
type NewControllerRequest struct {
	AccountRepository              domain.AccountRepository
	LedgerRepository               domain.LedgerRepository
	RecurringTransactionRepository domain.RecurringTransactionRepository
	ReminderRepository             domain.ReminderRepository
	BankAccountRepository          domain.BankAccountRepository
	CategoryRepository             domain.CategoryRepository // Added
	BudgetRepository               domain.BudgetRepository   // Added
}

// NewController creates a new controller
func NewController(req NewControllerRequest) *Controller {
	mainService := bookkeeping.NewService(bookkeeping.NewServiceRequest{
		AccountRepo:              req.AccountRepository,
		LedgerRepo:               req.LedgerRepository,
		RecurringTransactionRepo: req.RecurringTransactionRepository,
		ReminderRepo:             req.ReminderRepository,
		BankAccountRepo:          req.BankAccountRepository,
		CategoryRepo:             req.CategoryRepository, // Pass to service
		BudgetRepo:               req.BudgetRepository,   // Pass to service
	})
	return &Controller{
		service:         mainService,
		categoryService: mainService.Category(), // Get from main service
		budgetService:   mainService.Budget(),   // Get from main service
	}
}
