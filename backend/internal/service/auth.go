package service

import (
	"errors"

	"novelflow/backend/pkg/jwt"
	"novelflow/database"
)

// AuthService 认证服务
type AuthService struct {
	userRepo database.UserRepository
	jwtUtil  *jwt.JWT
}

// NewAuthService 创建认证服务
func NewAuthService(userRepo database.UserRepository, jwtUtil *jwt.JWT) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtUtil:  jwtUtil,
	}
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
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RefreshRequest 刷新请求
type RefreshRequest struct {
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
func (s *AuthService) Login(req *LoginRequest) (*TokenResponse, error) {
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, ErrInvalidCredential
	}

	// 验证密码
	if !s.userRepo.VerifyPassword(user, req.Password) {
		return nil, ErrInvalidCredential
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	// 生成令牌
	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}, nil
}

// Register 用户注册
func (s *AuthService) Register(req *RegisterRequest) (*database.User, error) {
	exists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// 检查邮箱是否已存在
	if req.Email != "" {
		emailExists, err := s.userRepo.ExistsByEmail(req.Email)
		if err != nil {
			return nil, err
		}
		if emailExists {
			return nil, errors.New("email already exists")
		}
	}

	// 创建用户
	user := &database.User{
		Username: req.Username,
		Email:    req.Email,
		Nickname: req.Nickname,
		Status:   1,
	}

	// 密码加密
	if err := s.userRepo.HashPassword(user, req.Password); err != nil {
		return nil, err
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// RefreshToken 刷新令牌
func (s *AuthService) RefreshToken(req *RefreshRequest) (*TokenResponse, error) {
	// 验证刷新令牌
	claims, err := s.jwtUtil.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 检查用户状态
	if user.Status != 1 {
		return nil, errors.New("user account is disabled")
	}

	// 生成新令牌
	accessToken, err := s.jwtUtil.GenerateAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtUtil.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600,
		TokenType:    "Bearer",
	}, nil
}

// Logout 用户登出
func (s *AuthService) Logout(token string) error {
	// TODO: 实现 Redis 将令牌加入黑名单
	return nil
}
