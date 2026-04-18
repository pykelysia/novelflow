package handler

import (
	"errors"
	"strconv"

	"novelflow/backend/internal/response"
	"novelflow/backend/internal/service"
	"novelflow/backend/internal/servicecontext"
	"novelflow/database"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	svc         *servicecontext.ServiceContext
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(svc *servicecontext.ServiceContext) *UserHandler {
	return &UserHandler{
		svc:         svc,
		userService: service.NewUserService(),
	}
}

// GetUser 获取单个用户
// @Summary 获取单个用户
// @Description 根据 ID 获取用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=database.UserResponse}
// @Failure 404 {object} response.Response
// @Router /users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	user, err := h.userService.GetUserByID(h.svc, uint(id))
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalServerError(c, "failed to get user")
		return
	}

	response.Success(c, user.ToResponse())
}

// UpdateUser 更新用户
// @Summary 更新用户
// @Description 更新用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Param request body database.UpdateUserRequest true "更新信息"
// @Success 200 {object} response.Response{data=database.UserResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req database.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.userService.UpdateUser(h.svc, uint(id), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalServerError(c, "failed to update user")
		return
	}

	response.Success(c, user.ToResponse())
}

// DeleteUser 删除用户
// @Summary 删除用户
// @Description 删除用户
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.userService.DeleteUser(h.svc, uint(id)); err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			response.NotFound(c, "user not found")
			return
		}
		response.InternalServerError(c, "failed to delete user")
		return
	}

	response.SuccessWithMessage(c, "user deleted successfully", nil)
}
