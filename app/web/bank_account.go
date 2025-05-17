package web

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/omegaatt36/bookly/app"
)

type bankAccount struct {
	ID            int32  `json:"id"`
	AccountID     int32  `json:"account_id"`
	AccountNumber string `json:"account_number"`
	BankName      string `json:"bank_name"`
	BranchName    string `json:"branch_name,omitempty"`
	SwiftCode     string `json:"swift_code,omitempty"`
}

func (s *Server) pageBankAccount(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.PathValue("account_id"))

	// Get account details
	var acc account
	err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/accounts/%d", accountID), nil, &acc)
	if err != nil {
		slog.Error("failed to get account", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to get account", http.StatusInternalServerError)
		return
	}

	// Get bank account details if exists
	var bankAcc *bankAccount
	if err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/accounts/%d/bank-account", accountID), nil, &bankAcc); err != nil {
		// Check if it's a 404 (bank account doesn't exist), which is okay
		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeNotFound {
			// This is fine, the account doesn't have a bank account yet
			bankAcc = nil
		} else if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		} else {
			slog.Error("failed to get bank account", slog.String("error", err.Error()))
			http.Error(w, "Failed to get bank account details", http.StatusInternalServerError)
			return
		}
	}

	result := struct {
		Account     account
		BankAccount *bankAccount
	}{
		Account:     acc,
		BankAccount: bankAcc,
	}

	if err := s.templates.ExecuteTemplate(w, "bank_account.html", result); err != nil {
		slog.Error("failed to render bank_account.html", slog.String("error", err.Error()))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *Server) createBankAccount(w http.ResponseWriter, r *http.Request) {
	accountID := parseInt32(r.PathValue("account_id"))

	var payload struct {
		AccountID     int32  `json:"account_id"`
		AccountNumber string `json:"account_number"`
		BankName      string `json:"bank_name"`
		BranchName    string `json:"branch_name,omitempty"`
		SwiftCode     string `json:"swift_code,omitempty"`
	}

	payload.AccountID = accountID
	payload.AccountNumber = r.FormValue("account_number")
	payload.BankName = r.FormValue("bank_name")
	payload.BranchName = r.FormValue("branch_name")
	payload.SwiftCode = r.FormValue("swift_code")

	if err := s.sendRequest(r, "POST", fmt.Sprintf("/v1/accounts/%d/bank-account", accountID), payload, nil); err != nil {
		slog.Error("failed to create bank account", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to create bank account", http.StatusInternalServerError)
		return
	}

	// Redirect to account details page
	http.Redirect(w, r, fmt.Sprintf("/page/accounts/%d", accountID), http.StatusSeeOther)
}

func (s *Server) updateBankAccount(w http.ResponseWriter, r *http.Request) {
	bankAccountID := parseInt32(r.PathValue("id"))

	var payload struct {
		AccountNumber string `json:"account_number,omitempty"`
		BankName      string `json:"bank_name,omitempty"`
		BranchName    string `json:"branch_name,omitempty"`
		SwiftCode     string `json:"swift_code,omitempty"`
	}

	payload.AccountNumber = r.FormValue("account_number")
	payload.BankName = r.FormValue("bank_name")
	payload.BranchName = r.FormValue("branch_name")
	payload.SwiftCode = r.FormValue("swift_code")

	if err := s.sendRequest(r, "PATCH", fmt.Sprintf("/v1/bank-accounts/%d", bankAccountID), payload, nil); err != nil {
		slog.Error("failed to update bank account", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to update bank account", http.StatusInternalServerError)
		return
	}

	// Get account ID to redirect back to the correct page
	var bankAcc bankAccount
	if err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/bank-accounts/%d", bankAccountID), nil, &bankAcc); err == nil {
		// Redirect to account details page
		http.Redirect(w, r, fmt.Sprintf("/page/accounts/%d", bankAcc.AccountID), http.StatusSeeOther)
	} else {
		// If can't get the account ID, just redirect to accounts page
		http.Redirect(w, r, "/page/accounts", http.StatusSeeOther)
	}
}

func (s *Server) deleteBankAccount(w http.ResponseWriter, r *http.Request) {
	bankAccountID := parseInt32(r.PathValue("id"))

	// Get bank account to get the account ID before deletion
	var bankAcc bankAccount
	var accountID int32
	if err := s.sendRequest(r, "GET", fmt.Sprintf("/v1/bank-accounts/%d", bankAccountID), nil, &bankAcc); err == nil {
		accountID = bankAcc.AccountID
	}

	if err := s.sendRequest(r, "DELETE", fmt.Sprintf("/v1/bank-accounts/%d", bankAccountID), nil, nil); err != nil {
		slog.Error("failed to delete bank account", slog.String("error", err.Error()))

		var sendRequestError *sendRequestError
		if errors.As(err, &sendRequestError) && sendRequestError.Code == app.CodeUnauthorized {
			s.clearTokenAndRedirect(w)
			return
		}

		http.Error(w, "Failed to delete bank account", http.StatusInternalServerError)
		return
	}

	if accountID != 0 {
		// Redirect to account details page
		http.Redirect(w, r, fmt.Sprintf("/page/accounts/%d", accountID), http.StatusSeeOther)
	} else {
		// If can't get the account ID, just redirect to accounts page
		http.Redirect(w, r, "/page/accounts", http.StatusSeeOther)
	}
}
