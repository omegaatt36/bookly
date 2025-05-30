package web

import (
	"net/http"
)

func (s *Server) registerRoutes() {
	router := http.NewServeMux()

	// page renderer
	// router.HandleFunc("GET /{$}", s.pageIndex)

	s.router = logging(router)
}
