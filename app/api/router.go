package api

import (
	"net/http"

	"github.com/omegaatt36/bookly/app/api/bookkeeping"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/repository"
)

// NewRouter creates a new router.
func NewRouter() http.Handler {
	router := http.NewServeMux()

	RegisterRouters(router)

	v1 := http.NewServeMux()
	v1.Handle("/v1/", http.StripPrefix("/v1", router))
	registerHealthCheck(v1)

	return chain(logging)(v1)
}

// RegisterRouters registers all routes on the provided router.
func RegisterRouters(router *http.ServeMux) {
	repo := repository.NewGORMRepository(database.GetDB())
	bookkeepingX := bookkeeping.NewController(repo, repo)

	bookkeepingX.RegisterAccountRouters(router)
	bookkeepingX.RegisterLedgerRouters(router)
}

func registerHealthCheck(router *http.ServeMux) {
	router.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
