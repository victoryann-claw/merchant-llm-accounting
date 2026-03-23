package handler

import (
	"net/http"
	"time"

	"merchant-llm-accounting/internal/service"

	"github.com/gin-gonic/gin"
)

type RecordHandler struct {
	recordService *service.RecordService
}

func NewRecordHandler(recordService *service.RecordService) *RecordHandler {
	return &RecordHandler{recordService: recordService}
}

type CreateRecordRequest struct {
	MerchantID string `json:"merchant_id" binding:"required"`
	UserInput  string `json:"user_input" binding:"required"`
}

func (h *RecordHandler) Create(c *gin.Context) {
	var req CreateRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	record, err := h.recordService.CreateRecord(c.Request.Context(), req.MerchantID, req.UserInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

// VoiceRecordRequest 语音录制请求（支持直接上传音频）
type VoiceRecordRequest struct {
	MerchantID string `form:"merchant_id" binding:"required"`
}

// CreateByVoice 通过语音创建记录
// 小程序上传音频文件，后端识别后创建记录
func (h *RecordHandler) CreateByVoice(c *gin.Context) {
	merchantID := c.PostForm("merchant_id")
	if merchantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "merchant_id is required"})
		return
	}

	// 获取音频文件
	file, _, err := c.Request.FormFile("audio")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "audio file is required"})
		return
	}
	defer file.Close()

	// 读取音频数据
	audioData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read audio"})
		return
	}

	record, err := h.recordService.CreateRecordByVoice(c.Request.Context(), merchantID, audioData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, record)
}

type ListRecordsRequest struct {
	MerchantID string `form:"merchant_id" binding:"required"`
	Start      string `form:"start"`
	End        string `form:"end"`
	Type       string `form:"type"`
}

func (h *RecordHandler) List(c *gin.Context) {
	var req ListRecordsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 默认查最近30天
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	if req.Start != "" {
		if t, err := time.Parse("2006-01-02", req.Start); err == nil {
			start = t
		}
	}
	if req.End != "" {
		if t, err := time.Parse("2006-01-02", req.End); err == nil {
			end = t
		}
	}

	records, err := h.recordService.ListRecords(c.Request.Context(), req.MerchantID, start, end, req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, records)
}

func (h *RecordHandler) Get(c *gin.Context) {
	id := c.Param("id")
	// TODO: 实现GetByID
	c.JSON(http.StatusOK, gin.H{"id": id})
}
