package mysql

import "gorm.io/gorm"

// UserSessionRepository 用户-会话关联仓库
type UserSessionRepository struct {
	db *gorm.DB
}

// NewUserSessionRepository 创建仓库实例
func NewUserSessionRepository(db *gorm.DB) *UserSessionRepository {
	return &UserSessionRepository{db: db}
}

// Create 创建会话关联
func (r *UserSessionRepository) Create(mapping *UserSession) error {
	return r.db.Create(mapping).Error
}

// FindByUserID 根据用户ID查询所有会话
func (r *UserSessionRepository) FindByUserID(userID uint) ([]UserSession, error) {
	var sessions []UserSession
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

// FindBySessionID 根据会话ID查询所属用户
func (r *UserSessionRepository) FindBySessionID(sessionID string) (*UserSession, error) {
	var session UserSession
	err := r.db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}
