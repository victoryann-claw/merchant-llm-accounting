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
		INSERT INTO merchant (name, business_type)
		VALUES ($1, $2)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query, merchant.Name, merchant.BusinessType).
		Scan(&merchant.ID, &merchant.CreatedAt, &merchant.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create merchant: %w", err)
	}

	return nil
}

func (r *MerchantRepository) GetByID(ctx context.Context, id string) (*model.Merchant, error) {
	query := `
		SELECT id, name, business_type, created_at, updated_at
		FROM merchant WHERE id = $1
	`

	var merchant model.Merchant
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&merchant.ID,
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
