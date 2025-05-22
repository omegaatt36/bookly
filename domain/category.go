package domain

import (
	"context"
	"time"
)

type Category struct {
	ID        int32
	CreatedAt time.Time
	UpdatedAt time.Time
	UserID    int32
	Name      string
}

type CreateCategoryRequest struct {
	UserID int32
	Name   string
}

type UpdateCategoryRequest struct {
	ID   int32
	Name string // Only name is updatable for now
}

type CategoryRepository interface {
	CreateCategory(ctx context.Context, req CreateCategoryRequest) (*Category, error)
	GetCategoryByID(ctx context.Context, id int32) (*Category, error)
	GetCategoriesByUserID(ctx context.Context, userID int32) ([]*Category, error)
	UpdateCategory(ctx context.Context, req UpdateCategoryRequest) (*Category, error)
	DeleteCategory(ctx context.Context, id int32) error
}
