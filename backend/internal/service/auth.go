package service

import (
	"context"
	"errors"
	"fmt"

	"novelflow/backend/internal/servicecontext"
	"novelflow/cache"
	sqldb "novelflow/database/mysql"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// TokenResponse 令牌响应
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshRequest 刷新请求
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// LogoutRequest 登出请求
type LogoutRequest struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"omitempty,email"`
	Nickname string `json:"nickname" binding:"omitempty,max=100"`
}

// Login 用户登录
func (l *AuthService) Login(svc *servicecontext.ServiceContext, req *LoginRequest) (*TokenResponse, error) {
	user, err := svc.UserModel.FindByUsername(req.Username)
	if err != nil {
		return nil, ErrInvalidCredential
	}

	if !svc.UserModel.VerifyPassword(user, req.Password) {
		return nil, ErrInvalidCredential
	}

	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	accessToken, err := svc.JwtUtil.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := svc.JwtUtil.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	warmUserCache(svc, user)

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Register 用户注册
func (l *AuthService) Register(svc *servicecontext.ServiceContext, req *RegisterRequest) (*sqldb.User, error) {
	exists, err := svc.UserModel.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	if req.Email != "" {
		emailExists, err := svc.UserModel.ExistsByEmail(req.Email)
		if err != nil {
			return nil, err
		}
		if emailExists {
			return nil, errors.New("email already exists")
		}
	}

	user := &sqldb.User{
		Username: req.Username,
		Email:    req.Email,
		Nickname: req.Nickname,
		Status:   1,
	}

	if err := svc.UserModel.HashPassword(user, req.Password); err != nil {
		return nil, err
	}

	if err := svc.UserModel.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// RefreshToken 刷新令牌
func (l *AuthService) RefreshToken(svc *servicecontext.ServiceContext, req *RefreshRequest) (*TokenResponse, error) {
	claims, err := svc.JwtUtil.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	user, err := svc.UserModel.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	accessToken, err := svc.JwtUtil.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := svc.JwtUtil.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout 用户登出
func (l *AuthService) Logout(svc *servicecontext.ServiceContext, req *LogoutRequest) error {
	ctx := context.Background()

	_ = svc.RedisClient.AddJWTToBlacklist(ctx, req.AccessToken)
	_ = svc.RedisClient.AddJWTToBlacklist(ctx, req.RefreshToken)

	return nil
}

func warmUserCache(svc *servicecontext.ServiceContext, user *sqldb.User) {
	ctx := context.Background()
	_ = svc.RedisClient.SetJSON(ctx, fmt.Sprintf("%s%d", cache.UserIDKeyPrefix, user.ID), user, cache.UserCacheTTL)
	_ = svc.RedisClient.SetJSON(ctx, fmt.Sprintf("%s%s", cache.UserUsernameKeyPrefix, user.Username), user, cache.UserCacheTTL)
}
