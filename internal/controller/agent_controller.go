package controller

import (
	"strings"
	"time"

	"github.com/MrMiaoMIMI/goshared/util/serverresp"
	"github.com/gin-gonic/gin"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/request"
	"github.com/MrMiaoMIMI/mockserver/internal/model/response"
	"github.com/MrMiaoMIMI/mockserver/internal/service"
)

type AgentController struct {
	scenarios service.ScenarioService
	traffic   service.TrafficService
}

func NewAgentController(scenarioService service.ScenarioService, trafficService service.TrafficService) *AgentController {
	return &AgentController{
		scenarios: scenarioService,
		traffic:   trafficService,
	}
}

func (c *AgentController) CreateScenario(ctx *gin.Context) {
	var req request.CreateScenarioRequest
	if !bindOptionalJSON(ctx, &req) {
		return
	}
	scenario := bo.Scenario{
		ID:          strings.TrimSpace(req.ScenarioID),
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
	if req.TTLSeconds > 0 {
		scenario.ExpireTime = scenarioExpireTimeFromTTLSeconds(req.TTLSeconds)
	}
	created, err := c.scenarios.CreateScenario(ctx.Request.Context(), scenario)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.NewScenarioResponse(created))
}

func (c *AgentController) ListScenarios(ctx *gin.Context) {
	scenarios, err := c.scenarios.ListScenarios(ctx.Request.Context(), bo.ScenarioQuery{
		Status:         strings.TrimSpace(ctx.Query("status")),
		IncludeExpired: boolQuery(ctx, "include_expired", false),
		Limit:          intQuery(ctx, "limit", 100),
		Offset:         intQuery(ctx, "offset", 0),
	})
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.NewListScenariosResponse(scenarios, c.scenarioRuleCounts(ctx, scenarios)))
}

func (c *AgentController) GetScenario(ctx *gin.Context) {
	scenario, err := c.scenarios.GetScenario(ctx.Request.Context(), ctx.Param("scenario_id"))
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	rules, err := c.scenarios.ListScenarioRules(ctx.Request.Context(), scenario.ID)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.NewScenarioResponseWithRuleCount(scenario, len(rules)))
}

func (c *AgentController) UpdateScenario(ctx *gin.Context) {
	var req request.UpdateScenarioRequest
	if !bindOptionalJSON(ctx, &req) {
		return
	}
	update := bo.ScenarioUpdate{
		Name:        req.Name,
		Description: req.Description,
		ExpireTime:  req.ExpireTime,
	}
	if req.TTLSeconds != nil && *req.TTLSeconds > 0 {
		expireTime := scenarioExpireTimeFromTTLSeconds(*req.TTLSeconds)
		update.ExpireTime = &expireTime
	}
	scenario, err := c.scenarios.UpdateScenario(ctx.Request.Context(), ctx.Param("scenario_id"), update)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.NewScenarioResponse(scenario))
}

func (c *AgentController) UpsertHTTPQuickRule(ctx *gin.Context) {
	var req request.HTTPQuickRuleRequest
	if !bindJSON(ctx, &req) {
		return
	}
	rule, err := c.scenarios.UpsertHTTPQuickRule(ctx.Request.Context(), ctx.Param("scenario_id"), bo.HTTPQuickRule{
		RuleID:    req.RuleID,
		Name:      req.Name,
		Namespace: req.Namespace,
		Enabled:   req.Enabled,
		Priority:  req.Priority,
		Match:     req.Match,
		Respond:   req.Respond,
	})
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.NewScenarioRuleResponse(rule))
}

func (c *AgentController) UpsertScenarioRule(ctx *gin.Context) {
	var req request.UpsertScenarioRuleRequest
	if !bindJSON(ctx, &req) {
		return
	}
	ruleID := strings.TrimSpace(ctx.Param("rule_id"))
	req.Rule.ID = ruleID
	rule, err := c.scenarios.UpsertScenarioRule(ctx.Request.Context(), bo.ScenarioRule{
		ScenarioID: ctx.Param("scenario_id"),
		RuleID:     ruleID,
		Name:       req.Rule.Name,
		Protocol:   req.Protocol,
		Enabled:    req.Rule.Enabled,
		Priority:   req.Rule.Priority,
		Rule:       req.Rule,
	})
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.NewScenarioRuleResponse(rule))
}

func (c *AgentController) ListScenarioRules(ctx *gin.Context) {
	rules, err := c.scenarios.ListScenarioRules(ctx.Request.Context(), ctx.Param("scenario_id"))
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.NewListScenarioRulesResponse(rules))
}

func (c *AgentController) DeleteScenarioRule(ctx *gin.Context) {
	if err := c.scenarios.DeleteScenarioRule(ctx.Request.Context(), ctx.Param("scenario_id"), ctx.Param("rule_id")); err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, gin.H{"deleted": true})
}

func (c *AgentController) SimulateScenario(ctx *gin.Context) {
	var req request.SimulateScenarioRequest
	if !bindJSON(ctx, &req) {
		return
	}
	result, err := c.scenarios.SimulateScenario(ctx.Request.Context(), ctx.Param("scenario_id"), req.Event, req.ExplainOnly, req.ExplainMaxDepth, req.ExplainCompact, req.ExplainSummary)
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.SimulateScenarioResponse{Result: result})
}

func (c *AgentController) ListScenarioTraffic(ctx *gin.Context) {
	list, err := c.traffic.ListTrafficEvents(ctx.Request.Context(), bo.TrafficQuery{
		Limit:        intQuery(ctx, "limit", 50),
		Offset:       intQuery(ctx, "offset", 0),
		StartTime:    uint64Query(ctx, "start_time", 0),
		EndTime:      uint64Query(ctx, "end_time", 0),
		ScenarioID:   strings.TrimSpace(ctx.Param("scenario_id")),
		ProtocolName: strings.TrimSpace(ctx.Query("protocol_name")),
		NamespaceID:  strings.TrimSpace(ctx.Query("namespace_id")),
		Outcome:      strings.TrimSpace(ctx.Query("outcome")),
		DecisionKind: strings.TrimSpace(ctx.Query("decision_kind")),
		RuleID:       strings.TrimSpace(ctx.Query("rule_id")),
	})
	if err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, response.NewListTrafficEventsResponse(list))
}

func (c *AgentController) DeleteScenario(ctx *gin.Context) {
	if err := c.scenarios.DeleteScenario(ctx.Request.Context(), ctx.Param("scenario_id")); err != nil {
		writeBusinessError(ctx, err)
		return
	}
	serverresp.Success(ctx, gin.H{"deleted": true})
}

func (c *AgentController) scenarioRuleCounts(ctx *gin.Context, scenarios []bo.Scenario) map[string]int {
	counts := make(map[string]int, len(scenarios))
	for _, scenario := range scenarios {
		rules, err := c.scenarios.ListScenarioRules(ctx.Request.Context(), scenario.ID)
		if err != nil {
			continue
		}
		counts[scenario.ID] = len(rules)
	}
	return counts
}

func scenarioExpireTimeFromTTLSeconds(ttlSeconds uint64) uint64 {
	return uint64(time.Now().UTC().UnixMilli()) + ttlSeconds*1000
}
