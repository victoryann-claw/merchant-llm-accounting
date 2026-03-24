package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"

	"merchant-llm-accounting/internal/model"
	"merchant-llm-accounting/pkg/db"
)

type MerchantRepository struct {
	db *db.PostgresDB
}

func NewMerchantRepository(database *db.PostgresDB) *MerchantRepository {
	return &MerchantRepository{db: database}
}

// generateInviteCode 生成6位邀请码
func generateInviteCode() string {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		code[i] = chars[n.Int64()]
	}
	return string(code)
}

// Create 创建商户（自动生成邀请码）
func (r *MerchantRepository) Create(ctx context.Context, merchant *model.Merchant) error {
	merchant.InviteCode = generateInviteCode()

	query := `
		INSERT INTO merchant (name, business_type, invite_code)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, merchant.Name, merchant.BusinessType, merchant.InviteCode).
		Scan(&merchant.ID, &merchant.CreatedAt, &merchant.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create merchant: %w", err)
	}

	return nil
}

// GetByID 获取商户
func (r *MerchantRepository) GetByID(ctx context.Context, id string) (*model.Merchant, error) {
	query := `
		SELECT id, name, business_type, COALESCE(invite_code, ''), created_at, updated_at
		FROM merchant WHERE id = $1
	`

	var m model.Merchant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.Name, &m.BusinessType, &m.InviteCode, &m.CreatedAt, &m.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}

	return &m, nil
}

// GetByInviteCode 根据邀请码获取商户
func (r *MerchantRepository) GetByInviteCode(ctx context.Context, code string) (*model.Merchant, error) {
	query := `
		SELECT id, name, business_type, COALESCE(invite_code, ''), created_at, updated_at
		FROM merchant WHERE invite_code = $1
	`

	var m model.Merchant
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&m.ID, &m.Name, &m.BusinessType, &m.InviteCode, &m.CreatedAt, &m.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant by invite code: %w", err)
	}

	return &m, nil
}

// Update 更新商户
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
