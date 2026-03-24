package model

import (
	"encoding/json"
	"time"
)

// 商户
type Merchant struct {
	ID           string    `json:"id"`
	OpenID       string    `json:"openid,omitempty"`
	Name         string    `json:"name"`
	BusinessType string    `json:"business_type"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// 业务配置
type BusinessConfig struct {
	ID           string          `json:"id"`
	MerchantID   string          `json:"merchant_id"`
	ConfigKey    string          `json:"config_key"`
	ConfigValue  json.RawMessage `json:"config_value"`
	CreatedAt    time.Time       `json:"created_at"`
}

// 业务记录
type Record struct {
	ID           string          `json:"id"`
	MerchantID   string          `json:"merchant_id"`
	RecordType   string          `json:"record_type"` // purchase/delivery/payment
	OccurredAt   time.Time       `json:"occurred_at"`
	Metadata     json.RawMessage `json:"metadata"`
	TotalAmount  *float64        `json:"total_amount,omitempty"`
	Counterparty string          `json:"counterparty,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
}

// 记录类型常量
const (
	RecordTypePurchase = "purchase"
	RecordTypeDelivery = "delivery"
	RecordTypePayment  = "payment"
)
