package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	*sql.DB
}

func NewPostgresDB(host, port, user, password, dbname string) (*PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{db}, nil
}

func InitSchema(database *PostgresDB) error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "pgcrypto";

	-- 用户表（通过openid识别）
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		openid VARCHAR(100) UNIQUE NOT NULL,
		nickname VARCHAR(100),
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_users_openid ON users(openid);

	-- 商户表
	CREATE TABLE IF NOT EXISTS merchant (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		name VARCHAR(100) NOT NULL,
		business_type VARCHAR(50),
		invite_code VARCHAR(20) UNIQUE,
		created_at TIMESTAMP DEFAULT NOW(),
		updated_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_merchant_invite_code ON merchant(invite_code);

	-- 商户成员表（用户-商户关系）
	CREATE TABLE IF NOT EXISTS merchant_member (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		merchant_id UUID NOT NULL REFERENCES merchant(id) ON DELETE CASCADE,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		role VARCHAR(20) NOT NULL DEFAULT 'employee', -- owner/employee
		joined_at TIMESTAMP DEFAULT NOW(),
		UNIQUE(merchant_id, user_id)
	);

	CREATE INDEX IF NOT EXISTS idx_merchant_member_user ON merchant_member(user_id);
	CREATE INDEX IF NOT EXISTS idx_merchant_member_merchant ON merchant_member(merchant_id);

	-- 业务配置表
	CREATE TABLE IF NOT EXISTS business_config (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		merchant_id UUID REFERENCES merchant(id) ON DELETE CASCADE,
		config_key VARCHAR(100),
		config_value JSONB,
		created_at TIMESTAMP DEFAULT NOW(),
		UNIQUE(merchant_id, config_key)
	);

	-- 业务记录表
	CREATE TABLE IF NOT EXISTS records (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		merchant_id UUID NOT NULL REFERENCES merchant(id) ON DELETE CASCADE,
		record_type VARCHAR(50) NOT NULL,
		occurred_at TIMESTAMP NOT NULL,
		metadata JSONB NOT NULL,
		total_amount DECIMAL(12,2),
		counterparty VARCHAR(100),
		created_at TIMESTAMP DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_records_merchant_id ON records(merchant_id);
	CREATE INDEX IF NOT EXISTS idx_records_occurred_at ON records(occurred_at);
	CREATE INDEX IF NOT EXISTS idx_records_type ON records(record_type);
	CREATE INDEX IF NOT EXISTS idx_records_metadata ON records USING GIN (metadata);
	`

	_, err := database.Exec(schema)
	return err
}
