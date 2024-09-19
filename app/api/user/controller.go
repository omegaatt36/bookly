package user

import (
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/service/user"
)

// Controller represents a controller
type Controller struct {
	service *user.Service
}

// NewController creates a new controller
func NewController(userRepo domain.UserRepository, options ...Option) *Controller {
	controller := Controller{
		service: user.NewService(userRepo),
	}

	for _, option := range options {
		option.apply(&controller)
	}

	return &controller
}

// Option defines an option for controller.
type Option interface {
	apply(*Controller)
}

// WithAuthenticatorOption defines an option to register an authenticator for a given identity provider.
type WithAuthenticatorOption struct {
	IdentityProvider domain.IdentityProvider
	Authenticator    domain.Authenticator
}

// WithAuthenticator creates an option to register an authenticator for a given identity provider.
func WithAuthenticator(identityProvider domain.IdentityProvider, authenticator domain.Authenticator) Option {
	return &WithAuthenticatorOption{
		IdentityProvider: identityProvider,
		Authenticator:    authenticator,
	}
}

func (o *WithAuthenticatorOption) apply(c *Controller) {
	c.service.RegisterAuthenticator(o.IdentityProvider, o.Authenticator)
}
