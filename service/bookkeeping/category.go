package bookkeeping

import (
	"context"

	"github.com/omegaatt36/bookly/domain"
)

// CategoryService provides category-related operations
type CategoryService struct {
	repo domain.CategoryRepository // Changed to specific repository
}

// NewCategoryService creates a new CategoryService
func NewCategoryService(repo domain.CategoryRepository) *CategoryService { // Changed parameter type
	return &CategoryService{repo: repo}
}

// CreateCategory creates a new category for a user
func (s *CategoryService) CreateCategory(ctx context.Context, req domain.CreateCategoryRequest) (*domain.Category, error) {
	// Basic validation
	if req.Name == "" {
		return nil, domain.ErrCategoryNameRequired
	}
	if req.UserID == 0 {
		return nil, domain.ErrUserIDRequired // Assuming you have this error defined
	}
	return s.repo.CreateCategory(ctx, req) // Direct call
}

// GetCategory retrieves a category by its ID, ensuring it belongs to the user (or is admin)
func (s *CategoryService) GetCategory(ctx context.Context, categoryID int32, userID int32) (*domain.Category, error) {
	category, err := s.repo.GetCategoryByID(ctx, categoryID) // Direct call
	if err != nil {
		return nil, err
	}
	if category.UserID != userID {
		return nil, domain.ErrForbidden // Ensure this error is defined
	}
	return category, nil
}

// ListCategories retrieves all categories for a user
func (s *CategoryService) ListCategories(ctx context.Context, userID int32) ([]*domain.Category, error) {
	if userID == 0 {
		return nil, domain.ErrUserIDRequired
	}
	return s.repo.GetCategoriesByUserID(ctx, userID) // Direct call
}

// UpdateCategory updates an existing category
func (s *CategoryService) UpdateCategory(ctx context.Context, req domain.UpdateCategoryRequest, userID int32) (*domain.Category, error) {
	if req.Name == "" {
		return nil, domain.ErrCategoryNameRequired
	}
	// Verify ownership
	category, err := s.repo.GetCategoryByID(ctx, req.ID) // Direct call
	if err != nil {
		return nil, err
	}
	if category.UserID != userID {
		return nil, domain.ErrForbidden
	}
	return s.repo.UpdateCategory(ctx, req) // Direct call
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, categoryID int32, userID int32) error {
	// Verify ownership
	category, err := s.repo.GetCategoryByID(ctx, categoryID) // Direct call
	if err != nil {
		return nil, err
	}
	if category.UserID != userID {
		return nil, domain.ErrForbidden
	}
	// Consider checking if category is in use by budgets or ledgers before deleting
	return s.repo.DeleteCategory(ctx, categoryID) // Direct call
}
