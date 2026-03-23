package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"merchant-llm-accounting/internal/llm"
	"merchant-llm-accounting/internal/model"
	"merchant-llm-accounting/internal/repository"
)

type RecordService struct {
	recordRepo  *repository.RecordRepository
	merchantRepo *repository.MerchantRepository
	llmAdapter  llm.LLMAdapter
}

func NewRecordService(
	recordRepo *repository.RecordRepository,
	merchantRepo *repository.MerchantRepository,
	llmAdapter llm.LLMAdapter,
) *RecordService {
	return &RecordService{
		recordRepo:   recordRepo,
		merchantRepo: merchantRepo,
		llmAdapter:  llmAdapter,
	}
}

// CreateRecordByVoice 通过语音创建记录
func (s *RecordService) CreateRecordByVoice(ctx context.Context, merchantID string, audioData []byte) (*model.Record, error) {
	// 1. 获取商户信息
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("merchant not found")
	}

	// 2. 调用腾讯云ASR识别语音
	asrAdapter := NewTencentASR(
		os.Getenv("TENCENT_ASR_SECRET_ID"),
		os.Getenv("TENCENT_ASR_SECRET_KEY"),
		0, // appId
	)
	text, err := asrAdapter.SimpleRecognize(ctx, audioData)
	if err != nil {
		return nil, fmt.Errorf("failed to recognize voice: %w", err)
	}

	// 3. 用识别出的文字创建记录
	return s.CreateRecord(ctx, merchantID, text)
}

// CreateRecord 将用户输入交给LLM解析，然后存储
func (s *RecordService) CreateRecord(ctx context.Context, merchantID string, userInput string) (*model.Record, error) {
	// 1. 获取商户信息
	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get merchant: %w", err)
	}
	if merchant == nil {
		return nil, fmt.Errorf("merchant not found")
	}

	// 2. 调用LLM解析
	llmResp, err := s.llmAdapter.Parse(ctx, llm.LLMRequest{
		MerchantID:   merchantID,
		BusinessType: merchant.BusinessType,
		UserInput:    userInput,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	// 3. 构建记录
	metadata, err := json.Marshal(llmResp.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	var totalAmount *float64
	if llmResp.TotalAmount > 0 {
		totalAmount = &llmResp.TotalAmount
	}

	record := &model.Record{
		MerchantID:   merchantID,
		RecordType:   llmResp.RecordType,
		OccurredAt:   time.Now(),
		Metadata:     metadata,
		TotalAmount:  totalAmount,
		Counterparty: llmResp.Counterparty,
	}

	// 4. 存储
	if err := s.recordRepo.Create(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create record: %w", err)
	}

	return record, nil
}

// ListRecords 查询记录
func (s *RecordService) ListRecords(ctx context.Context, merchantID string, start, end time.Time, recordType string) ([]*model.Record, error) {
	return s.recordRepo.ListByMerchant(ctx, merchantID, start, end, recordType)
}
