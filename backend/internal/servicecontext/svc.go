package servicecontext

import (
	"context"
	"novelflow/backend/pkg/jwt"
	"novelflow/backend/pkg/logger"
	"novelflow/cache"
	"novelflow/database/mongodb"
	sqldb "novelflow/database/mysql"
	"sync"
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
	RuleRepo         *mongodb.RuleRepository
	WG               sync.WaitGroup
	Ctx              context.Context
	Cancel           context.CancelFunc
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
		logger.Fatal("failed to init database", "err", err)
	}
	svc.db = db
	svc.UserModel = sqldb.NewUserRepository(db)
	svc.UserSessionModel = sqldb.NewUserSessionRepository(db)

	// 初始化 Redis
	redisClient, err := cache.InitRedis()
	if err != nil {
		logger.Fatal("failed to init redis", "err", err)
	}
	svc.RedisClient = redisClient

	mdb, err := mongodb.NewMongoDB()
	if err != nil {
		logger.Fatal("failed to init mongodb", "err", err)
	}
	svc.MongoDB = mdb
	svc.RuleRepo = mongodb.NewRuleRepository(mdb)

	svc.Ctx, svc.Cancel = context.WithCancel(context.Background())

	return svc
}

func (svc *ServiceContext) Close() {
	svc.RedisClient.Close()
	svc.MongoDB.Close()
	if sqlDB, err := svc.db.DB(); err == nil {
		sqlDB.Close()
	}
}
