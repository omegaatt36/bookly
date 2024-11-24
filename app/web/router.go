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
