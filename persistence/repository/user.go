package repository

import (
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/omegaatt36/bookly/domain"
)

// Ensure GORMRepository implements domain.UserRepository interface
var _ domain.UserRepository = (*GORMRepository)(nil)

// User represents the database model for a user
type User struct {
	ID        string `gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Disabled  bool   `gorm:"not null;default:false"`
	Name      string `gorm:"type:varchar(255);not null"`
	Nickname  string `gorm:"type:varchar(255)"`
}

// toDomainUser converts repository User to domain.User
func (u *User) toDomainUser() *domain.User {
	return &domain.User{
		ID:        u.ID,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		Disabled:  u.Disabled,
		Name:      u.Name,
		Nickname:  u.Nickname,
	}
}

// Identity represents the database model for an identity
type Identity struct {
	ID         int    `gorm:"primary_key"`
	UserID     string `gorm:"type:uuid;not null;uniqueIndex:idx_user_provider_identifier"`
	Provider   string `gorm:"type:varchar(20);not null;uniqueIndex:idx_user_provider_identifier;uniqueIndex:idx_provider_identifier"`
	Identifier string `gorm:"type:varchar(255);not null;uniqueIndex:idx_user_provider_identifier;uniqueIndex:idx_provider_identifier"`
	Credential string `gorm:"type:varchar(255);not null"`
	LastUsedAt time.Time
}

// CreateUser creates a new user
func (r *GORMRepository) CreateUser(req domain.CreateUserRequest) (string, error) {
	user := User{
		Name:     req.Name,
		Nickname: req.Nickname,
	}
	if err := r.db.Create(&user).Error; err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	return user.ID, nil
}

// GetAllUsers retrieves all users from the database
func (r *GORMRepository) GetAllUsers() ([]*domain.User, error) {
	var users []User
	if err := r.db.Order("updated_at").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	domainUsers := make([]*domain.User, len(users))
	for i, user := range users {
		domainUsers[i] = user.toDomainUser()
	}

	return domainUsers, nil
}

// GetUserByID retrieves a user by their ID
func (r *GORMRepository) GetUserByID(id string) (*domain.User, error) {
	var user User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user.toDomainUser(), nil
}

// UpdateUser updates an existing user
func (r *GORMRepository) UpdateUser(req domain.UpdateUserRequest) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.First(&user, "id = ?", req.ID).Error; err != nil {
			return fmt.Errorf("failed to find user: %w", err)
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

		if err := tx.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		return nil
	})
}

// DeactivateUserByID deactivates a user by setting their disabled status to true
func (r *GORMRepository) DeactivateUserByID(id string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&User{}).Where("id = ?", id).Update("disabled", true)
		if result.Error != nil {
			return fmt.Errorf("failed to deactivate user: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			return fmt.Errorf("user not found: %s", id)
		}
		return nil
	})
}

// AddIdentity adds a new identity for a user
func (r *GORMRepository) AddIdentity(userID string, identity domain.Identity) error {
	id := Identity{
		UserID:     userID,
		Provider:   string(identity.Provider),
		Identifier: identity.Identifier,
		Credential: identity.Credential,
	}

	if err := r.db.Create(&id).Error; err != nil {
		return fmt.Errorf("failed to add identity: %w", err)
	}

	return nil
}

// GetUserByIdentity retrieves a user by their identity provider and identifier
func (r *GORMRepository) GetUserByIdentity(provider domain.IdentityProvider, identifier string) (*domain.User, error) {
	var user User
	if err := r.db.Where("id = (?)", r.db.Model(&Identity{}).
		Select("user_id").
		Where("provider = ? AND identifier = ?", provider, identifier)).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with identity not found: %s %s", provider, identifier)
		}
		return nil, fmt.Errorf("failed to get user by identity: %w", err)
	}

	return user.toDomainUser(), nil
}
