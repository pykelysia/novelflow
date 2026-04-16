package internal

import (
	"novelflow/backend/internal/handler"
	"novelflow/backend/internal/middleware"
	"novelflow/backend/internal/repository"
	"novelflow/backend/internal/service"
	"novelflow/backend/pkg/jwt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

// SetupRoutes 配置所有路由
func SetupRoutes(router *gin.Engine) {
	// 初始化 JWT
	jwtUtil := jwt.NewJWT(
		viper.GetString("jwt.access_secret"),
		viper.GetString("jwt.refresh_secret"),
		time.Duration(viper.GetInt("jwt.access_expire"))*time.Second,
		time.Duration(viper.GetInt("jwt.refresh_expire"))*time.Second,
	)

	// 初始化仓储层
	userRepo := repository.NewUserRepository()

	// 初始化服务层
	authService := service.NewAuthService(userRepo, jwtUtil)
	userService := service.NewUserService(userRepo)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)

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
	authGroup.Use(middleware.AuthMiddleware(jwtUtil))
	{
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
}
