package web

import (
	"log/slog"
	"net/http"
)

func (s *Server) registerRoutes() {
	router := http.NewServeMux()

	// page renderer
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			slog.Error("not found", slog.String("path", r.URL.Path))
			http.NotFound(w, r)
			return
		}

		s.pageIndex(w, r)
	})
	router.HandleFunc("GET /page/accounts", authenticatedHandler(s.pageCreateAccount))
	router.HandleFunc("GET /page/accounts/{account_id}", authenticatedHandler(s.getAccount))
	router.HandleFunc("GET /page/accounts/{account_id}/ledgers", authenticatedHandler(s.pageCreateLedger))
	router.HandleFunc("GET /page/ledgers/{ledger_id}", authenticatedHandler(s.pageLedgerDetails))

	// Authentication
	router.HandleFunc("POST /login", s.login)
	router.HandleFunc("POST /logout", s.logout)

	// Accounts
	router.HandleFunc("GET /accounts", authenticatedHandler(s.accountList))
	router.HandleFunc("POST /accounts", authenticatedHandler(s.createAccount))

	// Ledgers
	router.HandleFunc("GET /accounts/{account_id}/ledgers", authenticatedHandler(s.pageLedgersByAccount))

	router.HandleFunc("POST /accounts/{account_id}/ledgers", authenticatedHandler(s.createLedger))
	router.HandleFunc("PATCH /ledgers/{ledger_id}", authenticatedHandler(s.updateLedger))

	s.router = logging(router)
}
