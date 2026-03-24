package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"merchant-llm-accounting/internal/model"
	"merchant-llm-accounting/internal/repository"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	userRepo         *repository.UserRepository
	merchantRepo     *repository.MerchantRepository
	memberRepo       *repository.MerchantMemberRepository
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	merchantRepo *repository.MerchantRepository,
	memberRepo *repository.MerchantMemberRepository,
) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		merchantRepo: merchantRepo,
		memberRepo:   memberRepo,
	}
}

// WeChatLoginRequest 微信登录请求
type WeChatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User      *model.User        `json:"user"`
	Merchants []*model.Merchant  `json:"merchants"`
	HasMerchant bool              `json:"has_merchant"` // 是否有商户
}

// WeChatLogin 微信登录
func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	// 调用微信接口获取 openid
	openid, err := h.getWeChatOpenID(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get openid"})
		return
	}

	// 获取或创建用户
	user, err := h.userRepo.GetByOpenID(c.Request.Context(), openid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query user"})
		return
	}

	if user == nil {
		user = &model.User{
			OpenID: openid,
		}
		if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}
	}

	// 获取用户的商户列表
	merchants, err := h.memberRepo.GetUserMerchants(c.Request.Context(), user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query merchants"})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		User:       user,
		Merchants:  merchants,
		HasMerchant: len(merchants) > 0,
	})
}

// CreateMerchantRequest 创建商户请求
type CreateMerchantRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	Name         string `json:"name" binding:"required"`
	BusinessType string `json:"business_type" binding:"required"`
}

// CreateMerchant 创建商户
func (h *AuthHandler) CreateMerchant(c *gin.Context) {
	var req CreateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}

	// 创建商户
	merchant := &model.Merchant{
		Name:         req.Name,
		BusinessType: req.BusinessType,
	}
	if err := h.merchantRepo.Create(c.Request.Context(), merchant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create merchant"})
		return
	}

	// 添加创建者为owner
	member := &model.MerchantMember{
		MerchantID: merchant.ID,
		UserID:     req.UserID,
		Role:       model.RoleOwner,
	}
	if err := h.memberRepo.Create(c.Request.Context(), member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add member"})
		return
	}

	c.JSON(http.StatusOK, merchant)
}

// JoinMerchantRequest 加入商户请求
type JoinMerchantRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	InviteCode  string `json:"invite_code" binding:"required"`
}

// JoinMerchant 通过邀请码加入商户
func (h *AuthHandler) JoinMerchant(c *gin.Context) {
	var req JoinMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid params"})
		return
	}

	// 通过邀请码查找商户
	merchant, err := h.merchantRepo.GetByInviteCode(c.Request.Context(), req.InviteCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query merchant"})
		return
	}

	if merchant == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite code"})
		return
	}

	// 检查是否已经是成员
	existing, err := h.memberRepo.GetByMerchantAndUser(c.Request.Context(), merchant.ID, req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check membership"})
		return
	}

	if existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "already joined"})
		return
	}

	// 添加为成员
	member := &model.MerchantMember{
		MerchantID: merchant.ID,
		UserID:     req.UserID,
		Role:       model.RoleEmployee,
	}
	if err := h.memberRepo.Create(c.Request.Context(), member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join merchant"})
		return
	}

	c.JSON(http.StatusOK, merchant)
}

// getWeChatOpenID 通过code获取openid
func (h *AuthHandler) getWeChatOpenID(code string) (string, error) {
	appID := os.Getenv("WECHAT_APP_ID")
	appSecret := os.Getenv("WECHAT_APP_SECRET")

	if appID == "" || appSecret == "" {
		// 测试环境，返回模拟openid
		return "test_openid_" + code[:8], nil
	}

	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		appID, appSecret, code,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if errMsg, ok := result["errmsg"]; ok {
		return "", fmt.Errorf("wechat error: %v", errMsg)
	}

	openid, ok := result["openid"].(string)
	if !ok {
		return "", fmt.Errorf("openid not found in response")
	}

	return openid, nil
}
