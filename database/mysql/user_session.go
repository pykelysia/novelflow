package mysql

import "time"

// UserSession 用户-会话关联模型
type UserSession struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index:idx_user_id;not null"`
	SessionID string    `gorm:"size:36;index:idx_session_id;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName 指定表名
func (UserSession) TableName() string {
	return "user_sessions"
}
