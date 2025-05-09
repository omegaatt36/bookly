package user

import (
	"errors"
	"fmt"

	"github.com/omegaatt36/bookly/domain"
)

// RegisterRequest defines the request to register a new user
type RegisterRequest struct {
	Name       string
	Nickname   string
	Provider   domain.IdentityProvider
	Identifier string
	Credential string
}

// Register registers a new user
func (s *Service) Register(req RegisterRequest) error {
	if s.mAuthenticator == nil {
		return errors.New("authentication provider not initialized")
	}

	if s.mAuthenticator[req.Provider] == nil {
		return fmt.Errorf("authentication provider not found: %s", req.Provider)
	}

	userID, err := s.userRepo.CreateUser(domain.CreateUserRequest{
		Name:     req.Name,
		Nickname: req.Nickname,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	var credentials string
	if req.Provider == domain.IdentityProviderPassword {
		credentials, err = s.mAuthenticator[req.Provider].HashPassword(req.Credential)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}
	}

	authProvider := domain.Identity{
		Provider:   req.Provider,
		Identifier: req.Identifier,
		Credential: credentials,
	}

	if err := s.userRepo.AddIdentity(userID, authProvider); err != nil {
		return fmt.Errorf("failed to add identity: %w", err)
	}

	return nil
}

// LoginRequest defines the request to login a user
type LoginRequest struct {
	Provider   domain.IdentityProvider
	Identifier string
	Credential string
}

// Login authenticates a user and returns a token
func (s *Service) Login(req LoginRequest) (string, error) {
	if s.mAuthenticator == nil {
		return "", errors.New("authentication provider not initialized")
	}

	authenticator, ok := s.mAuthenticator[req.Provider]
	if !ok {
		return "", fmt.Errorf("authentication provider not found: %s", req.Provider)
	}

	user, identity, err := s.userRepo.GetUserByIdentity(req.Provider, req.Identifier)
	if err != nil {
		return "", errors.New("invalid identifier or credentials")
	}

	valid, err := authenticator.VerifyCredential(req.Credential, identity)
	if err != nil {
		return "", fmt.Errorf("credential verification failed: %w", err)
	}
	if !valid {
		return "", errors.New("invalid identifier or credentials")
	}

	token, err := authenticator.GenerateToken(domain.GenerateTokenRequest{UserID: user.ID})
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return token, nil
}
