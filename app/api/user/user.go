package user

import (
	"errors"
	"net/http"
	"time"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
)

type jsonUser struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Name      string `json:"name"`
	Nickname  string `json:"nickname"`
	Disabled  bool   `json:"disabled"`
}

func (r *jsonUser) fromDomain(u *domain.User) {
	r.ID = u.ID
	r.CreatedAt = u.CreatedAt.Format(time.RFC3339)
	r.UpdatedAt = u.UpdatedAt.Format(time.RFC3339)
	r.Name = u.Name
	r.Nickname = u.Nickname
	r.Disabled = u.Disabled
}

// CreateUser handles the creation of a new user.
func (x *Controller) CreateUser() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Name     string `json:"name"`
			Nickname string `json:"nickname"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
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
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) ([]jsonUser, error) {
			users, err := x.service.GetAllUsers()
			if err != nil {
				return nil, err
			}

			jsonUsers := make([]jsonUser, len(users))
			for index, u := range users {
				jsonUsers[index].fromDomain(u)
			}

			return jsonUsers, nil
		}).Call(&engine.Empty{}).ResponseJSON()
	}
}

// GetUserByID retrieves a user by their ID.
func (x *Controller) GetUserByID() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var id string
		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*jsonUser, error) {
			u, err := x.service.GetUserByID(id)
			if err != nil {
				return nil, err
			}

			var jsonUser jsonUser
			jsonUser.fromDomain(u)

			return &jsonUser, nil
		}).Param("id", &id).Call(nil).ResponseJSON()
	}
}

// UpdateUser handles updating a user's information.
func (x *Controller) UpdateUser() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			id       string
			Name     *string `json:"name"`
			Nickname *string `json:"nickname"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {

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
		var id string

		engine.Chain(r, w, func(ctx *engine.Context, _ *engine.Empty) (*engine.Empty, error) {
			return nil, x.service.DeactivateUserByID(id)
		}).Param("id", &id).Call(&engine.Empty{}).ResponseJSON()
	}
}
