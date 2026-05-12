package servicecontext

import (
	"log"
	"novelflow/backend/pkg/jwt"
	"novelflow/cache"
	"novelflow/database/mongodb"
	sqldb "novelflow/database/mysql"
	"time"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

type ServiceContext struct {
	JwtUtil          *jwt.JWT
	db               *gorm.DB
	UserModel        *sqldb.UserRepository
	UserSessionModel *sqldb.UserSessionRepository
	RedisClient      *cache.Client
	MongoDB          *mongodb.MongoClient
}

func NewServiceContext() *ServiceContext {
	svc := &ServiceContext{}
	// 初始化 JWT
	jwtUtil := jwt.NewJWT(
		viper.GetString("jwt.access_secret"),
		viper.GetString("jwt.refresh_secret"),
		time.Duration(viper.GetInt("jwt.access_expire"))*time.Second,
		time.Duration(viper.GetInt("jwt.refresh_expire"))*time.Second,
	)
	svc.JwtUtil = jwtUtil

	db, err := sqldb.NewDB()
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	svc.db = db
	svc.UserModel = sqldb.NewUserRepository(db)
	svc.UserSessionModel = sqldb.NewUserSessionRepository(db)

	// 初始化 Redis
	redisClient, err := cache.InitRedis()
	if err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}
	svc.RedisClient = redisClient

	mdb, err := mongodb.NewMongoDB()
	if err != nil {
		log.Fatalf("Failed to init mongodb: %v", err)
	}
	svc.MongoDB = mdb

	return svc
}

func (svc *ServiceContext) Close() {
	svc.RedisClient.Close()
	svc.MongoDB.Close()
}
