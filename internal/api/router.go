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
	userRepo := repository.NewUserRepository(database)
	merchantRepo := repository.NewMerchantRepository(database)
	memberRepo := repository.NewMerchantMemberRepository(database)
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
	authHandler := handler.NewAuthHandler(userRepo, merchantRepo, memberRepo)

	// API路由
	v1 := r.Group("/api/v1")
	{
		// 微信登录/用户相关
		v1.POST("/auth/wechat", authHandler.WeChatLogin)
		v1.POST("/auth/create-merchant", authHandler.CreateMerchant)
		v1.POST("/auth/join-merchant", authHandler.JoinMerchant)

		// 商户相关
		v1.POST("/merchant", merchantHandler.Create)

		// 记录相关
		v1.POST("/record", recordHandler.Create)
		v1.POST("/record/voice", recordHandler.CreateByVoice)
		v1.POST("/record/image", recordHandler.CreateByImage)
		v1.GET("/records", recordHandler.List)
		v1.GET("/records/:id", recordHandler.Get)
	}
	
	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
