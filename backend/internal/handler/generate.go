package handler

import (
	"errors"
	"net/http"

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
		generateService: service.NewGenerateService(),
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
		response.InternalServerError(c, "start generation failed")
		return
	}

	c.JSON(http.StatusOK, response.Response{
		Code:    http.StatusOK,
		Message: "generation task created",
		Data:    resp,
	})
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
