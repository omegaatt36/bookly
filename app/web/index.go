package web

import (
	"log/slog"
	"net/http"
)

func (s *Server) pageIndex(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	isAuthenticated := err == nil && token.Value != ""

	data := struct {
		IsAuthenticated bool
	}{
		IsAuthenticated: isAuthenticated,
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", data); err != nil {
		slog.Error("failed to render layout.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
