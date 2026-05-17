package controller

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/MrMiaoMIMI/goshared/db/dbspi"
	"github.com/MrMiaoMIMI/goshared/util/servererr"
	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"

	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/request"
	"github.com/MrMiaoMIMI/mockserver/internal/model/response"
	"github.com/MrMiaoMIMI/mockserver/internal/view"
	"github.com/MrMiaoMIMI/mockserver/mockprotocol"
)

type AdminController struct {
	ruleSetView   view.RuleSetView
	namespaceView view.NamespaceView
}

func NewAdminController(ruleSetView view.RuleSetView, namespaceView view.NamespaceView) *AdminController {
	return &AdminController{
		ruleSetView:   ruleSetView,
		namespaceView: namespaceView,
	}
}

func (c *AdminController) CreateOrUpdateDraft(ctx *gin.Context) {
	var req request.UpsertRuleSetRequest
	if !bindJSON(ctx, &req) {
		return
	}
	result, err := c.ruleSetView.UpsertDraft(ctx.Request.Context(), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) ListProtocols(ctx *gin.Context) {
	items := mockprotocol.RegisteredSpecs()
	serverresp.Success(ctx, response.ListProtocolsResponse{
		Items: items,
		Total: len(items),
	})
}

func (c *AdminController) ListDrafts(ctx *gin.Context) {
	result, err := c.ruleSetView.ListDrafts(ctx.Request.Context())
	if err != nil {
		serverresp.InternalServerError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) GetDraft(ctx *gin.Context) {
	result, err := c.ruleSetView.GetDraft(ctx.Request.Context(), ctx.Param("ruleset_id"))
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) AddDraftRule(ctx *gin.Context) {
	var req request.DraftRuleRequest
	if !bindJSON(ctx, &req) {
		return
	}
	result, err := c.ruleSetView.AddDraftRule(ctx.Request.Context(), ctx.Param("ruleset_id"), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) UpdateDraftRule(ctx *gin.Context) {
	var req request.DraftRuleRequest
	if !bindJSON(ctx, &req) {
		return
	}
	result, err := c.ruleSetView.UpdateDraftRule(ctx.Request.Context(), ctx.Param("ruleset_id"), ctx.Param("rule_id"), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) DeleteDraftRule(ctx *gin.Context) {
	result, err := c.ruleSetView.DeleteDraftRule(ctx.Request.Context(), ctx.Param("ruleset_id"), ctx.Param("rule_id"))
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) EnableDraftRule(ctx *gin.Context) {
	c.setDraftRuleEnabled(ctx, true)
}

func (c *AdminController) DisableDraftRule(ctx *gin.Context) {
	c.setDraftRuleEnabled(ctx, false)
}

func (c *AdminController) setDraftRuleEnabled(ctx *gin.Context, enabled bool) {
	result, err := c.ruleSetView.SetDraftRuleEnabled(ctx.Request.Context(), ctx.Param("ruleset_id"), ctx.Param("rule_id"), enabled)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) SetDraftRulePriority(ctx *gin.Context) {
	var req request.UpdateRulePriorityRequest
	if !bindJSON(ctx, &req) {
		return
	}
	result, err := c.ruleSetView.SetDraftRulePriority(ctx.Request.Context(), ctx.Param("ruleset_id"), ctx.Param("rule_id"), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) ValidateDraft(ctx *gin.Context) {
	result, err := c.ruleSetView.ValidateDraft(ctx.Request.Context(), ctx.Param("ruleset_id"))
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) PublishDraft(ctx *gin.Context) {
	var req request.PublishRuleSetRequest
	if !bindOptionalJSON(ctx, &req) {
		return
	}
	req.Audit = auditFromGinContext(ctx)
	result, err := c.ruleSetView.Publish(ctx.Request.Context(), ctx.Param("ruleset_id"), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) SimulateDraft(ctx *gin.Context) {
	var req request.SimulateRuleSetRequest
	if !bindJSON(ctx, &req) {
		return
	}
	result, err := c.ruleSetView.SimulateDraft(ctx.Request.Context(), ctx.Param("ruleset_id"), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) ListPublished(ctx *gin.Context) {
	result, err := c.ruleSetView.ListPublished(ctx.Request.Context())
	if err != nil {
		serverresp.InternalServerError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) GetPublished(ctx *gin.Context) {
	result, err := c.ruleSetView.GetPublished(ctx.Request.Context(), ctx.Param("ruleset_id"))
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) ListPublishedSnapshots(ctx *gin.Context) {
	result, err := c.ruleSetView.ListPublishedSnapshots(ctx.Request.Context(), ctx.Param("ruleset_id"))
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) SimulatePublished(ctx *gin.Context) {
	var req request.SimulateRuleSetRequest
	if !bindJSON(ctx, &req) {
		return
	}
	result, err := c.ruleSetView.SimulatePublished(ctx.Request.Context(), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) RollbackPreview(ctx *gin.Context) {
	var req request.RollbackPreviewRuleSetRequest
	if !bindJSON(ctx, &req) {
		return
	}
	result, err := c.ruleSetView.RollbackPreview(ctx.Request.Context(), ctx.Param("ruleset_id"), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) Rollback(ctx *gin.Context) {
	var req request.RollbackRuleSetRequest
	if !bindOptionalJSON(ctx, &req) {
		return
	}
	req.Audit = auditFromGinContext(ctx)
	result, err := c.ruleSetView.Rollback(ctx.Request.Context(), ctx.Param("ruleset_id"), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) CreateNamespace(ctx *gin.Context) {
	var req request.UpsertNamespaceRequest
	if !bindJSON(ctx, &req) {
		return
	}
	result, err := c.namespaceView.UpsertNamespace(ctx.Request.Context(), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) ListNamespaces(ctx *gin.Context) {
	result, err := c.namespaceView.ListNamespaces(ctx.Request.Context())
	if err != nil {
		serverresp.InternalServerError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) GetNamespace(ctx *gin.Context) {
	result, err := c.namespaceView.GetNamespace(ctx.Request.Context(), ctx.Param("namespace_id"))
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func (c *AdminController) UpdateNamespace(ctx *gin.Context) {
	namespaceID := ctx.Param("namespace_id")
	var req request.UpsertNamespaceRequest
	if !bindJSON(ctx, &req) {
		return
	}
	if _, err := c.namespaceView.GetNamespace(ctx.Request.Context(), namespaceID); err != nil {
		writeBusinessError(ctx, err)
		return
	}
	req.ID = namespaceID
	result, err := c.namespaceView.UpsertNamespace(ctx.Request.Context(), req)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, result)
}

func bindJSON(ctx *gin.Context, target any) bool {
	if err := ctx.ShouldBindJSON(target); err != nil {
		serverresp.BadRequestError(ctx, fmt.Errorf("invalid json body: %w", err))
		return false
	}
	return true
}

func bindOptionalJSON(ctx *gin.Context, target any) bool {
	if ctx.Request.Body == nil || ctx.Request.ContentLength == 0 {
		return true
	}
	if err := ctx.ShouldBindJSON(target); err != nil && !errors.Is(err, io.EOF) {
		serverresp.BadRequestError(ctx, fmt.Errorf("invalid json body: %w", err))
		return false
	}
	return true
}

func writeBusinessError(ctx *gin.Context, err error) {
	var bizErr *servererr.BizError
	if errors.As(err, &bizErr) {
		serverresp.Error(ctx, bizErr)
		return
	}
	if errors.Is(err, dao.ErrConflict) {
		ctx.JSON(http.StatusConflict, serverresp.Response{
			Code:    servererr.ErrConflict,
			Message: err.Error(),
		})
		return
	}
	if errors.Is(err, errNotFound) || strings.Contains(strings.ToLower(err.Error()), "not found") {
		serverresp.NotFoundError(ctx, err)
		return
	}
	serverresp.InternalServerError(ctx, err)
}

var errNotFound = errors.New("not found")

func auditFromGinContext(ctx *gin.Context) bo.AuditInfo {
	operator, _ := dbspi.OperatorFromContext(ctx.Request.Context())
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = strings.TrimSpace(ctx.GetHeader("X-Mockserver-Operator"))
	}
	if operator == "" {
		operator = strings.TrimSpace(ctx.GetHeader("X-Operator"))
	}
	return bo.AuditInfo{
		Operator: operator,
		TraceID:  strings.TrimSpace(ctx.GetHeader("X-Trace-ID")),
	}
}
