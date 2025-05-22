package bookkeeping

import (
	"errors"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
)

type jsonCategory struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	UserID    int32     `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (jc *jsonCategory) fromDomain(category *domain.Category) {
	jc.ID = category.ID
	jc.Name = category.Name
	jc.UserID = category.UserID
	jc.CreatedAt = category.CreatedAt
	jc.UpdatedAt = category.UpdatedAt
}

// CreateCategory handles the creation of a new category
func (x *Controller) CreateCategory() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Name string `json:"name"`
		}
		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*jsonCategory, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}
			if req.Name == "" {
				return nil, app.ParamError(errors.New("name is required"))
			}

			category, err := x.categoryService.CreateCategory(ctx, domain.CreateCategoryRequest{
				UserID: userID,
				Name:   req.Name,
			})
			if err != nil {
				return nil, err
			}
			var res jsonCategory
			res.fromDomain(category)
			return &res, nil
		}).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// ListCategories handles listing all categories for the authenticated user
func (x *Controller) ListCategories() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]jsonCategory, error) {
			userID := ctx.GetUserID()
			if userID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}
			categories, err := x.categoryService.ListCategories(ctx, userID)
			if err != nil {
				return nil, err
			}
			jsonCategories := make([]jsonCategory, len(categories))
			for i, cat := range categories {
				jsonCategories[i].fromDomain(cat)
			}
			return jsonCategories, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetCategoryByID handles retrieving a specific category by its ID
func (x *Controller) GetCategoryByID() func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        var categoryID int32
        engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*jsonCategory, error) {
            userID := ctx.GetUserID()
            if userID == 0 {
                return nil, app.Unauthorized(errors.New("user not authenticated"))
            }

            category, err := x.categoryService.GetCategory(ctx, categoryID, userID)
            if err != nil {
                return nil, err
            }
            var res jsonCategory
            res.fromDomain(category)
            return &res, nil
        }).Param("category_id", &categoryID).Call(&engine.Empty{}).ResponseJSON()
    }
}

// UpdateCategory handles updating a category
func (x *Controller) UpdateCategory() func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        type request struct {
            Name string `json:"name"`
        }
        var req request
        var categoryID int32
        engine.Chain(r, w, func(ctx *engine.Context, req request) (*jsonCategory, error) {
            userID := ctx.GetUserID()
            if userID == 0 {
                return nil, app.Unauthorized(errors.New("user not authenticated"))
            }
            if req.Name == "" {
                return nil, app.ParamError(errors.New("name is required"))
            }

            category, err := x.categoryService.UpdateCategory(ctx, domain.UpdateCategoryRequest{
                ID:   categoryID,
                Name: req.Name,
            }, userID)
            if err != nil {
                return nil, err
            }
            var res jsonCategory
            res.fromDomain(category)
            return &res, nil
        }).Param("category_id", &categoryID).BindJSON(&req).Call(req).ResponseJSON()
    }
}

// DeleteCategory handles deleting a category
func (x *Controller) DeleteCategory() func(w http.ResponseWriter, r *http.Request) {
    return func(w http.ResponseWriter, r *http.Request) {
        var categoryID int32
        engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*engine.Empty, error) {
            userID := ctx.GetUserID()
            if userID == 0 {
                return nil, app.Unauthorized(errors.New("user not authenticated"))
            }
            err := x.categoryService.DeleteCategory(ctx, categoryID, userID)
            return nil, err
        }).Param("category_id", &categoryID).Call(&engine.Empty{}).ResponseJSON()
    }
}
