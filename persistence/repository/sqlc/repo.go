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
	_ domain.AccountRepository     = (*Repository)(nil)
	_ domain.LedgerRepository      = (*Repository)(nil)
	_ domain.UserRepository        = (*Repository)(nil)
	_ domain.BankAccountRepository = (*Repository)(nil)
	_ domain.CategoryRepository    = (*Repository)(nil)
	_ domain.BudgetRepository      = (*Repository)(nil)
)

// Repository implements repository interfaces using SQLC-generated code
type Repository struct {
	db                 *pgxpool.Pool
	querier            *sqlcgen.Queries
	ctx                context.Context
	CategoryRepository domain.CategoryRepository
	BudgetRepository   domain.BudgetRepository
}

// NewRepository creates a new SQLC repository
func NewRepository(db *pgxpool.Pool) *Repository {
	querier := sqlcgen.New(db)
	return &Repository{
		db:                 db,
		querier:            querier,
		ctx:                context.Background(),
		CategoryRepository: NewCategoryRepository(db, querier),
		BudgetRepository:   NewBudgetRepository(db, querier),
	}
}

// WithTx creates a new repository with transaction context
func (r *Repository) WithTx(tx pgx.Tx) *Repository {
	txQuerier := r.querier.WithTx(tx)
	return &Repository{
		db:                 r.db,
		querier:            txQuerier,
		ctx:                r.ctx,
		CategoryRepository: NewCategoryRepository(tx, txQuerier),
		BudgetRepository:   NewBudgetRepository(tx, txQuerier),
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