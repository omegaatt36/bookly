package web

import (
	"net/http"
)

func (s *Server) registerRoutes() {
	router := http.NewServeMux()

	// page renderer
	router.HandleFunc("GET /{$}", s.pageIndex)
	router.HandleFunc("GET /", s.page404)

	router.HandleFunc("GET /page/accounts/create", authenticatedHandler(s.pageCreateAccount))
	router.HandleFunc("GET /page/accounts", authenticatedHandler(s.pageAccounts))
	router.HandleFunc("GET /page/accounts/list", authenticatedHandler(s.pageAccountList))
	router.HandleFunc("GET /page/accounts/{account_id}", authenticatedHandler(s.pageAccount))
	router.HandleFunc("GET /page/accounts/{account_id}/ledgers/create", authenticatedHandler(s.pageCreateLedger))
	router.HandleFunc("GET /page/accounts/{account_id}/ledgers", authenticatedHandler(s.pageLedgersByAccount))
	router.HandleFunc("GET /page/accounts/{account_id}/bank-account", authenticatedHandler(s.pageBankAccount))
	router.HandleFunc("GET /page/ledgers/{ledger_id}/details", authenticatedHandler(s.pageLedgerDetails))
	router.HandleFunc("GET /page/ledgers/{ledger_id}", authenticatedHandler(s.pageLedger))
	router.HandleFunc("GET /page/recurring", authenticatedHandler(s.pageRecurringList))
	router.HandleFunc("GET /page/recurring/create", authenticatedHandler(s.pageCreateRecurring))
	router.HandleFunc("GET /page/recurring/{recurring_id}", authenticatedHandler(s.pageRecurringDetails))
	router.HandleFunc("GET /page/reminders", authenticatedHandler(s.pageReminders))

	// Authentication
	router.HandleFunc("POST /login", s.login)
	router.HandleFunc("POST /logout", s.logout)

	// Accounts
	router.HandleFunc("POST /accounts", authenticatedHandler(s.createAccount))
	
	// Bank Accounts
	router.HandleFunc("POST /accounts/{account_id}/bank-account", authenticatedHandler(s.createBankAccount))
	router.HandleFunc("PATCH /bank-accounts/{id}", authenticatedHandler(s.updateBankAccount))
	router.HandleFunc("DELETE /bank-accounts/{id}", authenticatedHandler(s.deleteBankAccount))

	// Ledgers
	router.HandleFunc("POST /accounts/{account_id}/ledgers", authenticatedHandler(s.createLedger))
	router.HandleFunc("PATCH /ledgers/{ledger_id}", authenticatedHandler(s.updateLedger))
	router.HandleFunc("DELETE /ledgers/{ledger_id}", authenticatedHandler(s.voidLedger))

	// Recurring transactions
	router.HandleFunc("POST /recurring", authenticatedHandler(s.createRecurring))
	router.HandleFunc("PUT /recurring/{recurring_id}", authenticatedHandler(s.updateRecurring))
	router.HandleFunc("DELETE /recurring/{recurring_id}", authenticatedHandler(s.deleteRecurring))
	router.HandleFunc("POST /reminders/{reminder_id}/read", authenticatedHandler(s.markReminderAsRead))

	// Categories
	router.HandleFunc("GET /categories", authenticatedHandler(s.pageCategories))
	router.HandleFunc("GET /categories/create", authenticatedHandler(s.pageCreateCategory))
	router.HandleFunc("POST /categories", authenticatedHandler(s.createCategory))
	router.HandleFunc("GET /categories/{id}/edit", authenticatedHandler(s.pageEditCategory))
	router.HandleFunc("POST /categories/{id}", authenticatedHandler(s.updateCategory)) // For updates
	router.HandleFunc("POST /categories/{id}/delete", authenticatedHandler(s.deleteCategory))

	// Budgets
	router.HandleFunc("GET /budgets", authenticatedHandler(s.pageBudgets))
	router.HandleFunc("GET /budgets/create", authenticatedHandler(s.pageCreateBudget))
	router.HandleFunc("POST /budgets", authenticatedHandler(s.createBudget))
	router.HandleFunc("GET /budgets/{id}", authenticatedHandler(s.pageBudgetDetails))
	router.HandleFunc("GET /budgets/{id}/edit", authenticatedHandler(s.pageEditBudget))
	router.HandleFunc("POST /budgets/{id}", authenticatedHandler(s.updateBudget)) // For updates
	router.HandleFunc("POST /budgets/{id}/delete", authenticatedHandler(s.deleteBudget))

	s.router = logging(router)
}
