package handler

import (
	"errors"

	"novelflow/backend/internal/response"
	"novelflow/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账号
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.RegisterRequest true "注册信息"
// @Success 201 {object} response.Response{data=model.User}
// @Failure 400 {object} response.Response
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalServerError(c, "registration failed")
		return
	}

	response.Created(c, user.ToResponse())
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=service.TokenResponse}
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokenResponse, err := h.authService.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredential) {
			response.Unauthorized(c, "invalid username or password")
			return
		}
		response.InternalServerError(c, "login failed")
		return
	}

	response.Success(c, tokenResponse)
}

// Refresh 刷新令牌
// @Summary 刷新令牌
// @Description 使用刷新令牌获取新的访问令牌
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.RefreshRequest true "刷新令牌"
// @Success 200 {object} response.Response{data=service.TokenResponse}
// @Failure 401 {object} response.Response
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokenResponse, err := h.authService.RefreshToken(&req)
	if err != nil {
		response.Unauthorized(c, "invalid refresh token")
		return
	}

	response.Success(c, tokenResponse)
}

// Logout 用户登出
// @Summary 用户登出
// @Description 用户登出，使令牌失效
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body service.LogoutRequest true "登出信息"
// @Success 200 {object} response.Response
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req service.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	_ = h.authService.Logout(&req)

	response.SuccessWithMessage(c, "logged out successfully", nil)
}
