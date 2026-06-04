package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

type Client struct {
	*redis.Client
}

const (
	JWTBlacklistPrefix    = "jwt:blacklist:"
	UserIDKeyPrefix       = "user:id:"
	UserUsernameKeyPrefix = "user:username:"
	TaskKeyPrefix         = "task:"

	UserCacheTTL = 10 * time.Minute
	TaskCacheTTL = 3 * time.Second
)

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
	addr := fmt.Sprintf("%s:%d", viper.GetString("redis.host"), viper.GetInt("redis.port"))

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
		PoolSize: viper.GetInt("redis.pool_size"),
	})

	prefix := viper.GetString("jwt.black_list_prefix")
	if prefix == "" {
		prefix = DefaultJWTBlacklistConfig.Prefix
	}
	exp := viper.GetInt("jwt.black_list_exp")
	expDuration := DefaultJWTBlacklistConfig.Expiration
	if exp > 0 {
		expDuration = time.Second * time.Duration(exp)
	}
	jwtBlacklistConfig = JWTBlacklistConfig{
		Prefix:     prefix,
		Expiration: expDuration,
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis连接失败: %w", err)
	}

	return &Client{client}, nil
}

// Close 关闭Redis连接
func (r *Client) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}

// Get 通用 JSON 读取，key 不存在时返回 ("", nil)
func (r *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := r.Client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

// Set 通用 JSON 写入
func (r *Client) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.Client.Set(ctx, key, value, ttl).Err()
}

// Del 删除一个或多个 key
func (r *Client) Del(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// GetJSON 读取并反序列化 JSON 值，key 不存在时 dest 不修改，返回 (false, nil)
func (r *Client) GetJSON(ctx context.Context, key string, dest any) (bool, error) {
	val, err := r.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if val == "" {
		return false, nil
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return false, fmt.Errorf("cache unmarshal %q: %w", key, err)
	}
	return true, nil
}

// SetJSON 序列化并写入 JSON 值
func (r *Client) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal %q: %w", key, err)
	}
	return r.Set(ctx, key, string(data), ttl)
}
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
