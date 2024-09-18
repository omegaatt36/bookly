package user

import "github.com/omegaatt36/bookly/domain"

// Controller represents a controller
type Controller struct {
	userRepo domain.UserRepository
}

// NewController creates a new controller
func NewController(usreRepo domain.UserRepository) *Controller {
	return &Controller{
		userRepo: usreRepo,
	}
}
