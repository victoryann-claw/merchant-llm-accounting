package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"merchant-llm-accounting/internal/model"
	"merchant-llm-accounting/pkg/db"
)

type RecordRepository struct {
	db *db.PostgresDB
}

func NewRecordRepository(database *db.PostgresDB) *RecordRepository {
	return &RecordRepository{db: database}
}

func (r *RecordRepository) Create(ctx context.Context, record *model.Record) error {
	query := `
		INSERT INTO records (merchant_id, record_type, occurred_at, metadata, total_amount, counterparty)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		record.MerchantID,
		record.RecordType,
		record.OccurredAt,
		record.Metadata,
		record.TotalAmount,
		record.Counterparty,
	).Scan(&record.ID, &record.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	return nil
}

func (r *RecordRepository) GetByID(ctx context.Context, id string) (*model.Record, error) {
	query := `
		SELECT id, merchant_id, record_type, occurred_at, metadata, total_amount, counterparty, created_at
		FROM records WHERE id = $1
	`

	var record model.Record
	var metadata []byte
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&record.ID,
		&record.MerchantID,
		&record.RecordType,
		&record.OccurredAt,
		&metadata,
		&record.TotalAmount,
		&record.Counterparty,
		&record.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get record: %w", err)
	}

	record.Metadata = json.RawMessage(metadata)
	return &record, nil
}

func (r *RecordRepository) ListByMerchant(ctx context.Context, merchantID string, start, end time.Time, recordType string) ([]*model.Record, error) {
	query := `
		SELECT id, merchant_id, record_type, occurred_at, metadata, total_amount, counterparty, created_at
		FROM records
		WHERE merchant_id = $1 AND occurred_at >= $2 AND occurred_at <= $3
	`
	args := []interface{}{merchantID, start, end}

	if recordType != "" {
		query += " AND record_type = $4"
		args = append(args, recordType)
	}

	query += " ORDER BY occurred_at DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list records: %w", err)
	}
	defer rows.Close()

	var records []*model.Record
	for rows.Next() {
		var record model.Record
		var metadata []byte
		if err := rows.Scan(
			&record.ID,
			&record.MerchantID,
			&record.RecordType,
			&record.OccurredAt,
			&metadata,
			&record.TotalAmount,
			&record.Counterparty,
			&record.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		record.Metadata = json.RawMessage(metadata)
		records = append(records, &record)
	}

	return records, nil
}
