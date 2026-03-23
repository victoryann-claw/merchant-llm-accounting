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

// QwenVLAdapter 通义千问视觉语言模型
type QwenVLAdapter struct {
	APIKey string
	Model  string // qwen-vl-plus 或 qwen-vl-max
}

func NewQwenVLAdapter(apiKey string) *QwenVLAdapter {
	return &QwenVLAdapter{
		APIKey: apiKey,
		Model:  "qwen-vl-plus", // 默认用 plus，支持图片输入
	}
}

func (q *QwenVLAdapter) Name() string {
	return "qwen_vl"
}

// ImageRecordRequest 图片记账请求
type ImageRecordRequest struct {
	ImageURL string // 图片URL或base64
}

// ImageRecordResponse 图片识别结果
type ImageRecordResponse struct {
	RecordType   string                 `json:"record_type"`   // purchase/delivery
	Counterparty string                 `json:"counterparty"`   // 对方（供应商/客户）
	TotalAmount  float64                `json:"total_amount"`  // 总金额
	Items        []ImageRecordItem      `json:"items"`         // 商品明细
	RawText      string                 `json:"raw_text"`      // 原始识别文字
	Metadata     map[string]interface{} `json:"metadata"`      // 其他信息
}

type ImageRecordItem struct {
	Name  string  `json:"name"`  // 商品名
	Qty   float64 `json:"qty"`   // 数量
	Unit  string  `json:"unit"`  // 单位
	Price float64 `json:"price"` // 单价
}

// RecognizeImage 识别图片中的单据
func (q *QwenVLAdapter) RecognizeImage(ctx context.Context, imageData []byte) (*ImageRecordResponse, error) {
	// 构建 prompt
	systemPrompt := `你是一个商户记账助手。用户会发送进货单或送货单的图片，请你识别其中的信息。

请提取以下内容并返回JSON格式：
- record_type: "purchase"(进货单) 或 "delivery"(送货单)
- counterparty: 对方名称（供应商或客户）
- total_amount: 总金额
- items: 商品明细列表，包含name(商品名)、qty(数量)、unit(单位)、price(单价)
- raw_text: 图片中识别出的原始文字

只返回JSON，不要其他内容。如果图片不清晰或无法识别，返回空JSON {}。`

	userPrompt := "请识别这张图片中的单据信息："

	// 构建消息 - 支持图片
	messages := []map[string]interface{}{
		{
			"role": "system",
			"content": []map[string]string{
				{"text": systemPrompt},
			},
		},
		{
			"role": "user",
			"content": append([]map[string]interface{}{
				{"text": userPrompt},
			}, map[string]interface{}{
				"image": "data:image/jpeg;base64," + base64Encode(imageData),
			}),
		},
	}

	payload := map[string]interface{}{
		"model":       q.Model,
		"messages":    messages,
		"max_tokens":  1000,
		"temperature": 0.1,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+q.APIKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call Qwen VL API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Qwen VL API error: %s", string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return nil, fmt.Errorf("invalid response format")
	}

	content, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to extract content")
	}

	// 解析LLM返回的JSON
	return q.parseResponse(content)
}

func (q *QwenVLAdapter) parseResponse(content string) (*ImageRecordResponse, error) {
	// 尝试提取JSON
	jsonStr := content
	
	// 去掉可能的markdown代码块
	if len(content) >= 7 && content[:7] == "```json" {
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

	var resp ImageRecordResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	return &resp, nil
}

// base64Encode 简单的base64编码
func base64Encode(data []byte) string {
	const encodeStd = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	const padding = '='

	if len(data) == 0 {
		return ""
	}

	result := make([]byte, (len(data)+2)/3*4)
	for i, j := 0, 0; i < len(data); i, j = i+3, j+4 {
		var val uint32
		switch len(data) - i {
		case 1:
			val = uint32(data[i]) << 16
			result[j] = encodeStd[val>>18&0x3F]
			result[j+1] = encodeStd[val>>12&0x3F]
			result[j+2] = padding
			result[j+3] = padding
		case 2:
			val = uint32(data[i])<<16 | uint32(data[i+1])<<8
			result[j] = encodeStd[val>>18&0x3F]
			result[j+1] = encodeStd[val>>12&0x3F]
			result[j+2] = encodeStd[val>>6&0x3F]
			result[j+3] = padding
		default:
			val = uint32(data[i])<<16 | uint32(data[i+1])<<8 | uint32(data[i+2])
			result[j] = encodeStd[val>>18&0x3F]
			result[j+1] = encodeStd[val>>12&0x3F]
			result[j+2] = encodeStd[val>>6&0x3F]
			result[j+3] = encodeStd[val&0x3F]
		}
	}
	return string(result)
}
