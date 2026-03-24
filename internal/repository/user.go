package repository

import (
	"context"
	"database/sql"
	"fmt"

	"merchant-llm-accounting/internal/model"
	"merchant-llm-accounting/pkg/db"
)

type UserRepository struct {
	db *db.PostgresDB
}

func NewUserRepository(database *db.PostgresDB) *UserRepository {
	return &UserRepository{db: database}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (openid, nickname)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, user.OpenID, user.Nickname).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByOpenID 根据openid获取用户
func (r *UserRepository) GetByOpenID(ctx context.Context, openid string) (*model.User, error) {
	query := `
		SELECT id, openid, COALESCE(nickname, ''), created_at, updated_at
		FROM users WHERE openid = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, openid).Scan(
		&user.ID,
		&user.OpenID,
		&user.Nickname,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by openid: %w", err)
	}

	return &user, nil
}

// GetByID 根据ID获取用户
func (r *UserRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, openid, COALESCE(nickname, ''), created_at, updated_at
		FROM users WHERE id = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.OpenID,
		&user.Nickname,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

// Update 更新用户
func (r *UserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users SET nickname = $1, updated_at = NOW()
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, user.Nickname, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}
