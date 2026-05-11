package dao

import (
	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

func testRuleSet(id, ruleID, path string, priority int) bo.RuleSet {
	return bo.RuleSet{
		ID:        id,
		Name:      id,
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector: bo.Selector{
			All: []bo.Condition{{Field: "request.path", Op: eo.OperatorPrefix, Value: path}},
		},
		Rules: []bo.Rule{
			{
				ID:       ruleID,
				Name:     ruleID + " rule",
				Enabled:  true,
				Priority: priority,
				When: bo.Condition{
					All: []bo.Condition{
						{
							Field: "request.method",
							Op:    eo.OperatorEQ,
							Value: "GET",
						},
						{
							Field: "request.path",
							Op:    eo.OperatorEQ,
							Value: path,
						},
					},
				},
				Action: bo.Action{
					Type:   eo.ActionTypeStaticResponse,
					Status: 200,
					Body:   map[string]any{"message": "ok"},
				},
			},
		},
	}
}
