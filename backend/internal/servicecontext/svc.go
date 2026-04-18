package servicecontext

import (
	"novelflow/backend/pkg/jwt"
	"novelflow/cache"
	"novelflow/database"
	"time"

	"github.com/spf13/viper"
)

type ServiceContext struct {
	JwtUtil     *jwt.JWT
	UserModel   *database.UserRepository
	RedisClient *cache.Client
}

func NewServiceContext() *ServiceContext {
	// 初始化 JWT
	jwtUtil := jwt.NewJWT(
		viper.GetString("jwt.access_secret"),
		viper.GetString("jwt.refresh_secret"),
		time.Duration(viper.GetInt("jwt.access_expire"))*time.Second,
		time.Duration(viper.GetInt("jwt.refresh_expire"))*time.Second,
	)

	// 初始化仓储层
	userRepo := database.NewUserRepository(database.GetDB())

	return &ServiceContext{
		JwtUtil:   jwtUtil,
		UserModel: userRepo,
	}
}
