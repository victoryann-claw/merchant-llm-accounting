package main

import (
	"log"
	"os"

	"merchant-llm-accounting/internal/api"
	"merchant-llm-accounting/pkg/db"
)

func main() {
	// 初始化数据库连接
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "merchant_accounting")

	database, err := db.NewPostgresDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer database.Close()

	// 初始化数据库表
	if err := db.InitSchema(database); err != nil {
		log.Fatalf("Failed to init schema: %v", err)
	}

	// 启动API服务
	port := getEnv("API_PORT", "8080")
	router := api.SetupRouter(database)

	log.Printf("Server starting on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
