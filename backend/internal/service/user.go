package service

import (
	"novelflow/backend/internal/servicecontext"
	sqldb "novelflow/database/mysql"
)

// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{}
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(svc *servicecontext.ServiceContext, id uint) (*sqldb.User, error) {
	user, err := svc.UserModel.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(svc *servicecontext.ServiceContext, id uint, req *sqldb.UpdateUserRequest) (*sqldb.User, error) {
	user, err := svc.UserModel.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 更新字段
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	if req.Status != nil {
		user.Status = *req.Status
	}

	if err := svc.UserModel.Update(user); err != nil {
		return nil, ErrUpdateFailed
	}

	return user, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(svc *servicecontext.ServiceContext, id uint) error {
	if err := svc.UserModel.Delete(id); err != nil {
		return ErrDeleteFailed
	}
	return nil
}
