package main

import (
	"fmt"
	"medical-qa-assistant/api"
	"medical-qa-assistant/internal/config"
	"medical-qa-assistant/internal/logger"
	"medical-qa-assistant/internal/models"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 首先初始化结构化日志
	logger.Init()
	defer logger.Sync()

	// 加载 .env 文件用于本地开发；如果不存在则使用现有环境变量
	if err := godotenv.Load(); err != nil {
		logger.L.Warn("no .env file loaded", zap.Error(err))
	}

	// 加载配置
	cfg := config.Load()
	logger.L.Info("configuration loaded",
		zap.String("db_host", cfg.DBHost),
		zap.String("db_name", cfg.DBName),
		zap.String("llm_provider", cfg.LLMProvider),
		zap.String("port", cfg.Port),
	)

	// 连接数据库
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.L.Fatal("failed to connect to database", zap.Error(err))
	}

	// 自动迁移（文档块和向量存储在 Chroma 中，不在 MySQL）
	if err := db.AutoMigrate(&models.User{}, &models.Document{}); err != nil {
		logger.L.Fatal("failed to migrate database", zap.Error(err))
	}

	// 设置路由
	router := api.SetupRoutes(db, cfg)

	// 启动服务器
	addr := fmt.Sprintf(":%s", cfg.Port)
	logger.L.Info("server starting",
		zap.String("addr", addr),
		zap.String("llm_provider", cfg.LLMProvider),
		zap.String("chroma_base_url", cfg.ChromaBaseURL),
		zap.String("chroma_collection", cfg.ChromaCollection),
	)
	if err := router.Run(addr); err != nil {
		logger.L.Fatal("failed to start server", zap.Error(err))
	}

}
