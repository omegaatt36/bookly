package api

import (
	"net/http"

	"github.com/omegaatt36/bookly/app/api/bookkeeping"
	"github.com/omegaatt36/bookly/app/api/user"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/repository"
	"github.com/omegaatt36/bookly/service/auth"
)

func (s *Server) registerRouters() {
	authenticators := make(map[domain.IdentityProvider]domain.Authenticator)
	if s.jwtSalt != nil && s.jwtSecret != nil {
		authenticators[domain.IdentityProviderPassword] =
			auth.NewJWTAuthorizator(*s.jwtSalt, *s.jwtSecret)
	}

	publicRouter := http.NewServeMux()
	internalRouter := http.NewServeMux()
	v1Router := http.NewServeMux()

	db := database.GetDB()
	repo := repository.NewSQLCRepository(db)
	{
		// Create and configure bookkeeping controller
		bookkeepingX := bookkeeping.NewController(bookkeeping.NewControllerRequest{
			AccountRepository:              repo,
			LedgerRepository:               repo,
			RecurringTransactionRepository: repo,
			ReminderRepository:             repo,
			BankAccountRepository:          repo,
		})

		// Register account routes
		v1Router.HandleFunc("POST /accounts", bookkeepingX.CreateAccount())
		v1Router.HandleFunc("GET /accounts", bookkeepingX.GetAllAccounts())
		v1Router.HandleFunc("GET /accounts/{id}", bookkeepingX.GetAccountByID())
		v1Router.HandleFunc("PATCH /accounts/{id}", bookkeepingX.UpdateAccount())
		v1Router.HandleFunc("DELETE /accounts/{id}", bookkeepingX.DeactivateAccountByID())
		v1Router.HandleFunc("GET /users/{user_id}/accounts", bookkeepingX.GetUserAccounts())

		// Register ledger routes
		v1Router.HandleFunc("POST /accounts/{account_id}/ledgers", bookkeepingX.CreateLedger())
		v1Router.HandleFunc("GET /accounts/{account_id}/ledgers", bookkeepingX.GetLedgersByAccount())
		v1Router.HandleFunc("GET /ledgers/{id}", bookkeepingX.GetLedgerByID())
		v1Router.HandleFunc("PATCH /ledgers/{id}", bookkeepingX.UpdateLedger())
		v1Router.HandleFunc("DELETE /ledgers/{id}", bookkeepingX.VoidLedger())
		v1Router.HandleFunc("POST /ledgers/{id}/adjust", bookkeepingX.AdjustLedger())

		// Register recurring transaction routes
		v1Router.HandleFunc("POST /recurring", bookkeepingX.CreateRecurringTransaction())
		v1Router.HandleFunc("GET /recurring", bookkeepingX.GetRecurringTransactions())
		v1Router.HandleFunc("GET /recurring/{id}", bookkeepingX.GetRecurringTransaction())
		v1Router.HandleFunc("PUT /recurring/{id}", bookkeepingX.UpdateRecurringTransaction())
		v1Router.HandleFunc("DELETE /recurring/{id}", bookkeepingX.DeleteRecurringTransaction())
		v1Router.HandleFunc("GET /recurring/reminders", bookkeepingX.GetReminders())
		v1Router.HandleFunc("POST /recurring/reminders/{id}/read", bookkeepingX.MarkReminderAsRead())
		
		// Register bank account routes
		v1Router.HandleFunc("POST /accounts/{account_id}/bank-account", bookkeepingX.CreateBankAccount())
		v1Router.HandleFunc("GET /accounts/{account_id}/bank-account", bookkeepingX.GetBankAccountByAccountID())
		v1Router.HandleFunc("GET /bank-accounts/{id}", bookkeepingX.GetBankAccountByID())
		v1Router.HandleFunc("PATCH /bank-accounts/{id}", bookkeepingX.UpdateBankAccount())
		v1Router.HandleFunc("DELETE /bank-accounts/{id}", bookkeepingX.DeleteBankAccount())
	}
	{
		userOptions := make([]user.Option, 0)
		for identityProvider, authenticator := range authenticators {
			userOptions = append(userOptions, user.WithAuthenticator(identityProvider, authenticator))
		}

		userX := user.NewController(repo, userOptions...)

		v1Router.HandleFunc("GET /users", userX.GetAllUsers())
		v1Router.HandleFunc("GET /users/{id}", userX.GetUserByID())
		v1Router.HandleFunc("PATCH /users/{id}", userX.UpdateUser())
		v1Router.HandleFunc("DELETE /users/{id}", userX.DeactivateUserByID())
		internalRouter.HandleFunc("POST /users", userX.CreateUser())

		internalRouter.HandleFunc("POST /auth/register", userX.RegisterUser())
		publicRouter.HandleFunc("POST /auth/login", userX.LoginUser())
	}

	authMiddlewares := []middleware{}
	if s.jwtSalt != nil && s.jwtSecret != nil {
		jwtAuthenticator := auth.NewJWTAuthorizator(*s.jwtSalt, *s.jwtSecret)
		authMiddlewares = append(authMiddlewares, authenticated(jwtAuthenticator))
	}

	router := http.NewServeMux()
	router.Handle("/v1/", http.StripPrefix("/v1", chainMiddleware(authMiddlewares...)(v1Router)))
	router.Handle("/internal/", http.StripPrefix("/internal", onlyInternal(*s.internalToken)(internalRouter)))
	router.Handle("/public/", http.StripPrefix("/public", publicRouter))

	s.router = chainMiddleware(rateLimiter(10, 100), logging)(router)
}
