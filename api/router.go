package api

import (
	"net/http"

	"github.com/omegaatt36/bookly/api/bookkeeping"
	"github.com/omegaatt36/bookly/domain/fake"
)

// RegisterRouters registers all routes on the provided router.
func RegisterRouters(router *http.ServeMux) {
	repo := fake.NewFakeRepository()
	bookkeepingX := bookkeeping.NewController(repo, repo)

	bookkeepingX.RegisterAccountRouters(router)
	bookkeepingX.RegisterLedgerRouters(router)
}
