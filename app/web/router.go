package web

import (
	"net/http"
)

func (s *Server) registerRoutes() {
	router := http.NewServeMux()

	// Static files
	router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("app/web/static"))))

	// Authentication routes
	router.HandleFunc("GET /auth/login", s.pageLogin)
	router.HandleFunc("GET /auth/register", s.pageRegister)
	router.HandleFunc("POST /auth/login", s.handleLogin)
	router.HandleFunc("POST /auth/register", s.handleRegister)
	router.HandleFunc("POST /auth/logout", s.handleLogout)

	// Protected routes (require authentication)
	protected := http.NewServeMux()

	// Main pages
	protected.HandleFunc("GET /{$}", authenticatedHandler(s.pageHome))
	protected.HandleFunc("GET /accounts", authenticatedHandler(s.pageAccounts))
	protected.HandleFunc("GET /accounts/{id}", authenticatedHandler(s.pageAccountDetail))
	protected.HandleFunc("GET /accounts/{id}/ledgers", authenticatedHandler(s.pageAccountLedgers))
	protected.HandleFunc("GET /recurring", authenticatedHandler(s.pageRecurring))
	protected.HandleFunc("GET /reminders", authenticatedHandler(s.pageReminders))

	// Modal/Form endpoints
	protected.HandleFunc("GET /accounts/new", authenticatedHandler(s.modalAccountForm))
	protected.HandleFunc("GET /accounts/{id}/edit", authenticatedHandler(s.modalAccountEdit))
	protected.HandleFunc("GET /ledgers/new", authenticatedHandler(s.modalLedgerForm))
	protected.HandleFunc("GET /ledgers/{id}/edit", authenticatedHandler(s.modalLedgerEdit))
	protected.HandleFunc("GET /recurring/new", authenticatedHandler(s.modalRecurringForm))
	protected.HandleFunc("GET /recurring/{id}/edit", authenticatedHandler(s.modalRecurringEdit))

	// HTMX endpoints
	protected.HandleFunc("POST /accounts", authenticatedHandler(s.handleCreateAccount))
	protected.HandleFunc("PUT /accounts/{id}", authenticatedHandler(s.handleUpdateAccount))
	protected.HandleFunc("POST /accounts/{id}/activate", authenticatedHandler(s.handleActivateAccount))
	protected.HandleFunc("POST /accounts/{id}/deactivate", authenticatedHandler(s.handleDeactivateAccount))

	protected.HandleFunc("POST /ledgers", authenticatedHandler(s.handleCreateLedger))
	protected.HandleFunc("PUT /ledgers/{id}", authenticatedHandler(s.handleUpdateLedger))
	protected.HandleFunc("POST /ledgers/{id}/void", authenticatedHandler(s.handleVoidLedger))

	protected.HandleFunc("POST /recurring", authenticatedHandler(s.handleCreateRecurring))
	protected.HandleFunc("PUT /recurring/{id}", authenticatedHandler(s.handleUpdateRecurring))
	protected.HandleFunc("POST /recurring/{id}/pause", authenticatedHandler(s.handlePauseRecurring))
	protected.HandleFunc("POST /recurring/{id}/resume", authenticatedHandler(s.handleResumeRecurring))
	protected.HandleFunc("DELETE /recurring/{id}", authenticatedHandler(s.handleDeleteRecurring))

	protected.HandleFunc("POST /reminders/{id}/read", authenticatedHandler(s.handleMarkReminderRead))
	protected.HandleFunc("POST /reminders/{id}/unread", authenticatedHandler(s.handleMarkReminderUnread))
	protected.HandleFunc("POST /reminders/{id}/execute", authenticatedHandler(s.handleExecuteReminder))
	protected.HandleFunc("POST /reminders/{id}/snooze", authenticatedHandler(s.handleSnoozeReminder))
	protected.HandleFunc("DELETE /reminders/{id}", authenticatedHandler(s.handleDeleteReminder))
	protected.HandleFunc("POST /reminders/mark-all-read", authenticatedHandler(s.handleMarkAllRemindersRead))

	// API endpoints for dashboard data
	protected.HandleFunc("GET /api/accounts/summary", authenticatedHandler(s.apiAccountsSummary))
	protected.HandleFunc("GET /api/ledgers/recent", authenticatedHandler(s.apiRecentLedgers))
	protected.HandleFunc("GET /api/statistics/monthly", authenticatedHandler(s.apiMonthlyStatistics))
	protected.HandleFunc("GET /api/reminders/upcoming", authenticatedHandler(s.apiUpcomingReminders))

	// Wrap protected routes with auth middleware
	router.Handle("/", protected)

	s.router = logging(router)
}
