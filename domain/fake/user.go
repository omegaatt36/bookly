package fake

import (
	"fmt"
	"time"

	"github.com/omegaatt36/bookly/domain"
)

// CreateUser creates a new user with the given request data
func (r *Repository) CreateUser(req domain.CreateUserRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := fmt.Sprintf("USER-%d", len(r.users)+1)
	now := time.Now()
	user := &domain.User{
		ID:        id,
		CreatedAt: now,
		UpdatedAt: now,
		Disabled:  false,
		Name:      req.Name,
		Nickname:  req.Nickname,
	}

	r.users[id] = user
	return nil
}

// GetAllUsers retrieves all users from the repository
func (r *Repository) GetAllUsers() ([]*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]*domain.User, 0, len(r.users))
	for _, user := range r.users {
		users = append(users, user)
	}

	return users, nil
}

// GetUserByID retrieves a user by their ID
func (r *Repository) GetUserByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", id)
	}

	return user, nil
}

// UpdateUser updates an existing user with the given request data
func (r *Repository) UpdateUser(req domain.UpdateUserRequest) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[req.ID]
	if !exists {
		return fmt.Errorf("user not found: %s", req.ID)
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.Disabled != nil {
		user.Disabled = *req.Disabled
	}
	user.UpdatedAt = time.Now()

	return nil
}

// DeactivateUserByID deactivates a user by setting their disabled status to true
func (r *Repository) DeactivateUserByID(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return fmt.Errorf("user not found: %s", id)
	}

	user.Disabled = true
	user.UpdatedAt = time.Now()

	return nil
}
