package web

import "net/http"

func (s *Server) registerRoutes() {
	router := http.NewServeMux()

	// page
	router.HandleFunc("GET /", s.pageIndex)
	router.HandleFunc("GET /accounts/new", authenticatedHandler(s.pageCreateAccount))
	router.HandleFunc("GET /accounts/{account_id}/ledgers/new", authenticatedHandler(s.pageCreateLedger))

	// api
	router.HandleFunc("POST /api/login", s.login)
	router.HandleFunc("POST /api/logout", s.logout)
	router.HandleFunc("GET /api/accounts", authenticatedHandler(s.accountList))
	router.HandleFunc("GET /api/accounts/{account_id}", authenticatedHandler(s.getAccount))
	router.HandleFunc("GET /api/accounts/{account_id}/ledgers", authenticatedHandler(s.getLedgersByAccount))

	router.HandleFunc("POST /api/accounts", authenticatedHandler(s.createAccount))
	router.HandleFunc("POST /api/accounts/{account_id}/ledgers", authenticatedHandler(s.createLedger))

	s.router = logging(router)
}
