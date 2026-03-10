package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kshitijdhara/blog/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserExists = errors.New("username or email already exists")

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
}

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		INSERT INTO users (username, email, password_hash, display_name, bio)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, password_hash, display_name, bio, avatar_url, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.DisplayName, user.Bio,
	).StructScan(user)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserExists
		}
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE id = $1", id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	var user domain.User
	err := r.db.GetContext(ctx, &user, "SELECT * FROM users WHERE username = $1", username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	return &user, err
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	query := `
		UPDATE users
		SET username = $1, display_name = $2, bio = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, username, email, password_hash, display_name, bio, avatar_url, banner_url, created_at, updated_at`

	err := r.db.QueryRowxContext(ctx, query,
		user.Username, user.DisplayName, user.Bio, user.ID,
	).StructScan(user)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUserExists
		}
		return nil, err
	}
	return user, nil
}

func isUniqueViolation(err error) bool {
	return err != nil && (contains(err.Error(), "duplicate key") || contains(err.Error(), "unique constraint"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
