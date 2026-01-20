package repository

import (
	"context"

	"github.com/anthropics/pickle-go/apps/api/internal/model"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// UserRepository handles user data access
type UserRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE id = $1`
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByLineUserID finds a user by Line user ID
func (r *UserRepository) FindByLineUserID(ctx context.Context, lineUserID string) (*model.User, error) {
	var user model.User
	query := `SELECT * FROM users WHERE line_user_id = $1`
	err := r.db.GetContext(ctx, &user, query, lineUserID)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, line_user_id, display_name, avatar_url, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowxContext(ctx, query,
		user.ID, user.LineUserID, user.DisplayName, user.AvatarURL, user.Email,
	).StructScan(user)
}

// Update updates an existing user
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET display_name = $2, avatar_url = $3, email = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at`
	return r.db.QueryRowxContext(ctx, query,
		user.ID, user.DisplayName, user.AvatarURL, user.Email,
	).Scan(&user.UpdatedAt)
}

// Upsert creates or updates a user based on Line user ID
func (r *UserRepository) Upsert(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, line_user_id, display_name, avatar_url, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (line_user_id)
		DO UPDATE SET
			display_name = EXCLUDED.display_name,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = NOW()
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowxContext(ctx, query,
		user.ID, user.LineUserID, user.DisplayName, user.AvatarURL, user.Email,
	).StructScan(user)
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// Exists checks if a user exists by ID
func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`
	err := r.db.GetContext(ctx, &exists, query, id)
	return exists, err
}

// ExistsByLineUserID checks if a user exists by Line user ID
func (r *UserRepository) ExistsByLineUserID(ctx context.Context, lineUserID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE line_user_id = $1)`
	err := r.db.GetContext(ctx, &exists, query, lineUserID)
	return exists, err
}
