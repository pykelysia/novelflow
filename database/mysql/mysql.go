package mysql

import (
	"fmt"
	"sync"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	db    *gorm.DB
	dbErr error
	once  sync.Once
)

// NewDB 初始化数据库连接（单例）
func NewDB() (*gorm.DB, error) {
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			viper.GetString("database.username"),
			viper.GetString("database.password"),
			viper.GetString("database.host"),
			viper.GetInt("database.port"),
			viper.GetString("database.dbname"),
		)

		var err error
		DB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			dbErr = fmt.Errorf("failed to connect to database: %w", err)
			return
		}

		// 自动迁移
		if err := DB.AutoMigrate(&User{}, &UserSession{}); err != nil {
			dbErr = fmt.Errorf("failed to migrate database: %w", err)
			return
		}

		db = DB
	})
	if dbErr != nil {
		return nil, dbErr
	}
	return db, nil
}
