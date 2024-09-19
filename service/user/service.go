package user

import (
	"github.com/omegaatt36/bookly/domain"
)

// Service represents a user service
type Service struct {
	userRepo       domain.UserRepository
	mAuthenticator map[domain.IdentityProvider]domain.Authenticator
}

// NewService creates a new user service
func NewService(userRepo domain.UserRepository) *Service {
	return &Service{
		userRepo:       userRepo,
		mAuthenticator: make(map[domain.IdentityProvider]domain.Authenticator),
	}
}

// RegisterAuthenticator registers an authenticator for a given identity provider.
func (s *Service) RegisterAuthenticator(identityProvider domain.IdentityProvider, authenticator domain.Authenticator) {
	s.mAuthenticator[identityProvider] = authenticator
}
