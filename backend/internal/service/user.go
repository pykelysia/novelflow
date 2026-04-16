package service

import (
	"novelflow/database"
)

// UserService 用户服务
type UserService struct {
	userRepo database.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo database.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(id uint) (*database.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers() ([]*database.User, error) {
	return s.userRepo.FindAll()
}

// UpdateUser 更新用户
func (s *UserService) UpdateUser(id uint, req *database.UpdateUserRequest) (*database.User, error) {
	user, err := s.userRepo.FindByID(id)
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

	if err := s.userRepo.Update(user); err != nil {
		return nil, ErrUpdateFailed
	}

	return user, nil
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	if err := s.userRepo.Delete(id); err != nil {
		return ErrDeleteFailed
	}
	return nil
}
