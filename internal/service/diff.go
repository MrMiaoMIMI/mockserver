package service

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"

	"mockserver/internal/model/bo"
)

func summarizeRollbackDiff(current bo.PublishedRuleSetSnapshot, target bo.PublishedRuleSetSnapshot) bo.RollbackDiffSummary {
	summary := bo.RollbackDiffSummary{
		CurrentSnapshotID: current.SnapshotID,
		TargetSnapshotID:  target.SnapshotID,
		RuleSetFieldDiffs: make([]bo.FieldDiffSummary, 0),
		RuleDiffs:         make([]bo.RuleDiffSummary, 0),
		Messages:          make([]string, 0),
	}

	if current.SnapshotID == "" {
		summary.Changed = true
		summary.Messages = append(summary.Messages, "no current published snapshot; preview target would become the first published version")
		return summary
	}

	if current.SnapshotID == target.SnapshotID {
		summary.Messages = append(summary.Messages, "target snapshot is already the current published snapshot")
		return summary
	}

	summary.RuleSetFieldDiffs = append(summary.RuleSetFieldDiffs, diffValues("ruleset.enabled", current.RuleSet.Enabled, target.RuleSet.Enabled)...)
	summary.RuleSetFieldDiffs = append(summary.RuleSetFieldDiffs, diffValues("ruleset.protocol", current.RuleSet.Protocol, target.RuleSet.Protocol)...)
	summary.RuleSetFieldDiffs = append(summary.RuleSetFieldDiffs, diffValues("ruleset.namespace", current.RuleSet.Namespace, target.RuleSet.Namespace)...)
	summary.RuleSetFieldDiffs = append(summary.RuleSetFieldDiffs, diffValues("ruleset.selector", current.RuleSet.Selector, target.RuleSet.Selector)...)

	if !reflect.DeepEqual(current.RuleSet.Selector, target.RuleSet.Selector) {
		summary.Changed = true
		summary.Messages = append(summary.Messages, "ruleset selector would change")
	}
	if current.RuleSet.Enabled != target.RuleSet.Enabled {
		summary.Changed = true
		summary.Messages = append(summary.Messages, fmt.Sprintf("ruleset enabled flag would change from %t to %t", current.RuleSet.Enabled, target.RuleSet.Enabled))
	}
	if current.RuleSet.Protocol != target.RuleSet.Protocol {
		summary.Changed = true
		summary.Messages = append(summary.Messages, fmt.Sprintf("ruleset protocol would change from %s to %s", current.RuleSet.Protocol, target.RuleSet.Protocol))
	}
	if current.RuleSet.Namespace != target.RuleSet.Namespace {
		summary.Changed = true
		summary.Messages = append(summary.Messages, fmt.Sprintf("ruleset namespace would change from %s to %s", current.RuleSet.Namespace, target.RuleSet.Namespace))
	}
	currentRules := make(map[string]bo.Rule, len(current.RuleSet.Rules))
	targetRules := make(map[string]bo.Rule, len(target.RuleSet.Rules))
	keys := make(map[string]struct{})
	for _, rule := range current.RuleSet.Rules {
		currentRules[rule.ID] = rule
		keys[rule.ID] = struct{}{}
	}
	for _, rule := range target.RuleSet.Rules {
		targetRules[rule.ID] = rule
		keys[rule.ID] = struct{}{}
	}

	ruleIDs := make([]string, 0, len(keys))
	for ruleID := range keys {
		ruleIDs = append(ruleIDs, ruleID)
	}
	sort.Strings(ruleIDs)

	for _, ruleID := range ruleIDs {
		currentRule, currentOK := currentRules[ruleID]
		targetRule, targetOK := targetRules[ruleID]
		switch {
		case !currentOK && targetOK:
			summary.Changed = true
			summary.RuleDiffs = append(summary.RuleDiffs, bo.RuleDiffSummary{
				RuleID:     ruleID,
				ChangeType: "added",
				Message:    "rule would be added by rollback target",
				FieldDiffs: []bo.FieldDiffSummary{
					{
						Path:    "rule",
						Message: "rule is absent in current snapshot and present in target snapshot",
						Current: nil,
						Target:  toGeneric(targetRule),
					},
				},
			})
		case currentOK && !targetOK:
			summary.Changed = true
			summary.RuleDiffs = append(summary.RuleDiffs, bo.RuleDiffSummary{
				RuleID:     ruleID,
				ChangeType: "removed",
				Message:    "rule would be removed by rollback target",
				FieldDiffs: []bo.FieldDiffSummary{
					{
						Path:    "rule",
						Message: "rule is present in current snapshot and absent in target snapshot",
						Current: toGeneric(currentRule),
						Target:  nil,
					},
				},
			})
		default:
			conditionChanged := !reflect.DeepEqual(currentRule.When, targetRule.When)
			actionChanged := !reflect.DeepEqual(currentRule.Action, targetRule.Action)
			enabledChanged := currentRule.Enabled != targetRule.Enabled
			priorityChanged := currentRule.Priority != targetRule.Priority
			if conditionChanged || actionChanged || enabledChanged || priorityChanged {
				summary.Changed = true
				fieldDiffs := make([]bo.FieldDiffSummary, 0)
				if enabledChanged {
					fieldDiffs = append(fieldDiffs, bo.FieldDiffSummary{
						Path:    "enabled",
						Message: "rule enabled flag would change",
						Current: currentRule.Enabled,
						Target:  targetRule.Enabled,
					})
				}
				if priorityChanged {
					fieldDiffs = append(fieldDiffs, bo.FieldDiffSummary{
						Path:    "priority",
						Message: "rule priority would change",
						Current: currentRule.Priority,
						Target:  targetRule.Priority,
					})
				}
				fieldDiffs = append(fieldDiffs, diffValues("when", currentRule.When, targetRule.When)...)
				fieldDiffs = append(fieldDiffs, diffValues("action", currentRule.Action, targetRule.Action)...)

				message := "rule definition would change"
				switch {
				case actionChanged && !conditionChanged:
					message = "rule action would change"
				case conditionChanged && !actionChanged:
					message = "rule condition would change"
				case enabledChanged:
					message = "rule enabled flag would change"
				case priorityChanged:
					message = "rule priority would change"
				}
				summary.RuleDiffs = append(summary.RuleDiffs, bo.RuleDiffSummary{
					RuleID:           ruleID,
					ChangeType:       "modified",
					ConditionChanged: conditionChanged,
					ActionChanged:    actionChanged,
					Message:          message,
					FieldDiffs:       fieldDiffs,
				})
			}
		}
	}

	if !summary.Changed && len(summary.Messages) == 0 {
		summary.Messages = append(summary.Messages, "rollback target is structurally identical to current published ruleset")
	}

	return summary
}

func diffValues(prefix string, current any, target any) []bo.FieldDiffSummary {
	return diffGenericValues(prefix, toGeneric(current), toGeneric(target))
}

func diffGenericValues(prefix string, current any, target any) []bo.FieldDiffSummary {
	if reflect.DeepEqual(current, target) {
		return nil
	}

	switch currentTyped := current.(type) {
	case map[string]any:
		targetTyped, ok := target.(map[string]any)
		if !ok {
			return []bo.FieldDiffSummary{newFieldDiff(prefix, current, target)}
		}
		keys := make(map[string]struct{})
		for key := range currentTyped {
			keys[key] = struct{}{}
		}
		for key := range targetTyped {
			keys[key] = struct{}{}
		}
		sortedKeys := make([]string, 0, len(keys))
		for key := range keys {
			sortedKeys = append(sortedKeys, key)
		}
		sort.Strings(sortedKeys)
		diffs := make([]bo.FieldDiffSummary, 0)
		for _, key := range sortedKeys {
			diffs = append(diffs, diffGenericValues(joinDiffPath(prefix, key), currentTyped[key], targetTyped[key])...)
		}
		return diffs
	case []any:
		targetTyped, ok := target.([]any)
		if !ok {
			return []bo.FieldDiffSummary{newFieldDiff(prefix, current, target)}
		}
		maxLen := len(currentTyped)
		if len(targetTyped) > maxLen {
			maxLen = len(targetTyped)
		}
		diffs := make([]bo.FieldDiffSummary, 0)
		for i := 0; i < maxLen; i++ {
			var currentValue any
			var targetValue any
			if i < len(currentTyped) {
				currentValue = currentTyped[i]
			}
			if i < len(targetTyped) {
				targetValue = targetTyped[i]
			}
			diffs = append(diffs, diffGenericValues(fmt.Sprintf("%s[%d]", prefix, i), currentValue, targetValue)...)
		}
		return diffs
	default:
		return []bo.FieldDiffSummary{newFieldDiff(prefix, current, target)}
	}
}

func newFieldDiff(path string, current any, target any) bo.FieldDiffSummary {
	return bo.FieldDiffSummary{
		Path:    path,
		Message: fmt.Sprintf("field %s would change", path),
		Current: current,
		Target:  target,
	}
}

func joinDiffPath(prefix string, key string) string {
	if prefix == "" {
		return key
	}
	return prefix + "." + key
}

func toGeneric(value any) any {
	raw, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%v", value)
	}
	var generic any
	if err := json.Unmarshal(raw, &generic); err != nil {
		return fmt.Sprintf("%v", value)
	}
	return normalizeJSONNumbers(generic)
}

func normalizeJSONNumbers(value any) any {
	switch typed := value.(type) {
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, normalizeJSONNumbers(item))
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = normalizeJSONNumbers(item)
		}
		return result
	case float64:
		if typed == float64(int64(typed)) {
			return int64(typed)
		}
		return typed
	default:
		return typed
	}
}
