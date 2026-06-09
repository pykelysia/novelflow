package handler

import (
	"errors"

	"novelflow/backend/internal/response"
	"novelflow/backend/internal/service"
	"novelflow/backend/internal/servicecontext"

	"github.com/gin-gonic/gin"
)

type GenerateHandler struct {
	svc             *servicecontext.ServiceContext
	generateService *service.GenerateService
}

func NewGenerateHandler(svc *servicecontext.ServiceContext) *GenerateHandler {
	return &GenerateHandler{
		svc:             svc,
		generateService: service.NewGenerateService(svc.Ctx),
	}
}

func (h *GenerateHandler) StartGeneration(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, ok := userID.(uint)
	if !ok {
		response.Unauthorized(c, "invalid user identity")
		return
	}

	var req service.GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := h.generateService.StartGeneration(h.svc, uid, &req)
	if err != nil {
		if errors.Is(err, service.ErrTooManyRequests) {
			response.TooManyRequests(c, "too many concurrent generations, please try again later")
			return
		}
		response.InternalServerError(c, "start generation failed")
		return
	}

	response.Created(c, resp)
}

func (h *GenerateHandler) GetGenerationStatus(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, ok := userID.(uint)
	if !ok {
		response.Unauthorized(c, "invalid user identity")
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.BadRequest(c, "missing session_id")
		return
	}

	resp, err := h.generateService.GetGenerationStatus(h.svc, uid, sessionID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			response.NotFound(c, "task not found")
			return
		}
		if errors.Is(err, service.ErrTaskForbidden) {
			response.Forbidden(c, "task does not belong to user")
			return
		}
		response.InternalServerError(c, "get generation status failed")
		return
	}

	response.Success(c, resp)
}

func (h *GenerateHandler) GetGenerationResult(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, ok := userID.(uint)
	if !ok {
		response.Unauthorized(c, "invalid user identity")
		return
	}

	sessionID := c.Param("session_id")
	if sessionID == "" {
		response.BadRequest(c, "missing session_id")
		return
	}

	resp, err := h.generateService.GetGenerationResult(h.svc, uid, sessionID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			response.NotFound(c, "task not found")
			return
		}
		if errors.Is(err, service.ErrTaskForbidden) {
			response.Forbidden(c, "task does not belong to user")
			return
		}
		response.InternalServerError(c, "get generation result failed")
		return
	}

	response.Success(c, resp)
}

func (h *GenerateHandler) ListTasks(c *gin.Context) {
	userID, _ := c.Get("userID")
	uid, ok := userID.(uint)
	if !ok {
		response.Unauthorized(c, "invalid user identity")
		return
	}

	resp, err := h.generateService.ListUserTasks(h.svc, uid)
	if err != nil {
		response.InternalServerError(c, "list tasks failed")
		return
	}

	response.Success(c, resp)
}
