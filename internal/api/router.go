package api

import (
	"os"

	"merchant-llm-accounting/internal/api/handler"
	"merchant-llm-accounting/internal/api/middleware"
	"merchant-llm-accounting/internal/llm"
	"merchant-llm-accounting/internal/repository"
	"merchant-llm-accounting/internal/service"
	"merchant-llm-accounting/pkg/db"

	"github.com/gin-gonic/gin"
)

func SetupRouter(database *db.PostgresDB) *gin.Engine {
	r := gin.Default()

	// 中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Recovery())

	// 初始化仓库
	merchantRepo := repository.NewMerchantRepository(database)
	recordRepo := repository.NewRecordRepository(database)

	// 初始化LLM适配器（通义千问）
	qwenAdapter := llm.NewQwenAdapter(
		os.Getenv("QWEN_API_KEY"),
		os.Getenv("QWEN_MODEL"),
	)

	// 初始化服务
	recordService := service.NewRecordService(recordRepo, merchantRepo, qwenAdapter)
	merchantService := service.NewMerchantService(merchantRepo)

	// 初始化处理器
	merchantHandler := handler.NewMerchantHandler(merchantService)
	recordHandler := handler.NewRecordHandler(recordService)

	// API路由
	v1 := r.Group("/api/v1")
	{
		// 商户相关
		v1.POST("/merchant", merchantHandler.Create)

		// 记录相关
		v1.POST("/record", recordHandler.Create)
		v1.POST("/record/voice", recordHandler.CreateByVoice) // 语音录入
		v1.POST("/record/image", recordHandler.CreateByImage) // 图片录入
		v1.GET("/records", recordHandler.List)
		v1.GET("/records/:id", recordHandler.Get)
	}
	
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
