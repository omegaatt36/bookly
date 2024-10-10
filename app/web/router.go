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
	router.HandleFunc("GET /page/accounts/create", authenticatedHandler(s.pageCreateAccount))
	router.HandleFunc("GET /page/accounts", authenticatedHandler(s.pageAccountList))
	router.HandleFunc("GET /page/accounts/{account_id}", authenticatedHandler(s.pageAccount))
	router.HandleFunc("GET /page/accounts/{account_id}/ledgers/create", authenticatedHandler(s.pageCreateLedger))
	router.HandleFunc("GET /page/accounts/{account_id}/ledgers", authenticatedHandler(s.pageLedgersByAccount))
	router.HandleFunc("GET /page/ledgers/{ledger_id}/details", authenticatedHandler(s.pageLedgerDetails))
	router.HandleFunc("GET /page/ledgers/{ledger_id}", authenticatedHandler(s.pageLedger))

	// Authentication
	router.HandleFunc("POST /login", s.login)
	router.HandleFunc("POST /logout", s.logout)

	// Accounts
	router.HandleFunc("POST /accounts", authenticatedHandler(s.createAccount))

	// Ledgers
	router.HandleFunc("POST /accounts/{account_id}/ledgers", authenticatedHandler(s.createLedger))
	router.HandleFunc("PATCH /ledgers/{ledger_id}", authenticatedHandler(s.updateLedger))
	router.HandleFunc("DELETE /ledgers/{ledger_id}", authenticatedHandler(s.voidLedger))

	s.router = logging(router)
}
