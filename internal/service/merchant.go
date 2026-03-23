package service

import (
	"context"
	"fmt"

	"merchant-llm-accounting/internal/model"
	"merchant-llm-accounting/internal/repository"
)

type MerchantService struct {
	merchantRepo *repository.MerchantRepository
}

func NewMerchantService(merchantRepo *repository.MerchantRepository) *MerchantService {
	return &MerchantService{merchantRepo: merchantRepo}
}

func (s *MerchantService) Create(ctx context.Context, merchant *model.Merchant) error {
	if merchant.Name == "" {
		return fmt.Errorf("merchant name is required")
	}
	return s.merchantRepo.Create(ctx, merchant)
}

func (s *MerchantService) GetByID(ctx context.Context, id string) (*model.Merchant, error) {
	return s.merchantRepo.GetByID(ctx, id)
}
