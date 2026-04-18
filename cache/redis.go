package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type Client struct {
	*redis.Client
}

const JWTBlacklistPrefix = "jwt:blacklist:"

// JWTBlacklistConfig JWT黑名单配置
type JWTBlacklistConfig struct {
	Prefix     string
	Expiration time.Duration
}

var DefaultJWTBlacklistConfig = JWTBlacklistConfig{
	Prefix:     JWTBlacklistPrefix,
	Expiration: 24 * time.Hour, // 默认24小时
}

var jwtBlacklistConfig JWTBlacklistConfig

// RedisConfig Redis配置
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// InitRedis 初始化Redis连接
func InitRedis() (*Client, error) {
	var redisClient *redis.Client
	addr := fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port"))

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
		PoolSize: viper.GetInt("redis.pool_size"),
	})

	jwtBlacklistConfig = JWTBlacklistConfig{
		Prefix:     viper.GetString("jwt.black_list_prefix"),
		Expiration: time.Second * time.Duration(viper.GetInt("jwt.balck_list_exp")),
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis连接失败: %w", err)
	}

	return &Client{
		redisClient,
	}, nil
}

// Close 关闭Redis连接
func (r *Client) Close() error {
	if r.Client != nil {
		return r.Close()
	}
	return nil
}

// AddJWTToBlacklist 将JWT添加到黑名单
func (r *Client) AddJWTToBlacklist(ctx context.Context, token string) error {
	if r.Client == nil {
		return fmt.Errorf("redis客户端未初始化")
	}
	key := jwtBlacklistConfig.Prefix + token
	return r.Client.Set(ctx, key, "1", jwtBlacklistConfig.Expiration).Err()
}

// IsJWTInBlacklist 检查JWT是否在黑名单中
func (r *Client) IsJWTInBlacklist(ctx context.Context, token string) (bool, error) {
	if r.Client == nil {
		return false, fmt.Errorf("redis客户端未初始化")
	}
	key := jwtBlacklistConfig.Prefix + token
	exists, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}
