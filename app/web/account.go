package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/omegaatt36/bookly/app"
)

type account struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Currency string `json:"currency"`
	Balance  string `json:"balance"`
}

func (s *Server) pageCreateAccount(w http.ResponseWriter, r *http.Request) {
	if err := s.templates.ExecuteTemplate(w, "create_account.html", nil); err != nil {
		slog.Error("failed to render new_account.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
func (s *Server) createAccount(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserID   string `json:"user_id"`
		Name     string `json:"name"`
		Currency string `json:"currency"`
	}

	payload.Name = r.FormValue("name")
	payload.Currency = r.FormValue("currency")

	// Get user ID from token
	token, _ := r.Cookie("token")
	userID, err := s.getUserIDFromToken(token.Value)
	if err != nil {
		slog.Error("failed to get user ID from token", slog.String("error", err.Error()))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	payload.UserID = userID

	if err := s.sendRequest(r, "POST", "/v1/accounts", payload, nil); err != nil {
		slog.Error("failed to create account", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to create account", http.StatusInternalServerError)
		return
	}

	s.accountList(w, r)
}

func (s *Server) getAccount(w http.ResponseWriter, r *http.Request) {
	accountID := r.PathValue("account_id")

	var acc account
	err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/accounts/%s", accountID), nil, &acc)
	if err != nil {
		slog.Error("failed to get accounts", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get account", http.StatusInternalServerError)
		return
	}

	result := struct {
		Account account
	}{
		Account: acc,
	}

	if err := s.templates.ExecuteTemplate(w, "account_details.html", result); err != nil {
		slog.Error("failed to render account_list.html", slog.String("error", err.Error()))
	}
}

func (s *Server) accountList(w http.ResponseWriter, r *http.Request) {
	var accounts []account
	err := s.sendRequest(r, "GET", "/v1/accounts", nil, &accounts)
	if err != nil {
		slog.Error("failed to get accounts", slog.String("error", err.Error()), slog.String("request", r.URL.String()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get accounts", http.StatusInternalServerError)
		return
	}

	result := struct {
		Accounts []account
	}{
		Accounts: accounts,
	}

	if err := s.templates.ExecuteTemplate(w, "account_list.html", result); err != nil {
		slog.Error("failed to render account_list.html", slog.String("error", err.Error()))
	}
}
