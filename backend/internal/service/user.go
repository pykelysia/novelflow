package service

import (
	"context"
	"fmt"

	"novelflow/backend/internal/servicecontext"
	"novelflow/cache"
	sqldb "novelflow/database/mysql"
)

// UserService 用户服务
type UserService struct{}

// NewUserService 创建用户服务
func NewUserService() *UserService {
	return &UserService{}
}

// GetUserByID 根据 ID 获取用户（cache-aside），返回不含密码的响应对象
func (s *UserService) GetUserByID(svc *servicecontext.ServiceContext, id uint) (*sqldb.UserResponse, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", cache.UserIDKeyPrefix, id)

	var resp sqldb.UserResponse
	if hit, err := svc.RedisClient.GetJSON(ctx, key, &resp); err == nil && hit {
		return &resp, nil
	}

	u, err := svc.UserModel.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

	r := u.ToResponse()
	_ = svc.RedisClient.SetJSON(ctx, key, r, cache.UserCacheTTL)
	_ = svc.RedisClient.SetJSON(ctx, fmt.Sprintf("%s%s", cache.UserUsernameKeyPrefix, u.Username), r, cache.UserCacheTTL)

	return r, nil
}

// UpdateUser 更新用户，成功后 invalidate 缓存
func (s *UserService) UpdateUser(svc *servicecontext.ServiceContext, id uint, req *sqldb.UpdateUserRequest) (*sqldb.User, error) {
	user, err := svc.UserModel.FindByID(id)
	if err != nil {
		return nil, ErrUserNotFound
	}

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

	ctx := context.Background()
	_ = svc.RedisClient.Del(ctx,
		fmt.Sprintf("%s%d", cache.UserIDKeyPrefix, id),
		fmt.Sprintf("%s%s", cache.UserUsernameKeyPrefix, user.Username),
	)

	return user, nil
}

// DeleteUser 删除用户，成功后 invalidate 缓存
func (s *UserService) DeleteUser(svc *servicecontext.ServiceContext, id uint) error {
	user, err := svc.UserModel.FindByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	if err := svc.UserModel.Delete(id); err != nil {
		return ErrDeleteFailed
	}

	ctx := context.Background()
	_ = svc.RedisClient.Del(ctx,
		fmt.Sprintf("%s%d", cache.UserIDKeyPrefix, id),
		fmt.Sprintf("%s%s", cache.UserUsernameKeyPrefix, user.Username),
	)

	return nil
}
