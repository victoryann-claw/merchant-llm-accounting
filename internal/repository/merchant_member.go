package repository

import (
	"context"
	"database/sql"
	"fmt"

	"merchant-llm-accounting/internal/model"
	"merchant-llm-accounting/pkg/db"
)

type MerchantMemberRepository struct {
	db *db.PostgresDB
}

func NewMerchantMemberRepository(database *db.PostgresDB) *MerchantMemberRepository {
	return &MerchantMemberRepository{db: database}
}

// Create 创建成员关系
func (r *MerchantMemberRepository) Create(ctx context.Context, member *model.MerchantMember) error {
	query := `
		INSERT INTO merchant_member (merchant_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, joined_at
	`

	err := r.db.QueryRowContext(ctx, query, member.MerchantID, member.UserID, member.Role).
		Scan(&member.ID, &member.JoinedAt)

	if err != nil {
		return fmt.Errorf("failed to create merchant member: %w", err)
	}

	return nil
}

// GetByUserID 获取用户的所有商户成员关系
func (r *MerchantMemberRepository) GetByUserID(ctx context.Context, userID string) ([]*model.MerchantMember, error) {
	query := `
		SELECT id, merchant_id, user_id, role, joined_at
		FROM merchant_member WHERE user_id = $1
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query merchant members: %w", err)
	}
	defer rows.Close()

	var members []*model.MerchantMember
	for rows.Next() {
		var m model.MerchantMember
		if err := rows.Scan(&m.ID, &m.MerchantID, &m.UserID, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}

	return members, nil
}

// GetMerchantWithMembers 获取用户的所有商户（含成员信息）
func (r *MerchantMemberRepository) GetUserMerchants(ctx context.Context, userID string) ([]*model.Merchant, error) {
	query := `
		SELECT m.id, m.name, m.business_type, COALESCE(m.invite_code, ''), m.created_at, m.updated_at, mm.role
		FROM merchant m
		INNER JOIN merchant_member mm ON m.id = mm.merchant_id
		WHERE mm.user_id = $1
		ORDER BY mm.joined_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user merchants: %w", err)
	}
	defer rows.Close()

	var merchants []*model.Merchant
	for rows.Next() {
		var m model.Merchant
		var role string
		if err := rows.Scan(&m.ID, &m.Name, &m.BusinessType, &m.InviteCode, &m.CreatedAt, &m.UpdatedAt, &role); err != nil {
			return nil, err
		}
		merchants = append(merchants, &m)
	}

	return merchants, nil
}

// GetByMerchantAndUser 根据商户和用户获取成员
func (r *MerchantMemberRepository) GetByMerchantAndUser(ctx context.Context, merchantID, userID string) (*model.MerchantMember, error) {
	query := `
		SELECT id, merchant_id, user_id, role, joined_at
		FROM merchant_member WHERE merchant_id = $1 AND user_id = $2
	`

	var m model.MerchantMember
	err := r.db.QueryRowContext(ctx, query, merchantID, userID).Scan(
		&m.ID, &m.MerchantID, &m.UserID, &m.Role, &m.JoinedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant member: %w", err)
	}

	return &m, nil
}

// Delete 删除成员关系
func (r *MerchantMemberRepository) Delete(ctx context.Context, merchantID, userID string) error {
	query := `DELETE FROM merchant_member WHERE merchant_id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, query, merchantID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete merchant member: %w", err)
	}

	return nil
}
