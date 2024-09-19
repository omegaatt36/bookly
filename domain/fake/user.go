package fake

import (
	"fmt"
	"time"

	"github.com/omegaatt36/bookly/domain"
)

// CreateUser creates a new user with the given request data
func (r *Repository) CreateUser(req domain.CreateUserRequest) (string, error) {
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
	return id, nil
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

// AddIdentity adds an identity to a user
func (r *Repository) AddIdentity(userID string, provider domain.Identity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	for _, identity := range user.Identities {
		if identity.Provider == provider.Provider && identity.Identifier == provider.Identifier {
			return fmt.Errorf("identity already exists")
		}
	}

	user.Identities = append(user.Identities, provider)
	user.UpdatedAt = time.Now()

	return nil
}

// GetUserByIdentity retrieves a user by their identity provider and identifier
func (r *Repository) GetUserByIdentity(provider domain.IdentityProvider, identifier string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		for _, identity := range user.Identities {
			if identity.Provider == provider && identity.Identifier == identifier {
				return user, nil
			}
		}
	}

	return nil, fmt.Errorf("user not found with the given identity")
}
