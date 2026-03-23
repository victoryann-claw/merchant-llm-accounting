package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go-intl/tencentcloud"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go-intl/tencentcloud/common/profile"
	asr "github.com/tencentcloud/tencentcloud-sdk-go-intl/tencentcloud/asr/v20190614"
)

// TencentASR 腾讯云语音识别
type TencentASR struct {
	secretId  string
	secretKey string
	appId     uint64
}

func NewTencentASR(secretId, secretKey string, appId uint64) *TencentASR {
	return &TencentASR{
		secretId:  secretId,
		secretKey: secretKey,
		appId:     appId,
	}
}

// Recognize 识别短音频文件（.wav, .mp3）
func (t *TencentASR) Recognize(ctx context.Context, audioPath string) (string, error) {
	// 读取音频文件
	file, err := os.Open(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %w", err)
	}

	// 构建请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加文件
	part, err := writer.CreateFormFile("audio", audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}
	part.Write(data)

	// 添加参数
	writer.WriteField("secretid", t.secretId)
	writer.WriteField("timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	writer.WriteField("expired", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))
	writer.WriteField("projectid", "0")
	writer.WriteField("subproduct", "short_audio")
	writer.WriteField("enginemodeltype", "16k_zh")

	writer.Close()

	// 发送请求
	url := fmt.Sprintf("https://asr.cloud.tencent.com/ASR/upload/%d", t.appId)
	req, err := http.NewRequestWithContext(ctx, "POST", url, body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call Tencent ASR: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Tencent ASR error: %s", string(respBody))
	}

	// 解析响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// 检查错误
	if code, ok := result["code"].(float64); ok && code != 0 {
		return "", fmt.Errorf("ASR error code: %.0f, msg: %v", code, result["msg"])
	}

	// 提取识别结果
	if text, ok := result["text"].(string); ok {
		return text, nil
	}

	return "", fmt.Errorf("no text in response")
}

// RecognizeByUrl 通过URL识别（腾讯云存储的音频）
func (t *TencentASR) RecognizeByUrl(ctx context.Context, audioUrl string) (string, error) {
	// 使用腾讯云ASR REST API
	url := fmt.Sprintf("https://asr.cloud.tencent.com/ASR/v1/%d", t.appId)

	payload := map[string]interface{}{
		"url":             audioUrl,
		"enginemodeltype": "16k_zh",
		"projectid":       0,
		"subproduct":      "short_audio",
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("XTC-Authorization", fmt.Sprintf("secretid=%s;timestamp=%d;expired=%d",
		t.secretId, time.Now().Unix(), time.Now().Add(time.Hour).Unix()))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	if text, ok := result["text"].(string); ok {
		return text, nil
	}

	return "", fmt.Errorf("ASR failed: %s", string(respBody))
}

// SimpleRecognize 简单的语音识别（测试用）
// 直接传入base64编码的音频数据
func (t *TencentASR) SimpleRecognize(ctx context.Context, audioData []byte) (string, error) {
	// 获取签名（简化版，实际生产应该用SDK）
	
	// 实际生产中建议使用腾讯云SDK，简化处理：
	// 1. 将音频数据临时保存
	// 2. 调用 Recognize 方法
	// 3. 删除临时文件
	
	tmpFile := "/tmp/asr_input_" + strconv.FormatInt(time.Now().UnixNano(), 10) + ".wav"
	defer os.Remove(tmpFile)

	if err := os.WriteFile(tmpFile, audioData, 0644); err != nil {
		return "", err
	}

	return t.Recognize(ctx, tmpFile)
}
