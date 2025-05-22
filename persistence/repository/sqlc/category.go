package sqlc

import (
	"context"
	// "database/sql" // No longer needed directly for sql.DB

	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
)

// CategoryRepository implements domain.CategoryRepository
type CategoryRepository struct {
	querier sqlcgen.Querier
	db      sqlcgen.DBTX // Changed from *sql.DB to sqlcgen.DBTX
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(db sqlcgen.DBTX, querier sqlcgen.Querier) domain.CategoryRepository { // Changed parameter type
	return &CategoryRepository{
		querier: querier,
		db:      db,
	}
}

// CreateCategory creates a new category
func (r *CategoryRepository) CreateCategory(ctx context.Context, req domain.CreateCategoryRequest) (*domain.Category, error) {
	dbCategory, err := r.querier.CreateCategory(ctx, r.db, sqlcgen.CreateCategoryParams{
		UserID: req.UserID,
		Name:   req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Category{
		ID:        dbCategory.ID,
		CreatedAt: dbCategory.CreatedAt,
		UpdatedAt: dbCategory.UpdatedAt,
		UserID:    dbCategory.UserID,
		Name:      dbCategory.Name,
	}, nil
}

// GetCategoryByID retrieves a category by its ID
func (r *CategoryRepository) GetCategoryByID(ctx context.Context, id int32) (*domain.Category, error) {
	dbCategory, err := r.querier.GetCategoryByID(ctx, r.db, id)
	if err != nil {
		return nil, err
	}
	return &domain.Category{
		ID:        dbCategory.ID,
		CreatedAt: dbCategory.CreatedAt,
		UpdatedAt: dbCategory.UpdatedAt,
		UserID:    dbCategory.UserID,
		Name:      dbCategory.Name,
	}, nil
}

// GetCategoriesByUserID retrieves categories by user ID
func (r *CategoryRepository) GetCategoriesByUserID(ctx context.Context, userID int32) ([]*domain.Category, error) {
	dbCategories, err := r.querier.GetCategoriesByUserID(ctx, r.db, userID)
	if err != nil {
		return nil, err
	}
	categories := make([]*domain.Category, len(dbCategories))
	for i, dbCategory := range dbCategories {
		categories[i] = &domain.Category{
			ID:        dbCategory.ID,
			CreatedAt: dbCategory.CreatedAt,
			UpdatedAt: dbCategory.UpdatedAt,
			UserID:    dbCategory.UserID,
			Name:      dbCategory.Name,
		}
	}
	return categories, nil
}

// UpdateCategory updates a category
func (r *CategoryRepository) UpdateCategory(ctx context.Context, req domain.UpdateCategoryRequest) (*domain.Category, error) {
	dbCategory, err := r.querier.UpdateCategory(ctx, r.db, sqlcgen.UpdateCategoryParams{
		ID:   req.ID,
		Name: req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &domain.Category{
		ID:        dbCategory.ID,
		CreatedAt: dbCategory.CreatedAt,
		UpdatedAt: dbCategory.UpdatedAt,
		UserID:    dbCategory.UserID,
		Name:      dbCategory.Name,
	}, nil
}

// DeleteCategory deletes a category by its ID
func (r *CategoryRepository) DeleteCategory(ctx context.Context, id int32) error {
	return r.querier.DeleteCategory(ctx, r.db, id)
}
