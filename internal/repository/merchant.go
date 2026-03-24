package repository

import (
	"context"
	"database/sql"
	"fmt"

	"merchant-llm-accounting/internal/model"
	"merchant-llm-accounting/pkg/db"
)

type MerchantRepository struct {
	db *db.PostgresDB
}

func NewMerchantRepository(database *db.PostgresDB) *MerchantRepository {
	return &MerchantRepository{db: database}
}

func (r *MerchantRepository) Create(ctx context.Context, merchant *model.Merchant) error {
	query := `
		INSERT INTO merchant (name, business_type, openid)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	var openid sql.NullString
	if merchant.OpenID != "" {
		openid = sql.NullString{String: merchant.OpenID, Valid: true}
	}

	err := r.db.QueryRowContext(ctx, query, merchant.Name, merchant.BusinessType, openid).
		Scan(&merchant.ID, &merchant.CreatedAt, &merchant.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create merchant: %w", err)
	}

	return nil
}

func (r *MerchantRepository) GetByID(ctx context.Context, id string) (*model.Merchant, error) {
	query := `
		SELECT id, COALESCE(openid, ''), name, business_type, created_at, updated_at
		FROM merchant WHERE id = $1
	`

	var merchant model.Merchant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&merchant.ID,
		&merchant.OpenID,
		&merchant.Name,
		&merchant.BusinessType,
		&merchant.CreatedAt,
		&merchant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	return &merchant, nil
}

// GetByOpenID 根据OpenID获取商户
func (r *MerchantRepository) GetByOpenID(ctx context.Context, openid string) (*model.Merchant, error) {
	query := `
		SELECT id, openid, name, business_type, created_at, updated_at
		FROM merchant WHERE openid = $1
	`

	var merchant model.Merchant
	err := r.db.QueryRowContext(ctx, query, openid).Scan(
		&merchant.ID,
		&merchant.OpenID,
		&merchant.Name,
		&merchant.BusinessType,
		&merchant.CreatedAt,
		&merchant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant by openid: %w", err)
	}

	return &merchant, nil
}

// Update 更新商户信息
func (r *MerchantRepository) Update(ctx context.Context, merchant *model.Merchant) error {
	query := `
		UPDATE merchant SET name = $1, business_type = $2, updated_at = NOW()
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, merchant.Name, merchant.BusinessType, merchant.ID)
	if err != nil {
		return fmt.Errorf("failed to update merchant: %w", err)
	}

	return nil
}
