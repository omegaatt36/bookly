package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/repository/sqlc"
)

// SQLCRepository is a wrapper around the SQLC repository implementation
type SQLCRepository struct {
	*sqlc.Repository
}

// NewSQLCRepository creates a new SQLC repository
func NewSQLCRepository(db *pgxpool.Pool) *SQLCRepository {
	return &SQLCRepository{
		Repository: sqlc.NewRepository(db),
	}
}

// Ensure SQLCRepository implements all required interfaces
var (
	_ domain.AccountRepository = (*SQLCRepository)(nil)
	_ domain.LedgerRepository  = (*SQLCRepository)(nil)
	_ domain.UserRepository    = (*SQLCRepository)(nil)
)