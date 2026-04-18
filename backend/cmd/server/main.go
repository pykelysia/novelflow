package main

import (
	"fmt"
	"log"

	"novelflow/backend/internal"
	"novelflow/backend/internal/servicecontext"
	"novelflow/cache"
	"novelflow/config"
	"novelflow/database"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// 加载配置
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err := database.InitDB()
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	// 初始化 Redis
	redisClient, err := cache.InitRedis()
	if err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}
	defer redisClient.Close()

	svc := servicecontext.NewServiceContext()
	svc.RedisClient = redisClient
	// 创建 Gin 路由器
	router := gin.Default()

	// 配置路由
	internal.SetupRoutes(svc, router)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", viper.GetString("server.host"), viper.GetInt("server.port"))
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
