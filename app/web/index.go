package web

import (
        "log/slog"
        "net/http"

        "github.com/a-h/templ"
        "github.com/omegaatt36/bookly/app/web/views/pages"
)

func (s *Server) pageIndex(w http.ResponseWriter, r *http.Request) {
	token, err := r.Cookie("token")
	isAuthenticated := err == nil && token.Value != ""

	if isAuthenticated {
		// 如果用戶已驗證，重定向到 accounts 頁面
		http.Redirect(w, r, "/page/accounts", http.StatusSeeOther)
		return
	}

        templ.Handler{
                Component: pages.Home(isAuthenticated),
        }.ServeHTTP(w, r)
}

// page404 renders the 404 page
func (s *Server) page404(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	if err := s.templates.ExecuteTemplate(w, "404.html", nil); err != nil {
		slog.Error("failed to render 404.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
