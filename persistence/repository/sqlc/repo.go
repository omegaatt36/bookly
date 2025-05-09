package sqlc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
)

var (
	_ domain.AccountRepository = (*Repository)(nil)
	_ domain.LedgerRepository  = (*Repository)(nil)
	_ domain.UserRepository    = (*Repository)(nil)
)

// Repository implements repository interfaces using SQLC-generated code
type Repository struct {
	db      *pgxpool.Pool
	querier *sqlcgen.Queries
	ctx     context.Context
}

// NewRepository creates a new SQLC repository
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db:      db,
		querier: sqlcgen.New(db),
		ctx:     context.Background(),
	}
}

// WithTx creates a new repository with transaction context
func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	return &Repository{
		db:      r.db,
		querier: r.querier.WithTx(tx),
		ctx:     r.ctx,
	}
}

// ExecuteTx executes a function within a transaction
func (r *Repository) ExecuteTx(ctx context.Context, fn func(repo *Repository) error) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	txRepo := r.WithTx(tx)
	
	if err := fn(txRepo); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("tx failed: %v, rollback failed: %v", err, rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}