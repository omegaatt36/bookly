package user

import (
	"fmt"

	"github.com/omegaatt36/bookly/domain"
)

// CreateUser creates a new user based on the provided CreateUserRequest.
func (s *Service) CreateUser(req domain.CreateUserRequest) error {
	if _, err := s.userRepo.CreateUser(req); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by its ID.
func (s *Service) GetUserByID(id string) (*domain.User, error) {
	return s.userRepo.GetUserByID(id)
}

// GetAllUsers retrieves all users.
func (s *Service) GetAllUsers() ([]*domain.User, error) {
	return s.userRepo.GetAllUsers()
}

// UpdateUser updates an existing user based on the provided UpdateUserRequest.
func (s *Service) UpdateUser(req domain.UpdateUserRequest) error {
	return s.userRepo.UpdateUser(req)
}

// DeactivateUserByID deactivates an user by setting its status to disabled.
func (s *Service) DeactivateUserByID(id string) error {
	return s.userRepo.DeactivateUserByID(id)
}
