package response

import "github.com/MrMiaoMIMI/mockserver/internal/model/bo"

type ScenarioResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	ExpireTime  uint64 `json:"expire_time"`
	RuleCount   int    `json:"rule_count,omitempty"`
	Version     int    `json:"version"`
}

type ListScenariosResponse struct {
	Items []ScenarioResponse `json:"items"`
	Total int                `json:"total"`
}

type ScenarioRuleResponse struct {
	ScenarioID string  `json:"scenario_id"`
	RuleID     string  `json:"rule_id"`
	Name       string  `json:"name,omitempty"`
	Protocol   string  `json:"protocol"`
	Enabled    bool    `json:"enabled"`
	Priority   int     `json:"priority"`
	ExpireTime uint64  `json:"expire_time"`
	Version    int     `json:"version"`
	Rule       bo.Rule `json:"rule"`
}

type ListScenarioRulesResponse struct {
	Items []ScenarioRuleResponse `json:"items"`
	Total int                    `json:"total"`
}

type SimulateScenarioResponse struct {
	Result bo.SimulationResult `json:"result"`
}

func NewScenarioResponse(item bo.Scenario) ScenarioResponse {
	return NewScenarioResponseWithRuleCount(item, 0)
}

func NewScenarioResponseWithRuleCount(item bo.Scenario, ruleCount int) ScenarioResponse {
	return ScenarioResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Status:      item.Status,
		ExpireTime:  item.ExpireTime,
		RuleCount:   ruleCount,
		Version:     item.Version,
	}
}

func NewListScenariosResponse(items []bo.Scenario, ruleCounts map[string]int) ListScenariosResponse {
	responses := make([]ScenarioResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, NewScenarioResponseWithRuleCount(item, ruleCounts[item.ID]))
	}
	return ListScenariosResponse{
		Items: responses,
		Total: len(responses),
	}
}

func NewScenarioRuleResponse(item bo.ScenarioRule) ScenarioRuleResponse {
	return ScenarioRuleResponse{
		ScenarioID: item.ScenarioID,
		RuleID:     item.RuleID,
		Name:       item.Name,
		Protocol:   item.Protocol,
		Enabled:    item.Enabled,
		Priority:   item.Priority,
		ExpireTime: item.ExpireTime,
		Version:    item.Version,
		Rule:       item.Rule,
	}
}

func NewListScenarioRulesResponse(items []bo.ScenarioRule) ListScenarioRulesResponse {
	responses := make([]ScenarioRuleResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, NewScenarioRuleResponse(item))
	}
	return ListScenarioRulesResponse{
		Items: responses,
		Total: len(responses),
	}
}
