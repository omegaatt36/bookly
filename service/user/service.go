package user

import (
	"github.com/omegaatt36/bookly/domain"
)

// Service represents a user service
type Service struct {
	userRepo domain.UserRepository
}

// NewService creates a new user service
func NewService(userRepo domain.UserRepository) *Service {
	return &Service{
		userRepo: userRepo,
	}
}
