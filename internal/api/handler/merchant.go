package handler

import (
	"net/http"

	"merchant-llm-accounting/internal/model"
	"merchant-llm-accounting/internal/service"

	"github.com/gin-gonic/gin"
)

type MerchantHandler struct {
	merchantService *service.MerchantService
}

func NewMerchantHandler(merchantService *service.MerchantService) *MerchantHandler {
	return &MerchantHandler{merchantService: merchantService}
}

type CreateMerchantRequest struct {
	Name         string `json:"name" binding:"required"`
	BusinessType string `json:"business_type"`
}

func (h *MerchantHandler) Create(c *gin.Context) {
	var req CreateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	merchant := &model.Merchant{
		Name:         req.Name,
		BusinessType: req.BusinessType,
	}

	if err := h.merchantService.Create(c.Request.Context(), merchant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, merchant)
}
