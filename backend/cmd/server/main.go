package main

import (
	"log"

	"novelflow/backend/config"
	"novelflow/backend/internal/handler"
	"novelflow/backend/internal/middleware"
	"novelflow/backend/internal/repository"
	"novelflow/backend/internal/service"
	"novelflow/backend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化 JWT
	jwtUtil := jwt.NewJWT(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.GetAccessExpireDuration(),
		cfg.GetRefreshExpireDuration(),
	)

	// 初始化仓储层
	userRepo := repository.NewUserRepository()

	// 初始化服务层
	authService := service.NewAuthService(userRepo, jwtUtil)
	userService := service.NewUserService(userRepo)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

	// 创建 Gin 路由器
	router := gin.Default()

	// 使用跨域中间件
	router.Use(middleware.CorsMiddleware())

	// 认证路由（无需登录）
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
	}

	// 用户路由（需要登录）
	userGroup := router.Group("/users")
	userGroup.Use(middleware.AuthMiddleware(jwtUtil))
	{
		userGroup.GET("", userHandler.GetUsers)
		userGroup.GET("/:id", userHandler.GetUser)
		userGroup.PUT("/:id", userHandler.UpdateUser)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// 启动服务器
	addr := cfg.GetServerAddr()
	log.Printf("Server starting on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
