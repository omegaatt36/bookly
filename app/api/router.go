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
		authenticators[domain.IdentityProviderPassword] = auth.NewJWTAuthorizator(*s.jwtSalt, *s.jwtSecret)
	}

	publicRouter := http.NewServeMux()
	internalRouter := http.NewServeMux()
	v1Router := http.NewServeMux()

	repo := repository.NewGORMRepository(database.GetDB())
	{
		bookkeepingX := bookkeeping.NewController(repo, repo)

		v1Router.HandleFunc("POST /accounts", bookkeepingX.CreateAccount)
		v1Router.HandleFunc("GET /accounts", bookkeepingX.GetAllAccounts)
		v1Router.HandleFunc("GET /accounts/{id}", bookkeepingX.GetAccountByID)
		v1Router.HandleFunc("PATCH /accounts/{id}", bookkeepingX.UpdateAccount)
		v1Router.HandleFunc("DELETE /accounts/{id}", bookkeepingX.DeactivateAccountByID)

		v1Router.HandleFunc("POST /accounts/{account_id}/ledgers", bookkeepingX.CreateLedger)
		v1Router.HandleFunc("GET /accounts/{account_id}/ledgers/", bookkeepingX.GetLedgers)
		v1Router.HandleFunc("GET /ledgers/{id}", bookkeepingX.GetLedgerByID)
		v1Router.HandleFunc("PATCH /ledgers/{id}", bookkeepingX.UpdateLedger)
		v1Router.HandleFunc("DELETE /ledgers/{id}", bookkeepingX.VoidLedger)
		v1Router.HandleFunc("POST /ledgers/{id}/adjust", bookkeepingX.AdjustLedger)
	}
	{
		userOptions := make([]user.Option, 0)
		for identityProvider, authenticator := range authenticators {
			userOptions = append(userOptions, user.WithAuthenticator(identityProvider, authenticator))
		}

		userX := user.NewController(repo, userOptions...)

		v1Router.HandleFunc("GET /users", userX.GetAllUsers)
		v1Router.HandleFunc("GET /users/{id}", userX.GetUserByID)
		v1Router.HandleFunc("PATCH /users/{id}", userX.UpdateUser)
		v1Router.HandleFunc("DELETE /users/{id}", userX.DeactivateUserByID)
		internalRouter.HandleFunc("POST /users", userX.CreateUser)

		authRouter := http.NewServeMux()
		authRouter.HandleFunc("POST /login", userX.LoginUser)
		publicRouter.Handle("/auth/", http.StripPrefix("/auth", authRouter))
		internalRouter.HandleFunc("POST /auth/register", userX.RegisterUser)
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
