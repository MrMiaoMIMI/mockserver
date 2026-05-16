package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/cel-go/cel"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
	"github.com/MrMiaoMIMI/mockserver/mockprotocol"
)

var sequenceState = struct {
	sync.Mutex
	counters map[string]int
}{
	counters: make(map[string]int),
}

var newWebhookHTTPClient = func(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

type MatchOptions struct {
	ExplainOnly     bool
	ExplainMaxDepth int
	ExplainCompact  bool
	ExplainSummary  bool
}

func Match(ruleSets []CompiledRuleSet, event bo.Event) (bo.SimulationResult, error) {
	return MatchWithOptions(ruleSets, event, MatchOptions{})
}

func MatchWithOptions(ruleSets []CompiledRuleSet, event bo.Event, options MatchOptions) (bo.SimulationResult, error) {
	normalizedEvent := normalizeEvent(event)
	type candidateEntry struct {
		ruleSet      CompiledRuleSet
		explainIndex int
		specificity  int
	}
	candidates := make([]candidateEntry, 0)
	result := bo.SimulationResult{
		Explain: bo.MatchExplanation{
			RuleSetExplanations: make([]bo.RuleSetExplanation, 0, len(ruleSets)),
		},
	}
	for _, ruleSet := range ruleSets {
		eligible, specificity, explanation, err := selectorMatchWithExplain(ruleSet.RuleSet, normalizedEvent)
		if err != nil {
			return bo.SimulationResult{}, err
		}
		result.Explain.RuleSetExplanations = append(result.Explain.RuleSetExplanations, explanation)
		if !eligible {
			continue
		}
		candidates = append(candidates, candidateEntry{
			ruleSet:      ruleSet,
			explainIndex: len(result.Explain.RuleSetExplanations) - 1,
			specificity:  specificity,
		})
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].specificity != candidates[j].specificity {
			return candidates[i].specificity > candidates[j].specificity
		}
		return candidates[i].ruleSet.RuleSet.ID > candidates[j].ruleSet.RuleSet.ID
	})
	for index := 1; index < len(candidates); index++ {
		explainIndex := candidates[index].explainIndex
		result.Explain.RuleSetExplanations[explainIndex].Message = "ruleset selector matched, but a more specific ruleset was selected"
	}
	if len(candidates) > 0 {
		result.Explain.WinnerRuleSetID = candidates[0].ruleSet.RuleSet.ID
		result.Explain.RuleSetCandidates = make([]bo.RuleSetCandidate, 0, len(candidates))
		for index, candidate := range candidates {
			selected := index == 0
			message := "ruleset selector matched"
			if selected {
				message = "selected as winner ruleset"
				if len(candidates) > 1 {
					message = "selected as most specific ruleset"
				}
			} else {
				message = "ruleset selector matched, but a more specific ruleset was selected"
			}
			result.Explain.RuleSetCandidates = append(result.Explain.RuleSetCandidates, bo.RuleSetCandidate{
				RuleSetID:           candidate.ruleSet.RuleSet.ID,
				SelectorMatched:     true,
				Selected:            selected,
				SelectorSpecificity: candidate.specificity,
				Message:             message,
			})
		}
	}
	if len(candidates) > 1 {
		candidates = candidates[:1]
	}

	for _, candidate := range candidates {
		ruleSet := candidate.ruleSet
		explainIndex := candidate.explainIndex
		result.Explain.RuleSetID = ruleSet.RuleSet.ID
		result.Trace.RulesetID = ruleSet.RuleSet.ID
		result.Explain.RuleSetExplanations[explainIndex].CandidateRules = make([]string, 0)
		result.Explain.RuleSetExplanations[explainIndex].RuleExplanations = make([]bo.RuleExplanation, 0)
		candidateIndexes := candidateRuleIndexes(ruleSet, normalizedEvent)
		explainedRuleIndexes := make(map[int]bool, len(candidateIndexes))
		for _, ruleIndex := range candidateIndexes {
			explainedRuleIndexes[ruleIndex] = true
			compiledRule := ruleSet.Compiled[ruleIndex]
			result.Candidates = append(result.Candidates, compiledRule.Rule.ID)
			result.Explain.RuleSetExplanations[explainIndex].CandidateRules = append(result.Explain.RuleSetExplanations[explainIndex].CandidateRules, compiledRule.Rule.ID)
			explanation, err := evalConditionWithExplain(compiledRule, compiledRule.Rule.When, normalizedEvent)
			if err != nil {
				return bo.SimulationResult{}, err
			}
			ruleExplanation := bo.RuleExplanation{
				RuleID:    compiledRule.Rule.ID,
				Priority:  compiledRule.Rule.Priority,
				Matched:   explanation.Matched,
				Condition: explanation,
			}
			result.Explain.RuleExplanations = append(result.Explain.RuleExplanations, ruleExplanation)
			result.Explain.RuleSetExplanations[explainIndex].RuleExplanations = append(result.Explain.RuleSetExplanations[explainIndex].RuleExplanations, ruleExplanation)
			if !explanation.Matched {
				continue
			}

			var action bo.ProtocolResponse
			var actionExplain bo.ActionExplanation
			if options.ExplainOnly {
				actionExplain = bo.ActionExplanation{
					Type:    compiledRule.Rule.Action.Type,
					Message: "action execution skipped because explain_only=true",
				}
			} else {
				var err error
				action, actionExplain, err = executeActionWithExplain(ruleSet.RuleSet.ID, compiledRule, normalizedEvent)
				if err != nil {
					return bo.SimulationResult{}, err
				}
			}

			result.Matched = true
			result.Trace = bo.MatchTrace{
				RulesetID: ruleSet.RuleSet.ID,
				RuleID:    compiledRule.Rule.ID,
			}
			if !options.ExplainOnly {
				result.Response = action
			}
			result.Explain.CandidateRules = append([]string(nil), result.Candidates...)
			result.Explain.RuleExplanations[len(result.Explain.RuleExplanations)-1].ActionInfo = &actionExplain
			result.Explain.RuleSetExplanations[explainIndex].RuleExplanations[len(result.Explain.RuleSetExplanations[explainIndex].RuleExplanations)-1].ActionInfo = &actionExplain
			result.Explain.RuleSetExplanations[explainIndex].Matched = true
			result.Explain.RuleSetExplanations[explainIndex].Message = "ruleset selector matched and at least one rule matched"
			result.Explain = normalizeExplain(result.Explain, options)
			return result, nil
		}
		appendIndexedOutRuleExplanations(&result.Explain, explainIndex, ruleSet, normalizedEvent, explainedRuleIndexes)
		result.Explain.RuleSetExplanations[explainIndex].Matched = false
		result.Explain.RuleSetExplanations[explainIndex].Message = "ruleset selector matched, but no rule matched"
	}

	result.Explain.CandidateRules = append([]string(nil), result.Candidates...)
	result.Explain = normalizeExplain(result.Explain, options)
	return result, nil
}

func candidateRuleIndexes(ruleSet CompiledRuleSet, event bo.Event) []int {
	methodAllowed := indexedAllowedRules(ruleSet.MethodIndex, ruleSet.MethodWildcardRules, eventStringField(event, "request.method"))
	hostAllowed := indexedAllowedRules(ruleSet.HostIndex, ruleSet.HostWildcardRules, eventStringField(event, "request.host"))
	pathAllowed := indexedAllowedRules(ruleSet.ExactPath, ruleSet.PathWildcardRules, eventStringField(event, "request.path"))

	result := make([]int, 0, len(ruleSet.Compiled))
	for i := range ruleSet.Compiled {
		if methodAllowed[i] && hostAllowed[i] && pathAllowed[i] {
			result = append(result, i)
		}
	}
	return result
}

func appendIndexedOutRuleExplanations(explain *bo.MatchExplanation, explainIndex int, ruleSet CompiledRuleSet, event bo.Event, explained map[int]bool) {
	for ruleIndex, compiledRule := range ruleSet.Compiled {
		if explained[ruleIndex] {
			continue
		}
		ruleExplanation := bo.RuleExplanation{
			RuleID:    compiledRule.Rule.ID,
			Priority:  compiledRule.Rule.Priority,
			Matched:   false,
			Condition: indexedOutConditionExplanation(compiledRule, event),
		}
		explain.RuleExplanations = append(explain.RuleExplanations, ruleExplanation)
		explain.RuleSetExplanations[explainIndex].RuleExplanations = append(explain.RuleSetExplanations[explainIndex].RuleExplanations, ruleExplanation)
	}
}

func indexedOutConditionExplanation(compiledRule CompiledRule, event bo.Event) bo.ConditionExplanation {
	children := make([]bo.ConditionExplanation, 0, 3)
	appendMismatch := func(field string, expected string, actual string) {
		if expected == "" || actual == expected {
			return
		}
		children = append(children, bo.ConditionExplanation{
			Kind:     "predicate",
			Matched:  false,
			Field:    field,
			Operator: eo.OperatorEQ,
			Expected: expected,
			Actual:   []any{actual},
			Message:  fmt.Sprintf("field %s did not match indexed equality", field),
		})
	}
	appendMismatch("request.method", compiledRule.ExactMethod, eventStringField(event, "request.method"))
	appendMismatch("request.host", compiledRule.ExactHost, eventStringField(event, "request.host"))
	appendMismatch("request.path", compiledRule.ExactPath, eventStringField(event, "request.path"))
	if len(children) == 1 {
		return children[0]
	}
	return bo.ConditionExplanation{
		Kind:     "prefilter",
		Matched:  false,
		Message:  "rule skipped by indexed condition prefilter",
		Children: children,
	}
}

func indexedAllowedRules(index map[string][]int, wildcardRules []int, value string) map[int]bool {
	allowed := make(map[int]bool, len(wildcardRules)+len(index[value]))
	for _, item := range wildcardRules {
		allowed[item] = true
	}
	for _, item := range index[value] {
		allowed[item] = true
	}
	return allowed
}

func normalizeExplain(explain bo.MatchExplanation, options MatchOptions) bo.MatchExplanation {
	if options.ExplainMaxDepth > 0 {
		for i := range explain.RuleExplanations {
			explain.RuleExplanations[i].Condition = trimConditionExplanationDepth(explain.RuleExplanations[i].Condition, options.ExplainMaxDepth)
		}
		for i := range explain.RuleSetExplanations {
			for j := range explain.RuleSetExplanations[i].RuleExplanations {
				explain.RuleSetExplanations[i].RuleExplanations[j].Condition = trimConditionExplanationDepth(explain.RuleSetExplanations[i].RuleExplanations[j].Condition, options.ExplainMaxDepth)
			}
		}
	}
	if options.ExplainCompact {
		for i := range explain.RuleExplanations {
			compactRuleExplanation(&explain.RuleExplanations[i])
		}
		for i := range explain.RuleSetExplanations {
			for j := range explain.RuleSetExplanations[i].SelectorChecks {
				explain.RuleSetExplanations[i].SelectorChecks[j].Expected = nil
				explain.RuleSetExplanations[i].SelectorChecks[j].Actual = nil
			}
			for j := range explain.RuleSetExplanations[i].RuleExplanations {
				compactRuleExplanation(&explain.RuleSetExplanations[i].RuleExplanations[j])
			}
		}
	}
	if options.ExplainSummary {
		explain = summarizeExplain(explain)
	}
	return explain
}

func trimConditionExplanationDepth(explanation bo.ConditionExplanation, maxDepth int) bo.ConditionExplanation {
	if maxDepth <= 0 {
		explanation.Children = nil
		return explanation
	}
	for i := range explanation.Children {
		explanation.Children[i] = trimConditionExplanationDepth(explanation.Children[i], maxDepth-1)
	}
	return explanation
}

func compactRuleExplanation(explanation *bo.RuleExplanation) {
	explanation.Condition = compactConditionExplanation(explanation.Condition)
	if explanation.ActionInfo != nil {
		explanation.ActionInfo.Template = ""
		explanation.ActionInfo.Expression = ""
		explanation.ActionInfo.RenderedResult = nil
	}
}

func compactConditionExplanation(explanation bo.ConditionExplanation) bo.ConditionExplanation {
	explanation.Expected = nil
	explanation.Actual = nil
	for i := range explanation.Children {
		explanation.Children[i] = compactConditionExplanation(explanation.Children[i])
	}
	return explanation
}

func summarizeExplain(explain bo.MatchExplanation) bo.MatchExplanation {
	summarized := bo.MatchExplanation{
		RuleSetID:           explain.RuleSetID,
		WinnerRuleSetID:     explain.WinnerRuleSetID,
		RuleSetCandidates:   append([]bo.RuleSetCandidate(nil), explain.RuleSetCandidates...),
		RuleSetExplanations: make([]bo.RuleSetExplanation, 0, len(explain.RuleSetExplanations)),
	}
	for _, ruleset := range explain.RuleSetExplanations {
		summary := bo.RuleSetExplanation{
			RuleSetID: ruleset.RuleSetID,
			Matched:   ruleset.Matched,
			Message:   ruleset.Message,
		}
		if ruleset.Matched {
			for _, rule := range ruleset.RuleExplanations {
				summary.RuleExplanations = append(summary.RuleExplanations, summarizeRuleExplanation(rule))
			}
		}
		summarized.RuleSetExplanations = append(summarized.RuleSetExplanations, summary)
	}
	for _, rule := range explain.RuleExplanations {
		if rule.Matched {
			summarized.RuleExplanations = append(summarized.RuleExplanations, summarizeRuleExplanation(rule))
			break
		}
	}
	return summarized
}

func summarizeRuleExplanation(rule bo.RuleExplanation) bo.RuleExplanation {
	summary := bo.RuleExplanation{
		RuleID:   rule.RuleID,
		Priority: rule.Priority,
		Matched:  rule.Matched,
		Condition: bo.ConditionExplanation{
			Kind:    rule.Condition.Kind,
			Matched: rule.Condition.Matched,
			Field:   rule.Condition.Field,
			Expr:    rule.Condition.Expr,
			Message: rule.Condition.Message,
		},
	}
	if rule.ActionInfo != nil {
		summary.ActionInfo = &bo.ActionExplanation{
			Type:    rule.ActionInfo.Type,
			Message: rule.ActionInfo.Message,
		}
	}
	return summary
}

func selectorMatchWithExplain(ruleSet bo.RuleSet, event bo.Event) (bool, int, bo.RuleSetExplanation, error) {
	explanation := bo.RuleSetExplanation{
		RuleSetID:      ruleSet.ID,
		Matched:        true,
		SelectorChecks: make([]bo.SelectorCheckExplanation, 0, 5),
	}

	appendCheck := func(name string, matched bool, expected any, actual any, message string) {
		explanation.SelectorChecks = append(explanation.SelectorChecks, bo.SelectorCheckExplanation{
			Name:     name,
			Matched:  matched,
			Expected: expected,
			Actual:   actual,
			Message:  message,
		})
		if !matched {
			explanation.Matched = false
		}
	}

	appendCheck("enabled", ruleSet.Enabled, true, ruleSet.Enabled, selectorMessage(ruleSet.Enabled, "ruleset is enabled", "ruleset is disabled"))
	appendCheck("protocol", ruleSet.Protocol == event.Protocol, ruleSet.Protocol, event.Protocol, selectorMessage(ruleSet.Protocol == event.Protocol, "protocol matched", "protocol did not match"))
	appendCheck("namespace", ruleSet.Namespace == event.Namespace, ruleSet.Namespace, event.Namespace, selectorMessage(ruleSet.Namespace == event.Namespace, "namespace matched", "namespace did not match"))

	for i, condition := range ruleSet.Selector.All {
		conditionExplanation, err := evalConditionWithExplain(CompiledRule{ConditionPrograms: map[string]cel.Program{}}, condition, event)
		if err != nil {
			return false, 0, bo.RuleSetExplanation{}, err
		}
		name := selectorCheckName(condition, i)
		appendCheck(name, conditionExplanation.Matched, condition.Value, conditionExplanation.Actual, conditionExplanation.Message)
	}

	if explanation.Matched {
		explanation.Message = "ruleset selector matched"
	} else {
		explanation.Message = "ruleset selector did not match"
	}
	return explanation.Matched, selectorSpecificity(ruleSet.Selector, event), explanation, nil
}

func selectorSpecificity(selector bo.Selector, event bo.Event) int {
	score := 0
	for _, condition := range selector.All {
		score += selectorConditionSpecificity(condition, event)
	}
	return score
}

func selectorConditionSpecificity(condition bo.Condition, event bo.Event) int {
	switch {
	case len(condition.All) > 0:
		score := 0
		for _, child := range condition.All {
			score += selectorConditionSpecificity(child, event)
		}
		return score
	case len(condition.Any) > 0:
		best := 0
		for _, child := range condition.Any {
			score := selectorConditionSpecificity(child, event)
			if score > best {
				best = score
			}
		}
		return best
	case condition.Not != nil:
		return selectorConditionSpecificity(*condition.Not, event)
	}

	value, ok := condition.Value.(string)
	if !ok {
		return 10
	}
	actual := eventStringField(event, condition.Field)
	switch condition.Op {
	case eo.OperatorEQ:
		if actual == value {
			return 10_000 + len(value)
		}
	case eo.OperatorPrefix:
		if prefixMatches(condition.Field, actual, value) {
			return 1_000 + len(value)*100
		}
	case eo.OperatorContains:
		if strings.Contains(actual, value) {
			return 500 + len(value)
		}
	case eo.OperatorExists, eo.OperatorIsNotNull:
		return 100
	}
	return 0
}

func prefixMatches(field string, actual string, expected string) bool {
	if field != "request.path" {
		return strings.HasPrefix(actual, expected)
	}
	return pathPrefixMatches(actual, expected)
}

func pathPrefixMatches(path string, prefix string) bool {
	if prefix == "" {
		return true
	}
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	if prefix == "/" || path == prefix || strings.HasSuffix(prefix, "/") {
		return true
	}
	return strings.HasPrefix(path, prefix+"/")
}

func selectorCheckName(condition bo.Condition, index int) string {
	if condition.Field != "" {
		return condition.Field
	}
	switch {
	case len(condition.All) > 0:
		return fmt.Sprintf("selector.all[%d]", index)
	case len(condition.Any) > 0:
		return fmt.Sprintf("selector.any[%d]", index)
	case condition.Not != nil:
		return fmt.Sprintf("selector.not[%d]", index)
	default:
		return fmt.Sprintf("selector[%d]", index)
	}
}

func selectorMessage(matched bool, matchedMessage, unmatchedMessage string) string {
	if matched {
		return matchedMessage
	}
	return unmatchedMessage
}

func anyString(values []string, fn func(string) bool) bool {
	for _, value := range values {
		if fn(value) {
			return true
		}
	}
	return false
}

func evalConditionWithExplain(compiledRule CompiledRule, condition bo.Condition, event bo.Event) (bo.ConditionExplanation, error) {
	switch {
	case len(condition.All) > 0:
		explanation := bo.ConditionExplanation{
			Kind:     "all",
			Matched:  true,
			Message:  "all child conditions must match",
			Children: make([]bo.ConditionExplanation, 0, len(condition.All)),
		}
		for _, child := range condition.All {
			childExplanation, err := evalConditionWithExplain(compiledRule, child, event)
			if err != nil {
				return bo.ConditionExplanation{}, err
			}
			explanation.Children = append(explanation.Children, childExplanation)
			if !childExplanation.Matched {
				explanation.Matched = false
			}
		}
		if explanation.Matched {
			explanation.Message = "all child conditions matched"
		} else {
			explanation.Message = "at least one child condition did not match"
		}
		return explanation, nil
	case len(condition.Any) > 0:
		explanation := bo.ConditionExplanation{
			Kind:     "any",
			Message:  "at least one child condition must match",
			Children: make([]bo.ConditionExplanation, 0, len(condition.Any)),
		}
		for _, child := range condition.Any {
			childExplanation, err := evalConditionWithExplain(compiledRule, child, event)
			if err != nil {
				return bo.ConditionExplanation{}, err
			}
			explanation.Children = append(explanation.Children, childExplanation)
			if childExplanation.Matched {
				explanation.Matched = true
			}
		}
		if explanation.Matched {
			explanation.Message = "at least one child condition matched"
		} else {
			explanation.Message = "no child condition matched"
		}
		return explanation, nil
	case condition.Not != nil:
		childExplanation, err := evalConditionWithExplain(compiledRule, *condition.Not, event)
		if err != nil {
			return bo.ConditionExplanation{}, err
		}
		return bo.ConditionExplanation{
			Kind:     "not",
			Matched:  !childExplanation.Matched,
			Message:  "child condition result is negated",
			Children: []bo.ConditionExplanation{childExplanation},
		}, nil
	case strings.TrimSpace(condition.Expr) != "":
		program, ok := compiledRule.ConditionPrograms[condition.Expr]
		if !ok {
			return bo.ConditionExplanation{}, fmt.Errorf("compiled CEL condition not found for expression %q", condition.Expr)
		}
		matched, err := evalCELBool(program, buildCELInput(event))
		if err != nil {
			return bo.ConditionExplanation{}, err
		}
		message := "CEL expression evaluated to false"
		if matched {
			message = "CEL expression evaluated to true"
		}
		return bo.ConditionExplanation{
			Kind:    "expr",
			Matched: matched,
			Expr:    condition.Expr,
			Message: message,
		}, nil
	default:
		return evalPredicateWithExplain(condition, event)
	}
}

func evalPredicateWithExplain(condition bo.Condition, event bo.Event) (bo.ConditionExplanation, error) {
	values, found := resolveField(event, condition.Field)
	explanation := bo.ConditionExplanation{
		Kind:     "predicate",
		Field:    condition.Field,
		Operator: condition.Op,
		Expected: condition.Value,
		Actual:   values,
	}
	switch condition.Op {
	case eo.OperatorExists:
		explanation.Matched = found
		if found {
			explanation.Message = "field exists"
		} else {
			explanation.Message = "field does not exist"
		}
		return explanation, nil
	case eo.OperatorNotExists:
		explanation.Matched = !found
		if !found {
			explanation.Message = "field does not exist"
		} else {
			explanation.Message = "field exists"
		}
		return explanation, nil
	case eo.OperatorIsNull:
		explanation.Matched = !found || anyValue(values, func(v any) bool { return v == nil })
		if explanation.Matched {
			explanation.Message = "field is missing or null"
		} else {
			explanation.Message = "field exists and is not null"
		}
		return explanation, nil
	case eo.OperatorIsNotNull:
		explanation.Matched = found && anyValue(values, func(v any) bool { return v != nil })
		if explanation.Matched {
			explanation.Message = "field exists and is not null"
		} else {
			explanation.Message = "field is missing or null"
		}
		return explanation, nil
	}

	if !found {
		explanation.Matched = false
		explanation.Message = "field not found"
		return explanation, nil
	}

	var matched bool
	switch condition.Op {
	case eo.OperatorEQ:
		matched = anyValue(values, func(v any) bool { return compareValue(v, condition.Value) == 0 })
	case eo.OperatorNE:
		matched = anyValue(values, func(v any) bool { return compareValue(v, condition.Value) != 0 })
	case eo.OperatorContains:
		matched = anyValue(values, func(v any) bool { return strings.Contains(toString(v), toString(condition.Value)) })
	case eo.OperatorNotContain:
		matched = anyValue(values, func(v any) bool { return !strings.Contains(toString(v), toString(condition.Value)) })
	case eo.OperatorPrefix:
		matched = anyValue(values, func(v any) bool {
			return prefixMatches(condition.Field, toString(v), toString(condition.Value))
		})
	case eo.OperatorSuffix:
		matched = anyValue(values, func(v any) bool { return strings.HasSuffix(toString(v), toString(condition.Value)) })
	case eo.OperatorIn:
		matched = compareIn(values, condition.Value)
	case eo.OperatorNotIn:
		matched = !compareIn(values, condition.Value)
	case eo.OperatorRegex:
		rx, err := regexp.Compile(toString(condition.Value))
		if err != nil {
			return bo.ConditionExplanation{}, err
		}
		matched = anyValue(values, func(v any) bool { return rx.MatchString(toString(v)) })
	case eo.OperatorGT:
		matched = anyValue(values, func(v any) bool {
			comparison, ok := compareNumeric(v, condition.Value)
			return ok && comparison > 0
		})
	case eo.OperatorGTE:
		matched = anyValue(values, func(v any) bool {
			comparison, ok := compareNumeric(v, condition.Value)
			return ok && comparison >= 0
		})
	case eo.OperatorLT:
		matched = anyValue(values, func(v any) bool {
			comparison, ok := compareNumeric(v, condition.Value)
			return ok && comparison < 0
		})
	case eo.OperatorLTE:
		matched = anyValue(values, func(v any) bool {
			comparison, ok := compareNumeric(v, condition.Value)
			return ok && comparison <= 0
		})
	default:
		return bo.ConditionExplanation{}, fmt.Errorf("unsupported operator: %s", condition.Op)
	}

	explanation.Matched = matched
	if matched {
		explanation.Message = fmt.Sprintf("field %s matched operator %s", condition.Field, condition.Op)
	} else {
		explanation.Message = fmt.Sprintf("field %s did not match operator %s", condition.Field, condition.Op)
	}
	return explanation, nil
}

func executeActionWithExplain(ruleSetID string, compiledRule CompiledRule, event bo.Event) (bo.ProtocolResponse, bo.ActionExplanation, error) {
	action := compiledRule.Rule.Action
	switch actionRenderer(action) {
	case eo.ActionRendererStatic:
		response, err := normalizeActionResponse(event.Protocol, action.Response)
		if err != nil {
			return bo.ProtocolResponse{}, bo.ActionExplanation{}, err
		}
		return response, bo.ActionExplanation{
			Type:           eo.ActionRendererStatic,
			RenderedResult: response.Payload,
			Message:        "static response returned as configured",
		}, nil
	case eo.ActionRendererTemplate:
		var out bytes.Buffer
		if err := compiledRule.Template.Execute(&out, buildEventDocument(event)); err != nil {
			return bo.ProtocolResponse{}, bo.ActionExplanation{}, fmt.Errorf("execute template: %w", err)
		}
		response, err := protocolResponseFromJSON(event.Protocol, out.Bytes())
		if err != nil {
			return bo.ProtocolResponse{}, bo.ActionExplanation{}, fmt.Errorf("decode template response payload: %w", err)
		}
		return response, bo.ActionExplanation{
			Type:           eo.ActionRendererTemplate,
			Template:       action.ResponseTemplate,
			RenderedResult: response.Payload,
			Message:        "template rendered successfully",
		}, nil
	case eo.ActionRendererCEL:
		payload, err := evalCELValue(compiledRule.ResponseProgram, buildCELInput(event))
		if err != nil {
			return bo.ProtocolResponse{}, bo.ActionExplanation{}, fmt.Errorf("execute CEL response: %w", err)
		}
		response, err := protocolResponseFromAny(event.Protocol, payload)
		if err != nil {
			return bo.ProtocolResponse{}, bo.ActionExplanation{}, err
		}
		return response, bo.ActionExplanation{
			Type:           eo.ActionRendererCEL,
			Expression:     action.ResponseExpression,
			RenderedResult: response.Payload,
			Message:        "CEL response rendered successfully",
		}, nil
	case eo.ActionRendererSequence:
		step, index := nextSequenceStep(ruleSetID, compiledRule.Rule)
		stepResponse := step.Response
		response, err := normalizeActionResponse(event.Protocol, &stepResponse)
		if err != nil {
			return bo.ProtocolResponse{}, bo.ActionExplanation{}, err
		}
		return response, bo.ActionExplanation{
			Type:           eo.ActionRendererSequence,
			RenderedResult: response.Payload,
			Message:        fmt.Sprintf("sequence step %d returned", index),
		}, nil
	case eo.ActionRendererWebhook:
		webhookResponse, err := executeWebhook(compiledRule.Rule.Action, event)
		if err != nil {
			return bo.ProtocolResponse{}, bo.ActionExplanation{}, err
		}
		return webhookResponse, bo.ActionExplanation{
			Type:           eo.ActionRendererWebhook,
			RenderedResult: webhookResponse.Payload,
			Message:        "webhook response returned",
		}, nil
	default:
		return bo.ProtocolResponse{}, bo.ActionExplanation{}, fmt.Errorf("unsupported response renderer: %s", action.Renderer)
	}
}

func nextSequenceStep(ruleSetID string, rule bo.Rule) (bo.SequenceStep, int) {
	sequenceState.Lock()
	defer sequenceState.Unlock()

	key := ruleSetID + "/" + rule.ID
	current := sequenceState.counters[key]
	stepIndex := current
	if stepIndex >= len(rule.Action.Sequence) {
		switch rule.Action.SequenceStrategy {
		case eo.SequenceStrategyLast:
			stepIndex = len(rule.Action.Sequence) - 1
		default:
			stepIndex = current % len(rule.Action.Sequence)
		}
	}
	sequenceState.counters[key] = current + 1
	return rule.Action.Sequence[stepIndex], stepIndex
}

func executeWebhook(action bo.Action, event bo.Event) (bo.ProtocolResponse, error) {
	if action.Webhook == nil {
		return bo.ProtocolResponse{}, fmt.Errorf("webhook config is required")
	}

	method := strings.ToUpper(strings.TrimSpace(action.Webhook.Method))
	if method == "" {
		method = http.MethodPost
	}
	timeout := time.Duration(action.Webhook.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 3 * time.Second
	}

	var body io.Reader
	if method != http.MethodGet && method != http.MethodHead {
		raw, err := json.Marshal(map[string]any{
			"event": buildEventDocument(event),
		})
		if err != nil {
			return bo.ProtocolResponse{}, fmt.Errorf("marshal webhook request: %w", err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, action.Webhook.URL, body)
	if err != nil {
		return bo.ProtocolResponse{}, fmt.Errorf("create webhook request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for key, values := range action.Webhook.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	resp, err := newWebhookHTTPClient(timeout).Do(req)
	if err != nil {
		return bo.ProtocolResponse{}, fmt.Errorf("execute webhook: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return bo.ProtocolResponse{}, fmt.Errorf("read webhook response: %w", err)
	}
	return protocolResponseFromJSON(event.Protocol, raw)
}

func normalizeActionResponse(protocol string, response *bo.ProtocolResponse) (bo.ProtocolResponse, error) {
	if response == nil {
		response = &bo.ProtocolResponse{}
	}
	if strings.TrimSpace(response.Protocol) != "" && !strings.EqualFold(response.Protocol, protocol) {
		return bo.ProtocolResponse{}, fmt.Errorf("response protocol %q does not match event protocol %q", response.Protocol, protocol)
	}
	payload, err := mockprotocol.NormalizeResponsePayload(protocol, response.Payload)
	if err != nil {
		return bo.ProtocolResponse{}, err
	}
	return bo.ProtocolResponse{
		Protocol: strings.ToLower(strings.TrimSpace(protocol)),
		Payload:  payload,
	}, nil
}

func protocolResponseFromJSON(protocol string, raw []byte) (bo.ProtocolResponse, error) {
	var payload map[string]any
	if len(strings.TrimSpace(string(raw))) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil {
			return bo.ProtocolResponse{}, err
		}
	}
	return protocolResponseFromAny(protocol, payload)
}

func protocolResponseFromAny(protocol string, value any) (bo.ProtocolResponse, error) {
	payload, ok := value.(map[string]any)
	if !ok && value != nil {
		return bo.ProtocolResponse{}, fmt.Errorf("response renderer must return an object payload")
	}
	return normalizeActionResponse(protocol, &bo.ProtocolResponse{Payload: payload})
}

func normalizeEvent(event bo.Event) bo.Event {
	event.Protocol = strings.ToLower(event.Protocol)
	event.Namespace = strings.ToLower(event.Namespace)
	if event.Request == nil {
		event.Request = bo.EventRequest{}
	}
	if method := bo.RequestString(event.Request, "method"); method != "" {
		event.Request["method"] = strings.ToUpper(method)
	}
	if host := bo.RequestString(event.Request, "host"); host != "" {
		event.Request["host"] = strings.ToLower(host)
	}
	headers := bo.RequestStringMap(event.Request, "headers")
	if headers != nil {
		normalizedHeaders := make(map[string][]string, len(headers))
		for key, values := range headers {
			normalizedHeaders[strings.ToLower(key)] = values
		}
		event.Request["headers"] = normalizedHeaders
	}
	return event
}

func eventStringField(event bo.Event, field string) string {
	values, found := resolveField(event, field)
	if !found || len(values) == 0 {
		return ""
	}
	return bo.RequestValueString(values[0])
}

func normalizeHeaderMap(headers map[string][]string) map[string][]string {
	normalizedHeaders := make(map[string][]string, len(headers))
	for key, values := range headers {
		normalizedHeaders[strings.ToLower(key)] = values
	}
	return normalizedHeaders
}

func buildCELInput(event bo.Event) map[string]any {
	document := buildEventDocument(event)
	return map[string]any{
		"event":     document,
		"request":   document["request"],
		"meta":      document["meta"],
		"protocol":  document["protocol"],
		"namespace": document["namespace"],
	}
}

func compareIn(values []any, expected any) bool {
	expectedList, ok := expected.([]any)
	if !ok {
		raw, marshalErr := json.Marshal(expected)
		if marshalErr != nil {
			return false
		}
		if err := json.Unmarshal(raw, &expectedList); err != nil {
			return false
		}
	}
	for _, value := range values {
		for _, candidate := range expectedList {
			if compareValue(value, candidate) == 0 {
				return true
			}
		}
	}
	return false
}

func compareValue(left any, right any) int {
	if jsonValueEqual(left, right) {
		return 0
	}
	return 1
}

func compareNumeric(left any, right any) (int, bool) {
	leftValue, leftOK := toFloat(left)
	rightValue, rightOK := toFloat(right)
	if !leftOK || !rightOK {
		return 0, false
	}
	switch {
	case leftValue > rightValue:
		return 1, true
	case leftValue < rightValue:
		return -1, true
	default:
		return 0, true
	}
}

func toFloat(value any) (float64, bool) {
	switch typed := value.(type) {
	case int:
		return float64(typed), true
	case int32:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case float32:
		return float64(typed), true
	case float64:
		return typed, true
	case json.Number:
		result, err := typed.Float64()
		return result, err == nil
	default:
		return 0, false
	}
}

func jsonValueEqual(left any, right any) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	if leftNumber, ok := toFloat(left); ok {
		rightNumber, rightOK := toFloat(right)
		return rightOK && leftNumber == rightNumber
	}
	if _, ok := toFloat(right); ok {
		return false
	}

	switch leftTyped := left.(type) {
	case string:
		rightTyped, ok := right.(string)
		return ok && leftTyped == rightTyped
	case bool:
		rightTyped, ok := right.(bool)
		return ok && leftTyped == rightTyped
	case []any:
		rightTyped, ok := toJSONSlice(right)
		return ok && jsonSliceEqual(leftTyped, rightTyped)
	case []string:
		rightTyped, ok := toJSONSlice(right)
		return ok && jsonSliceEqual(stringSliceToAny(leftTyped), rightTyped)
	case map[string]any:
		rightTyped, ok := toJSONMap(right)
		return ok && jsonMapEqual(leftTyped, rightTyped)
	case map[string]string:
		rightTyped, ok := toJSONMap(right)
		return ok && jsonMapEqual(stringMapToAny(leftTyped), rightTyped)
	default:
		return false
	}
}

func toJSONSlice(value any) ([]any, bool) {
	switch typed := value.(type) {
	case []any:
		return typed, true
	case []string:
		return stringSliceToAny(typed), true
	default:
		return nil, false
	}
}

func jsonSliceEqual(left []any, right []any) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !jsonValueEqual(left[i], right[i]) {
			return false
		}
	}
	return true
}

func toJSONMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case map[string]string:
		return stringMapToAny(typed), true
	default:
		return nil, false
	}
}

func jsonMapEqual(left map[string]any, right map[string]any) bool {
	if len(left) != len(right) {
		return false
	}
	for key, leftValue := range left {
		rightValue, ok := right[key]
		if !ok || !jsonValueEqual(leftValue, rightValue) {
			return false
		}
	}
	return true
}

func stringMapToAny(values map[string]string) map[string]any {
	result := make(map[string]any, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case []byte:
		return string(typed)
	default:
		raw, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprintf("%v", typed)
		}
		return string(raw)
	}
}

func anyValue(values []any, fn func(any) bool) bool {
	for _, value := range values {
		if fn(value) {
			return true
		}
	}
	return false
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}

func stringSliceToAny(values []string) []any {
	result := make([]any, 0, len(values))
	for _, value := range values {
		result = append(result, value)
	}
	return result
}

func cloneHeaders(headers map[string][]string) map[string][]string {
	if len(headers) == 0 {
		return nil
	}
	cloned := make(map[string][]string, len(headers))
	for key, values := range headers {
		copied := make([]string, len(values))
		copy(copied, values)
		cloned[key] = copied
	}
	return cloned
}
