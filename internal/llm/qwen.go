package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// 通义千问适配器
type QwenAdapter struct {
	APIKey string
	Model  string
}

func NewQwenAdapter(apiKey string) *QwenAdapter {
	return &QwenAdapter{
		APIKey: apiKey,
		Model:  "qwen-turbo",
	}
}

func (q *QwenAdapter) Name() string {
	return "qwen"
}

func (q *QwenAdapter) Parse(ctx context.Context, req LLMRequest) (*LLMResponse, error) {
	systemPrompt := q.buildSystemPrompt(req.BusinessType)
	userPrompt := req.UserInput

	payload := map[string]interface{}{
		"model": q.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"temperature": 0.7,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+q.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Qwen API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Qwen API error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("invalid Qwen response format")
	}

	content, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to extract content")
	}

	// 解析LLM返回的JSON
	return q.parseResponse(content)
}

func (q *QwenAdapter) buildSystemPrompt(businessType string) string {
	base := `你是一个商户记账助手，帮助用户将自然语言输入转换为结构化的记账记录。
请根据用户的输入，提取以下信息：
- record_type: 记录类型 (purchase=进货, delivery=送货, payment=付款)
- metadata: 业务相关数据（由具体业务决定，灵活字段）
- total_amount: 总金额
- counterparty: 对方（供应商或客户）
- suggestions: 建议（如有）

只返回JSON格式，不要其他内容。`

	if businessType == "" {
		return base
	}

	typeHints := map[string]string{
		"fish":      "你是鱼类商贩，需要记录品种、重量、鲜活度、批发价等信息",
		"vegetable": "你是蔬菜商贩，需要记录菜名、单价、进货量、供应商等信息",
		"oil":       "你是粮油商贩，需要记录规格、多少升/桶、单价等信息",
	}

	if hint, ok := typeHints[businessType]; ok {
		return base + "\n\n" + hint
	}
	return base
}

func (q *QwenAdapter) parseResponse(content string) (*LLMResponse, error) {
	// 尝试提取JSON（可能包含在markdown代码块中）
	jsonStr := content
	if len(content) > 6 && content[:6] == "```json" {
		start := 7
		end := len(content) - 3
		for start < end && content[start] == '\n' {
			start++
		}
		for end > start && (content[end] == '`' || content[end] == '\n') {
			end--
		}
		jsonStr = content[start : end+1]
	}

	var resp LLMResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &resp, nil
}
