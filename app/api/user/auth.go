package user

import (
	"errors"
	"net/http"

	"github.com/omegaatt36/bookly/app"
	"github.com/omegaatt36/bookly/app/api/engine"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/service/user"
)

// RegisterUser registers a new user.
func (x *Controller) RegisterUser() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (*engine.Empty, error) {
			if req.Email == "" {
				return nil, app.ParamError(errors.New("email is required"))
			}
			if req.Password == "" {
				return nil, app.ParamError(errors.New("password is required"))
			}

			return nil, x.service.Register(user.RegisterRequest{
				Name:       req.Email,
				Provider:   domain.IdentityProviderPassword,
				Identifier: req.Email,
				Credential: req.Password,
			})
		}).BindJSON(&req).Call(req).ResponseCreated()
	}
}

// LoginUser logs in a user.
func (x *Controller) LoginUser() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		type request struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		type response struct {
			Token string `json:"token"`
		}

		var req request
		engine.Chain(r, w, func(ctx *engine.Context, req request) (response, error) {
			if req.Email == "" {
				return response{}, app.ParamError(errors.New("email is required"))
			}
			if req.Password == "" {
				return response{}, app.ParamError(errors.New("password is required"))
			}

			token, err := x.service.Login(user.LoginRequest{
				Provider:   domain.IdentityProviderPassword,
				Identifier: req.Email,
				Credential: req.Password,
			})
			return response{Token: token}, err
		}).BindJSON(&req).Call(req).ResponseJSON()
	}
}
