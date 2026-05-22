package internal

import (
	"novelflow/backend/internal/handler"
	"novelflow/backend/internal/middleware"
	"novelflow/backend/internal/servicecontext"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 配置所有路由
func SetupRoutes(svc *servicecontext.ServiceContext, router *gin.Engine) {

	// 初始化处理器
	authHandler := handler.NewAuthHandler(svc)
	userHandler := handler.NewUserHandler(svc)

	// 使用跨域中间件
	router.Use(middleware.CorsMiddleware())

	// 认证路由（无需登录）
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/refresh", authHandler.Refresh)
	}

	// 认证路由（需要登录）
	authGroup = router.Group("/auth")
	authGroup.Use(middleware.AuthMiddleware(svc))
	{
		authGroup.POST("/logout", authHandler.Logout)
	}

	// 用户路由（需要登录）
	userGroup := router.Group("/users")
	userGroup.Use(middleware.AuthMiddleware(svc))
	{
		userGroup.GET("/:id", userHandler.GetUser)
		userGroup.PUT("/:id", userHandler.UpdateUser)
		userGroup.DELETE("/:id", userHandler.DeleteUser)
	}

	// 生成路由（需要登录）
	generateHandler := handler.NewGenerateHandler(svc)
	generateGroup := router.Group("/generate")
	generateGroup.Use(middleware.AuthMiddleware(svc))
	{
		generateGroup.POST("", generateHandler.StartGeneration)
		generateGroup.GET("/tasks", generateHandler.ListTasks)
		generateGroup.GET("/:session_id", generateHandler.GetGenerationStatus)
		generateGroup.GET("/:session_id/result", generateHandler.GetGenerationResult)
	}
}
