package domain

import "time"

// User represents a user
type User struct {
	ID        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Disabled  bool
	Name      string
	Nickname  string
}

// CreateUserRequest defines the request to create a user
type CreateUserRequest struct {
	Name     string
	Nickname string
}

// UpdateUserRequest defines the request to update a user
type UpdateUserRequest struct {
	ID       string
	Name     *string
	Nickname *string
	Disabled *bool
}

// UserRepository represents a user repository interface
type UserRepository interface {
	CreateUser(CreateUserRequest) error
	GetAllUsers() ([]*User, error)
	GetUserByID(string) (*User, error)
	UpdateUser(UpdateUserRequest) error
	DeactivateUserByID(string) error
}
