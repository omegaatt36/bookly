package user

import (
	"errors"
	"net/http"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/sdk/datatype"
)

func convertUserFromDomain(u *domain.User) datatype.User {
	return datatype.User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt.Format(datatype.TimeFormat),
		UpdatedAt: u.UpdatedAt.Format(datatype.TimeFormat),
		Name:      u.Name,
		Nickname:  u.Nickname,
		Disabled:  u.Disabled,
	}
}

// CreateUser handles the creation of a new user.
func (x *Controller) CreateUser() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		var req datatype.CreateUserRequest
		engine.Chain(r, w, func(_ *engine.Context, req datatype.CreateUserRequest) (*engine.Empty, error) {
			// Allow user creation without auth - this is typically for signup
			// In a real system with admin roles, we would check if the authenticated user has admin privileges
			if req.Name == "" {
				return nil, app.ParamError(errors.New("name is required"))
			}

			if req.Nickname == "" {
				return nil, app.ParamError(errors.New("nickname is required"))
			}

			return nil, x.service.CreateUser(domain.CreateUserRequest{
				Name:     req.Name,
				Nickname: req.Nickname,
			})
		}).BindJSON(&req).Call(req).ResponseCreated()
	}
}

// GetAllUsers retrieves all users from the system.
func (x *Controller) GetAllUsers() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]datatype.User, error) {
			// Admin validation would go here in a real system
			// For now, check if the user is authenticated
			if ctx.GetUserID() == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}
			users, err := x.service.GetAllUsers()
			if err != nil {
				return nil, err
			}

			jsonUsers := make([]datatype.User, len(users))
			for index, u := range users {
				jsonUsers[index] = convertUserFromDomain(u)
			}

			return jsonUsers, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

func (x *Controller) GetUserSelf() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*datatype.User, error) {
			authenticatedUserID := ctx.GetUserID()
			if authenticatedUserID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// User can only retrieve their own information unless they're an admin
			// In a real system, we would check admin role here
			if id != authenticatedUserID {
				return nil, app.Forbidden(errors.New("access denied: cannot view other user's information"))
			}
			u, err := x.service.GetUserByID(id)
			if err != nil {
				return nil, err
			}

			jsonUser := convertUserFromDomain(u)

			return &jsonUser, nil
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}

// GetUserByID retrieves a user by their ID.
func (x *Controller) GetUserByID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*datatype.User, error) {
			authenticatedUserID := ctx.GetUserID()
			if authenticatedUserID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// User can only retrieve their own information unless they're an admin
			// In a real system, we would check admin role here
			if id != authenticatedUserID {
				return nil, app.Forbidden(errors.New("access denied: cannot view other user's information"))
			}
			u, err := x.service.GetUserByID(id)
			if err != nil {
				return nil, err
			}

			jsonUser := convertUserFromDomain(u)

			return &jsonUser, nil
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}

// UpdateUser handles updating a user's information.
func (x *Controller) UpdateUser() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id       int32
			Name     *string `json:"name"`
			Nickname *string `json:"nickname"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			authenticatedUserID := ctx.GetUserID()
			if authenticatedUserID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// User can only update their own information unless they're an admin
			// In a real system, we would check admin role here
			if req.id != authenticatedUserID {
				return nil, app.Forbidden(errors.New("access denied: cannot update other user's information"))
			}

			return nil, x.service.UpdateUser(domain.UpdateUserRequest{
				ID:       req.id,
				Name:     req.Name,
				Nickname: req.Nickname,
			})
		}).Param("id", &req.id).BindJSON(&req).Call(req).ResponseJSON()
	}
}

// DeactivateUserByID handles the deactivation of a user by their ID.
func (x *Controller) DeactivateUserByID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id int32

		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*engine.Empty, error) {
			authenticatedUserID := ctx.GetUserID()
			if authenticatedUserID == 0 {
				return nil, app.Unauthorized(errors.New("user not authenticated"))
			}

			// User can only deactivate their own account unless they're an admin
			// In a real system, we would check admin role here
			if id != authenticatedUserID {
				return nil, app.Forbidden(errors.New("access denied: cannot deactivate other user's account"))
			}
			return nil, x.service.DeactivateUserByID(id)
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}
