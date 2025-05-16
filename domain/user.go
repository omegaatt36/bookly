package domain

import "time"

// User represents a user
type User struct {
	ID         int32
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Disabled   bool
	Name       string
	Nickname   string
	Identities []Identity
}

// CreateUserRequest defines the request to create a user
type CreateUserRequest struct {
	Name     string
	Nickname string
}

// UpdateUserRequest defines the request to update a user
type UpdateUserRequest struct {
	ID       int32
	Name     *string
	Nickname *string
	Disabled *bool
}

// UserRepository represents a user repository interface
type UserRepository interface {
	CreateUser(CreateUserRequest) (userID int32, err error)
	GetAllUsers() ([]*User, error)
	GetUserByID(int32) (*User, error)
	UpdateUser(UpdateUserRequest) error
	DeactivateUserByID(int32) error
	DeleteUser(id int32) error
	GetUserByIdentity(provider IdentityProvider, identifier string) (*User, *Identity, error)
	AddIdentity(userID int32, provider Identity) error
}

// IdentityProvider represents an identity provider
type IdentityProvider string

const (
	// IdentityProviderPassword is a IdentityProvider of type password.
	IdentityProviderPassword IdentityProvider = "password"
)

// Identity represents an identity
type Identity struct {
	Provider   IdentityProvider
	Identifier string // like email, or google id or telegram id etc
	Credential string // like password hash or token etc
	LastUsedAt time.Time
}
