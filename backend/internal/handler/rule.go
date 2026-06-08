package handler

import (
	"errors"
	"novelflow/backend/internal/response"
	"novelflow/backend/internal/service"
	"novelflow/backend/internal/servicecontext"
	"novelflow/database/mongodb"

	"github.com/gin-gonic/gin"
)

type RuleHandler struct {
	svc         *servicecontext.ServiceContext
	ruleService *service.RuleService
}

func NewRuleHandler(svc *servicecontext.ServiceContext) *RuleHandler {
	return &RuleHandler{
		svc:         svc,
		ruleService: service.NewRuleService(),
	}
}

func (h *RuleHandler) CreateRule(c *gin.Context) {
	userID := mustUserID(c)
	var req mongodb.CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rule, err := h.ruleService.CreateRule(h.svc, userID, &req)
	if err != nil {
		response.InternalServerError(c, "failed to create rule")
		return
	}
	response.Created(c, rule)
}

func (h *RuleHandler) ListRules(c *gin.Context) {
	userID := mustUserID(c)
	rules, err := h.ruleService.ListRules(h.svc, userID)
	if err != nil {
		response.InternalServerError(c, "failed to list rules")
		return
	}
	response.Success(c, gin.H{"rules": rules})
}

func (h *RuleHandler) GetRule(c *gin.Context) {
	userID := mustUserID(c)
	rule, err := h.ruleService.GetRule(h.svc, userID, c.Param("id"))
	if err != nil {
		handleRuleErr(c, err)
		return
	}
	response.Success(c, rule)
}

func (h *RuleHandler) UpdateRule(c *gin.Context) {
	userID := mustUserID(c)
	var req mongodb.UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	rule, err := h.ruleService.UpdateRule(h.svc, userID, c.Param("id"), &req)
	if err != nil {
		handleRuleErr(c, err)
		return
	}
	response.Success(c, rule)
}

func (h *RuleHandler) ToggleRule(c *gin.Context) {
	userID := mustUserID(c)
	rule, err := h.ruleService.ToggleRule(h.svc, userID, c.Param("id"))
	if err != nil {
		handleRuleErr(c, err)
		return
	}
	response.Success(c, rule)
}

func (h *RuleHandler) DeleteRule(c *gin.Context) {
	userID := mustUserID(c)
	if err := h.ruleService.DeleteRule(h.svc, userID, c.Param("id")); err != nil {
		handleRuleErr(c, err)
		return
	}
	response.SuccessWithMessage(c, "rule deleted successfully", nil)
}

func mustUserID(c *gin.Context) uint {
	v, _ := c.Get("userID")
	id, _ := v.(uint)
	return id
}

func handleRuleErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrRuleNotFound):
		response.NotFound(c, "rule not found")
	case errors.Is(err, service.ErrRuleForbidden):
		response.Forbidden(c, "rule does not belong to you")
	default:
		response.InternalServerError(c, "internal error")
	}
}
