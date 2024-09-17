package api

import (
	"net/http"

	"github.com/omegaatt36/bookly/app/api/bookkeeping"
	"github.com/omegaatt36/bookly/persistence/database"
	"github.com/omegaatt36/bookly/persistence/repository"
)

// RegisterRouters registers all routes on the provided router.
func RegisterRouters(router *http.ServeMux) {
	repo := repository.NewGORMRepository(database.GetDB())
	bookkeepingX := bookkeeping.NewController(repo, repo)

	bookkeepingX.RegisterAccountRouters(router)
	bookkeepingX.RegisterLedgerRouters(router)
}
