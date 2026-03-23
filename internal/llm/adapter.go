package llm

import (
	"context"
)

// LLM 请求
type LLMRequest struct {
	MerchantID   string                 `json:"merchant_id"`
	BusinessType string                 `json:"business_type"`
	UserInput    string                 `json:"user_input"`
	Context      map[string]interface{} `json:"context,omitempty"`
}

// LLM 响应
type LLMResponse struct {
	RecordType   string                 `json:"record_type"`
	Metadata     map[string]interface{} `json:"metadata"`
	TotalAmount  float64                `json:"total_amount"`
	Counterparty string                 `json:"counterparty"`
	Suggestions  []string               `json:"suggestions,omitempty"`
}

// LLM 适配器接口
type LLMAdapter interface {
	// Parse 解析用户输入，生成结构化记录
	Parse(ctx context.Context, req LLMRequest) (*LLMResponse, error)
	// Name 返回适配器名称
	Name() string
}
