package sqlc

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/omegaatt36/bookly/domain"
	"github.com/omegaatt36/bookly/persistence/sqlcgen"
)

func convertToDomainUser(user sqlcgen.User) *domain.User {
	return &domain.User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Disabled:  user.Disabled,
		Name:      user.Name,
		Nickname:  user.Nickname.String,
	}
}

// CreateUser implements the domain.UserRepository interface
func (r *Repository) CreateUser(req domain.CreateUserRequest) (string, error) {
	userID, err := r.querier.CreateUser(r.ctx, sqlcgen.CreateUserParams{
		Name:     req.Name,
		Nickname: pgtype.Text{String: req.Nickname, Valid: req.Nickname != ""},
	})
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}
	return userID, nil
}

// GetAllUsers implements the domain.UserRepository interface
func (r *Repository) GetAllUsers() ([]*domain.User, error) {
	users, err := r.querier.GetAllUsers(r.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get all users: %w", err)
	}

	domainUsers := make([]*domain.User, len(users))
	for i, user := range users {
		domainUsers[i] = convertToDomainUser(user)
	}

	return domainUsers, nil
}

// GetUserByID implements the domain.UserRepository interface
func (r *Repository) GetUserByID(id string) (*domain.User, error) {
	user, err := r.querier.GetUserByID(r.ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &domain.User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
		Disabled:  user.Disabled,
		Name:      user.Name,
		Nickname:  user.Nickname.String,
	}, nil
}

// UpdateUser implements the domain.UserRepository interface
func (r *Repository) UpdateUser(req domain.UpdateUserRequest) error {
	var params sqlcgen.UpdateUserParams
	params.ID = req.ID

	if req.Name != nil {
		params.Name = pgtype.Text{
			String: *req.Name,
			Valid:  true,
		}
	}

	if req.Nickname != nil {
		params.Nickname = pgtype.Text{
			String: *req.Nickname,
			Valid:  true,
		}
	}

	if req.Disabled != nil {
		params.Disabled = pgtype.Bool{
			Bool:  *req.Disabled,
			Valid: true,
		}
	}

	if err := r.querier.UpdateUser(r.ctx, params); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeactivateUserByID implements the domain.UserRepository interface
// This method now performs a soft delete by setting the deleted_at timestamp.
func (r *Repository) DeactivateUserByID(id string) error {
	// The DeactivateUserByID query in SQL now also sets deleted_at.
	// We keep the method name for backward compatibility in the service layer.
	if err := r.querier.DeactivateUserByID(r.ctx, id); err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}
	return nil
}

// AddIdentity implements the domain.UserRepository interface
func (r *Repository) AddIdentity(userID string, provider domain.Identity) error {
	if err := r.querier.AddIdentity(r.ctx, sqlcgen.AddIdentityParams{
		UserID:     userID,
		Provider:   string(provider.Provider),
		Identifier: provider.Identifier,
		Credential: provider.Credential,
	}); err != nil {
		return fmt.Errorf("failed to add identity: %w", err)
	}
	return nil
}

// DeleteUser implements the domain.UserRepository interface for soft delete
func (r *Repository) DeleteUser(id string) error {
	if err := r.querier.DeleteUser(r.ctx, id); err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}
	return nil
}

// GetUserByIdentity implements the domain.UserRepository interface
func (r *Repository) GetUserByIdentity(provider domain.IdentityProvider, identifier string) (*domain.User, *domain.Identity, error) {
	row, err := r.querier.GetUserByIdentity(r.ctx, sqlcgen.GetUserByIdentityParams{
		Provider:   string(provider),
		Identifier: identifier,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil, fmt.Errorf("user or identity not found: %w", err)
		}
		return nil, nil, fmt.Errorf("failed to get user by identity: %w", err)
	}

	user := &domain.User{
		ID:        row.UserID,
		CreatedAt: row.UserCreatedAt.Time,
		UpdatedAt: row.UserUpdatedAt.Time,
		Disabled:  row.UserDisabled,
		Name:      row.UserName,
		Nickname:  row.UserNickname.String,
	}

	identity := &domain.Identity{
		Provider:   domain.IdentityProvider(row.IdentityProvider),
		Identifier: row.IdentityIdentifier,
		Credential: row.IdentityCredential,
		LastUsedAt: row.IdentityLastUsedAt.Time,
	}

	return user, identity, nil
}
