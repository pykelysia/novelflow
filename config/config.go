package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// LoadConfig 加载配置文件，使用 BindEnv 实现环境变量读取
func LoadConfig(path string) error {
	// 设置配置文件
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	// 读取配置文件作为默认值
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 设置环境变量前缀
	viper.SetEnvPrefix("NOVELFLOW")

	// 绑定环境变量
	bindEnvVariables()

	return nil
}

// bindEnvVariables 绑定所有环境变量到配置路径
func bindEnvVariables() {
	// Server 配置
	viper.BindEnv("server.host", "NOVELFLOW_SERVER_HOST")
	viper.BindEnv("server.port", "NOVELFLOW_SERVER_PORT")

	// Database 配置
	viper.BindEnv("database.host", "NOVELFLOW_DATABASE_HOST")
	viper.BindEnv("database.port", "NOVELFLOW_DATABASE_PORT")
	viper.BindEnv("database.username", "NOVELFLOW_DATABASE_USERNAME")
	viper.BindEnv("database.password", "NOVELFLOW_DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "NOVELFLOW_DATABASE_DBNAME")

	// Redis 配置
	viper.BindEnv("redis.host", "NOVELFLOW_REDIS_HOST")
	viper.BindEnv("redis.port", "NOVELFLOW_REDIS_PORT")
	viper.BindEnv("redis.password", "NOVELFLOW_REDIS_PASSWORD")
	viper.BindEnv("redis.db", "NOVELFLOW_REDIS_DB")

	// JWT 配置
	viper.BindEnv("jwt.access_secret", "NOVELFLOW_JWT_ACCESS_SECRET")
	viper.BindEnv("jwt.refresh_secret", "NOVELFLOW_JWT_REFRESH_SECRET")
	viper.BindEnv("jwt.access_expire", "NOVELFLOW_JWT_ACCESS_EXPIRE")
	viper.BindEnv("jwt.refresh_expire", "NOVELFLOW_JWT_REFRESH_EXPIRE")

	// Model 配置
	viper.BindEnv("llm.model_type")
	viper.BindEnv("llm.model_name")
	viper.BindEnv("llm.base_url")
	viper.BindEnv("llm.api_key")
	viper.BindEnv("llm.max_tokens")

	// Lite LLM 配置
	viper.BindEnv("lite_llm.model_type")
	viper.BindEnv("lite_llm.model_name")
	viper.BindEnv("lite_llm.base_url")
	viper.BindEnv("lite_llm.api_key")
	viper.BindEnv("lite_llm.max_tokens")

	// Mongo 配置
	viper.BindEnv("mongo", "NOVELFLOW_MONGO")

	// Skills 配置
	viper.BindEnv("skills.base_dir")

	// Storage 配置
	viper.BindEnv("storage.novels_dir", "NOVELFLOW_STORAGE_NOVELS_DIR")

	// Log 配置
	viper.BindEnv("log.output_dir", "NOVELFLOW_LOG_OUTPUT_DIR")
	viper.BindEnv("log.level", "NOVELFLOW_LOG_LEVEL")
	viper.BindEnv("log.console", "NOVELFLOW_LOG_CONSOLE")

	// Context compress 配置
	viper.BindEnv("context_compress.threshold_chars", "NOVELFLOW_CONTEXT_COMPRESS_THRESHOLD_CHARS")
	viper.BindEnv("context_compress.keep_recent_msgs", "NOVELFLOW_CONTEXT_COMPRESS_KEEP_RECENT_MSGS")
	viper.BindEnv("context_compress.keep_recent_tools", "NOVELFLOW_CONTEXT_COMPRESS_KEEP_RECENT_TOOLS")
}
