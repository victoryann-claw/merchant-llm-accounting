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
	merchantRepo *repository.MerchantRepository
}

func NewAuthHandler(merchantRepo *repository.MerchantRepository) *AuthHandler {
	return &AuthHandler{merchantRepo: merchantRepo}
}

// WeChatLoginRequest 微信登录请求
type WeChatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// WeChatLogin 微信登录/注册
// 小程序调用 wx.login() 获取 code，传给后端
// 后端用 code 调用微信接口获取 openid
// 如果商户已存在则返回，如果不存在则自动创建
func (h *AuthHandler) WeChatLogin(c *gin.Context) {
	var req WeChatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}

	// 调用微信接口获取 openid
	openid, err := h.getWeChatOpenID(req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get openid: " + err.Error()})
		return
	}

	// 查找是否已存在该openid的商户
	merchant, err := h.merchantRepo.GetByOpenID(c.Request.Context(), openid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query merchant"})
		return
	}

	// 如果不存在，自动创建
	if merchant == nil {
		merchant = &model.Merchant{
			OpenID:       openid,
			Name:         "我的店铺",
			BusinessType: "fish",
		}
		if err := h.merchantRepo.Create(c.Request.Context(), merchant); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create merchant"})
			return
		}
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
