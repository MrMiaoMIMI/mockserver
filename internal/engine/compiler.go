package engine

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/google/cel-go/cel"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
	"github.com/MrMiaoMIMI/mockserver/mockprotocol"
)

const maxRegexPatternLength = 512
const maxRuleCodeLength = 64

type CompiledRuleSet struct {
	RuleSet             bo.RuleSet
	Compiled            []CompiledRule
	MethodIndex         map[string][]int
	HostIndex           map[string][]int
	ExactPath           map[string][]int
	MethodWildcardRules []int
	HostWildcardRules   []int
	PathWildcardRules   []int
}

type CompiledRule struct {
	Rule              bo.Rule
	Specificity       int
	Template          *template.Template
	ConditionPrograms map[string]cel.Program
	BodyProgram       cel.Program
	ExactMethod       string
	ExactHost         string
	ExactPath         string
}

func ValidateRuleSet(ruleSet bo.RuleSet) bo.ValidationResult {
	issues := make([]bo.ValidationIssue, 0)
	warnings := make([]bo.ValidationIssue, 0)

	if strings.TrimSpace(ruleSet.ID) == "" {
		issues = append(issues, bo.ValidationIssue{Path: "id", Message: "id is required"})
	}
	if strings.TrimSpace(ruleSet.Protocol) == "" {
		issues = append(issues, bo.ValidationIssue{Path: "protocol", Message: "protocol is required"})
	} else if _, ok := mockprotocol.DefaultRegistry().Get(ruleSet.Protocol); !ok {
		issues = append(issues, bo.ValidationIssue{Path: "protocol", Message: "unsupported protocol"})
	}
	if strings.TrimSpace(ruleSet.Namespace) == "" {
		issues = append(issues, bo.ValidationIssue{Path: "namespace", Message: "namespace is required"})
	}
	if len(ruleSet.Selector.All) == 0 {
		issues = append(issues, bo.ValidationIssue{Path: "selector", Message: "selector must contain at least one condition"})
	} else {
		validateSelector(ruleSet.Protocol, ruleSet.Selector, "selector", &issues)
	}
	if len(ruleSet.Rules) == 0 {
		issues = append(issues, bo.ValidationIssue{Path: "rules", Message: "rules must contain at least one rule"})
	}
	ruleIDs := make(map[string]int, len(ruleSet.Rules))
	for i, rule := range ruleSet.Rules {
		rulePath := fmt.Sprintf("rules[%d]", i)
		ruleID := strings.TrimSpace(rule.ID)
		if ruleID != "" {
			if firstIndex, exists := ruleIDs[ruleID]; exists {
				issues = append(issues, bo.ValidationIssue{Path: rulePath + ".id", Message: fmt.Sprintf("duplicate rule id %q already used at rules[%d]", ruleID, firstIndex)})
			} else {
				ruleIDs[ruleID] = i
			}
		}
		validateRule(ruleSet.Protocol, rule, rulePath, &issues)
	}
	if len(issues) == 0 {
		validateRuleReachability(ruleSet, &warnings)
	}

	return bo.ValidationResult{
		Valid:    len(issues) == 0,
		Issues:   issues,
		Warnings: warnings,
	}
}

func CompileRuleSet(ruleSet bo.RuleSet) (CompiledRuleSet, error) {
	validation := ValidateRuleSet(ruleSet)
	if !validation.Valid {
		return CompiledRuleSet{}, fmt.Errorf("ruleset validation failed")
	}

	compiled := CompiledRuleSet{
		RuleSet:     normalizeRuleSet(ruleSet),
		MethodIndex: make(map[string][]int),
		HostIndex:   make(map[string][]int),
		ExactPath:   make(map[string][]int),
	}

	for _, rule := range compiled.RuleSet.Rules {
		if !rule.Enabled {
			continue
		}

		exactPath, _ := extractEquality(rule.When, "request.path")
		exactMethod, _ := extractEquality(rule.When, "request.method")
		exactHost, _ := extractEquality(rule.When, "request.host")
		entry := CompiledRule{
			Rule:              rule,
			Specificity:       specificity(rule.When),
			ConditionPrograms: make(map[string]cel.Program),
			ExactMethod:       strings.ToUpper(exactMethod),
			ExactHost:         strings.ToLower(exactHost),
			ExactPath:         exactPath,
		}
		if rule.Action.Type == eo.ActionTypeTemplateResponse && rule.Action.BodyTemplate != "" {
			tpl, err := newRuleTemplate(rule.ID, rule.Action.BodyTemplate)
			if err != nil {
				return CompiledRuleSet{}, fmt.Errorf("parse template for rule %s: %w", rule.ID, err)
			}
			entry.Template = tpl
		}
		for _, expression := range collectConditionExpressions(rule.When) {
			program, err := compileCELCondition(expression)
			if err != nil {
				return CompiledRuleSet{}, fmt.Errorf("compile CEL condition for rule %s: %w", rule.ID, err)
			}
			entry.ConditionPrograms[expression] = program
		}
		if rule.Action.Type == eo.ActionTypeCELResponse && strings.TrimSpace(rule.Action.BodyExpression) != "" {
			program, err := compileCELBody(rule.Action.BodyExpression)
			if err != nil {
				return CompiledRuleSet{}, fmt.Errorf("compile CEL body for rule %s: %w", rule.ID, err)
			}
			entry.BodyProgram = program
		}
		compiled.Compiled = append(compiled.Compiled, entry)
	}

	sort.SliceStable(compiled.Compiled, func(i, j int) bool {
		if compiled.Compiled[i].Rule.Priority != compiled.Compiled[j].Rule.Priority {
			return compiled.Compiled[i].Rule.Priority > compiled.Compiled[j].Rule.Priority
		}
		if compiled.Compiled[i].Specificity != compiled.Compiled[j].Specificity {
			return compiled.Compiled[i].Specificity > compiled.Compiled[j].Specificity
		}
		return compiled.Compiled[i].Rule.ID < compiled.Compiled[j].Rule.ID
	})
	indexCompiledRules(&compiled)

	return compiled, nil
}

func normalizeRuleSet(ruleSet bo.RuleSet) bo.RuleSet {
	ruleSet.Protocol = strings.ToLower(ruleSet.Protocol)
	ruleSet.Namespace = strings.ToLower(ruleSet.Namespace)
	ruleSet.Selector.All = normalizeConditions(ruleSet.Selector.All)
	for i := range ruleSet.Rules {
		ruleSet.Rules[i].When = normalizeCondition(ruleSet.Rules[i].When)
	}
	return ruleSet
}

func normalizeConditions(conditions []bo.Condition) []bo.Condition {
	for i := range conditions {
		conditions[i] = normalizeCondition(conditions[i])
	}
	return conditions
}

func normalizeCondition(condition bo.Condition) bo.Condition {
	condition.All = normalizeConditions(condition.All)
	condition.Any = normalizeConditions(condition.Any)
	if condition.Not != nil {
		normalized := normalizeCondition(*condition.Not)
		condition.Not = &normalized
	}
	if value, ok := condition.Value.(string); ok {
		switch condition.Field {
		case "request.host", "request.original_host":
			condition.Value = strings.ToLower(value)
		case "request.method":
			condition.Value = strings.ToUpper(value)
		case "request.operation":
			condition.Value = strings.ToLower(value)
		}
	}
	return condition
}

func validateSelector(protocol string, selector bo.Selector, path string, issues *[]bo.ValidationIssue) {
	for i, condition := range selector.All {
		validateSelectorCondition(protocol, condition, fmt.Sprintf("%s.all[%d]", path, i), issues)
	}
}

func validateSelectorCondition(protocol string, condition bo.Condition, path string, issues *[]bo.ValidationIssue) {
	groupCount := 0
	if len(condition.All) > 0 {
		groupCount++
	}
	if len(condition.Any) > 0 {
		groupCount++
	}
	if condition.Not != nil {
		groupCount++
	}
	if strings.TrimSpace(condition.Expr) != "" {
		groupCount++
	}
	if condition.Field != "" || condition.Op != "" {
		groupCount++
	}
	if groupCount != 1 {
		*issues = append(*issues, bo.ValidationIssue{Path: path, Message: "selector node must define exactly one of all/any/not/predicate"})
		return
	}
	if strings.TrimSpace(condition.Expr) != "" {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".expr", Message: "selector does not support expr"})
		return
	}
	if condition.Field != "" || condition.Op != "" {
		if strings.TrimSpace(condition.Field) == "" {
			*issues = append(*issues, bo.ValidationIssue{Path: path + ".field", Message: "field is required"})
		}
		if !isSupportedOperator(condition.Op) {
			*issues = append(*issues, bo.ValidationIssue{Path: path + ".op", Message: "unsupported operator"})
		}
		if _, err := parsePath(condition.Field); err != nil {
			*issues = append(*issues, bo.ValidationIssue{Path: path + ".field", Message: err.Error()})
		}
		if strings.TrimSpace(protocol) != "" {
			if _, ok := mockprotocol.SelectorForPath(protocol, condition.Field); !ok {
				*issues = append(*issues, bo.ValidationIssue{Path: path + ".field", Message: "field is not registered as selector for protocol"})
			} else if !mockprotocol.SelectorOperatorAllowed(protocol, condition.Field, condition.Op) {
				*issues = append(*issues, bo.ValidationIssue{Path: path + ".op", Message: "operator is not allowed for selector field"})
			}
		}
		if condition.Op == eo.OperatorRegex {
			validateRegexCondition(condition, path, issues)
		}
	}
	for i, child := range condition.All {
		validateSelectorCondition(protocol, child, fmt.Sprintf("%s.all[%d]", path, i), issues)
	}
	for i, child := range condition.Any {
		validateSelectorCondition(protocol, child, fmt.Sprintf("%s.any[%d]", path, i), issues)
	}
	if condition.Not != nil {
		validateSelectorCondition(protocol, *condition.Not, path+".not", issues)
	}
}

func validateStringList(values []string, path string, label string, valid func(string) bool, issues *[]bo.ValidationIssue) {
	seen := make(map[string]int, len(values))
	for i, value := range values {
		itemPath := fmt.Sprintf("%s[%d]", path, i)
		trimmed := strings.TrimSpace(value)
		if !valid(trimmed) {
			*issues = append(*issues, bo.ValidationIssue{Path: itemPath, Message: fmt.Sprintf("invalid %s", label)})
			continue
		}
		key := strings.ToLower(trimmed)
		if firstIndex, exists := seen[key]; exists {
			*issues = append(*issues, bo.ValidationIssue{Path: itemPath, Message: fmt.Sprintf("duplicate %s already used at %s[%d]", label, path, firstIndex)})
			continue
		}
		seen[key] = i
	}
}

func isValidHTTPMethod(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		default:
			return false
		}
	}
	return true
}

func isValidAbsolutePath(value string) bool {
	return strings.HasPrefix(value, "/")
}

func validateRule(protocol string, rule bo.Rule, path string, issues *[]bo.ValidationIssue) {
	ruleID := strings.TrimSpace(rule.ID)
	if ruleID == "" {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".id", Message: "rule id is required"})
	} else if !isValidRuleCode(ruleID) {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".id", Message: "rule id can only contain letters, numbers, underscores and hyphens, and must be at most 64 characters"})
	}
	if strings.TrimSpace(rule.Name) == "" {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".name", Message: "rule name is required"})
	}
	if rule.Action.Type == "" {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".action.type", Message: "action type is required"})
	}
	if requiresTopLevelStatus(rule.Action.Type) && rule.Action.Status == 0 {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".action.status", Message: "action status is required"})
	}
	if rule.Action.Status != 0 && (rule.Action.Status < 100 || rule.Action.Status > 599) {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".action.status", Message: "status must be between 100 and 599"})
	}
	validateHeaders(rule.Action.Headers, path+".action.headers", issues)
	if rule.Action.Type != eo.ActionTypeStaticResponse &&
		rule.Action.Type != eo.ActionTypeTemplateResponse &&
		rule.Action.Type != eo.ActionTypeCELResponse &&
		rule.Action.Type != eo.ActionTypeSequenceResponse &&
		rule.Action.Type != eo.ActionTypeWebhookResponse {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".action.type", Message: "unsupported action type"})
	}
	if rule.Action.Type == eo.ActionTypeTemplateResponse && strings.TrimSpace(rule.Action.BodyTemplate) == "" {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".action.body_template", Message: "body_template is required for template response"})
	}
	if rule.Action.Type == eo.ActionTypeTemplateResponse && strings.TrimSpace(rule.Action.BodyTemplate) != "" {
		if _, err := newRuleTemplate(rule.ID, rule.Action.BodyTemplate); err != nil {
			*issues = append(*issues, bo.ValidationIssue{Path: path + ".action.body_template", Message: err.Error()})
		}
	}
	if rule.Action.Type == eo.ActionTypeCELResponse && strings.TrimSpace(rule.Action.BodyExpression) == "" {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".action.body_expression", Message: "body_expression is required for cel response"})
	}
	if rule.Action.Type == eo.ActionTypeCELResponse && strings.TrimSpace(rule.Action.BodyExpression) != "" {
		if _, err := compileCELBody(rule.Action.BodyExpression); err != nil {
			*issues = append(*issues, bo.ValidationIssue{Path: path + ".action.body_expression", Message: err.Error()})
		}
	}
	if rule.Action.Type == eo.ActionTypeSequenceResponse {
		validateSequenceAction(rule.Action, path+".action", issues)
	}
	if rule.Action.Type == eo.ActionTypeWebhookResponse {
		validateWebhookAction(rule.Action, path+".action", issues)
	}
	validateCondition(protocol, rule.When, path+".when", issues)
}

func isValidRuleCode(id string) bool {
	if id == "" || len(id) > maxRuleCodeLength {
		return false
	}
	for _, item := range id {
		if item >= 'a' && item <= 'z' || item >= 'A' && item <= 'Z' || item >= '0' && item <= '9' || item == '_' || item == '-' {
			continue
		}
		return false
	}
	return true
}

func requiresTopLevelStatus(actionType string) bool {
	return actionType == eo.ActionTypeStaticResponse ||
		actionType == eo.ActionTypeTemplateResponse ||
		actionType == eo.ActionTypeCELResponse
}

type stringConstraint struct {
	Field  string
	Op     string
	Values []string
	Path   string
}

func validateRuleReachability(ruleSet bo.RuleSet, warnings *[]bo.ValidationIssue) {
	selectorConstraints := collectAllStringConstraints(ruleSet.Selector.All, "selector.all")
	if len(selectorConstraints) == 0 {
		return
	}
	for i, rule := range ruleSet.Rules {
		if !rule.Enabled {
			continue
		}
		rulePath := fmt.Sprintf("rules[%d].when", i)
		ruleConstraints := collectStringConstraints(rule.When, rulePath)
		for _, selectorConstraint := range selectorConstraints {
			for _, ruleConstraint := range ruleConstraints {
				if selectorConstraint.Field != ruleConstraint.Field {
					continue
				}
				if constraintsOverlap(selectorConstraint, ruleConstraint) {
					continue
				}
				*warnings = append(*warnings, bo.ValidationIssue{
					Path: rulePath,
					Message: fmt.Sprintf(
						"rule %q may be unreachable: selector %s conflicts with condition %s",
						rule.ID,
						formatConstraint(selectorConstraint),
						formatConstraint(ruleConstraint),
					),
				})
				goto nextRule
			}
		}
	nextRule:
	}
}

func collectAllStringConstraints(conditions []bo.Condition, path string) []stringConstraint {
	result := make([]stringConstraint, 0)
	for i, condition := range conditions {
		result = append(result, collectStringConstraints(condition, fmt.Sprintf("%s[%d]", path, i))...)
	}
	return result
}

func collectStringConstraints(condition bo.Condition, path string) []stringConstraint {
	if len(condition.All) > 0 {
		return collectAllStringConstraints(condition.All, path+".all")
	}
	if len(condition.Any) > 0 || condition.Not != nil || strings.TrimSpace(condition.Expr) != "" {
		return nil
	}
	switch condition.Op {
	case eo.OperatorEQ, eo.OperatorIn, eo.OperatorPrefix:
	default:
		return nil
	}
	values := normalizeConstraintValues(condition.Field, stringValues(condition.Value))
	if len(values) == 0 {
		return nil
	}
	return []stringConstraint{{
		Field:  condition.Field,
		Op:     condition.Op,
		Values: values,
		Path:   path,
	}}
}

func stringValues(value any) []string {
	switch typed := value.(type) {
	case string:
		return []string{typed}
	case []string:
		return append([]string(nil), typed...)
	case []any:
		values := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				values = append(values, text)
			}
		}
		return values
	default:
		return nil
	}
}

func normalizeConstraintValues(field string, values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		normalized := strings.TrimSpace(value)
		switch field {
		case "request.method":
			normalized = strings.ToUpper(normalized)
		case "request.host", "request.original_host":
			normalized = strings.ToLower(normalized)
		case "request.operation":
			normalized = strings.ToLower(normalized)
		case "request.path":
			if normalized != "" && !strings.HasPrefix(normalized, "/") {
				normalized = "/" + normalized
			}
		}
		if normalized != "" {
			result = append(result, normalized)
		}
	}
	return result
}

func constraintsOverlap(left stringConstraint, right stringConstraint) bool {
	leftSet, leftIsSet := exactConstraintValues(left)
	rightSet, rightIsSet := exactConstraintValues(right)
	if leftIsSet && rightIsSet {
		return stringSetsOverlap(leftSet, rightSet)
	}
	if leftIsSet && right.Op == eo.OperatorPrefix {
		return valuesMatchAnyPrefix(leftSet, right.Values)
	}
	if left.Op == eo.OperatorPrefix && rightIsSet {
		return valuesMatchAnyPrefix(rightSet, left.Values)
	}
	if left.Op == eo.OperatorPrefix && right.Op == eo.OperatorPrefix {
		for _, leftPrefix := range left.Values {
			for _, rightPrefix := range right.Values {
				if strings.HasPrefix(leftPrefix, rightPrefix) || strings.HasPrefix(rightPrefix, leftPrefix) {
					return true
				}
			}
		}
		return false
	}
	return true
}

func exactConstraintValues(constraint stringConstraint) ([]string, bool) {
	switch constraint.Op {
	case eo.OperatorEQ, eo.OperatorIn:
		return constraint.Values, true
	default:
		return nil, false
	}
}

func stringSetsOverlap(left []string, right []string) bool {
	rightSet := make(map[string]struct{}, len(right))
	for _, value := range right {
		rightSet[value] = struct{}{}
	}
	for _, value := range left {
		if _, ok := rightSet[value]; ok {
			return true
		}
	}
	return false
}

func valuesMatchAnyPrefix(values []string, prefixes []string) bool {
	for _, value := range values {
		for _, prefix := range prefixes {
			if strings.HasPrefix(value, prefix) {
				return true
			}
		}
	}
	return false
}

func formatConstraint(constraint stringConstraint) string {
	value := strings.Join(constraint.Values, ", ")
	if len(constraint.Values) > 1 {
		value = "[" + value + "]"
	}
	return fmt.Sprintf("%s %s %q", constraint.Field, constraint.Op, value)
}

func validateSequenceAction(action bo.Action, path string, issues *[]bo.ValidationIssue) {
	if len(action.Sequence) == 0 {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".sequence", Message: "sequence must contain at least one step"})
		return
	}
	if action.SequenceStrategy != "" &&
		action.SequenceStrategy != eo.SequenceStrategyLoop &&
		action.SequenceStrategy != eo.SequenceStrategyLast {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".sequence_strategy", Message: "unsupported sequence strategy"})
	}
	for i, step := range action.Sequence {
		stepPath := fmt.Sprintf("%s.sequence[%d]", path, i)
		if step.Status < 100 || step.Status > 599 {
			*issues = append(*issues, bo.ValidationIssue{Path: stepPath + ".status", Message: "status must be between 100 and 599"})
		}
		validateHeaders(step.Headers, stepPath+".headers", issues)
	}
}

func validateWebhookAction(action bo.Action, path string, issues *[]bo.ValidationIssue) {
	if action.Webhook == nil {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".webhook", Message: "webhook config is required"})
		return
	}
	parsed, err := url.Parse(strings.TrimSpace(action.Webhook.URL))
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".webhook.url", Message: "webhook url must be absolute"})
	} else if parsed.Scheme != "http" && parsed.Scheme != "https" {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".webhook.url", Message: "webhook url must use http or https"})
	}
	if action.Webhook.Method != "" && !isValidHTTPMethod(action.Webhook.Method) {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".webhook.method", Message: "invalid webhook method"})
	}
	if action.Webhook.TimeoutMS < 0 || action.Webhook.TimeoutMS > 30000 {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".webhook.timeout_ms", Message: "timeout_ms must be between 0 and 30000"})
	}
	validateHeaders(action.Webhook.Headers, path+".webhook.headers", issues)
}

func validateHeaders(headers map[string][]string, path string, issues *[]bo.ValidationIssue) {
	for key, values := range headers {
		if !isValidHeaderName(key) {
			*issues = append(*issues, bo.ValidationIssue{Path: path + "." + key, Message: "invalid header name"})
		}
		for i, value := range values {
			if strings.ContainsAny(value, "\r\n") {
				*issues = append(*issues, bo.ValidationIssue{Path: fmt.Sprintf("%s.%s[%d]", path, key, i), Message: "header value must not contain CR or LF"})
			}
		}
	}
}

func isValidHeaderName(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		switch {
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9':
		case strings.ContainsRune("!#$%&'*+-.^_`|~", r):
		default:
			return false
		}
	}
	return true
}

func validateCondition(protocol string, condition bo.Condition, path string, issues *[]bo.ValidationIssue) {
	groupCount := 0
	if len(condition.All) > 0 {
		groupCount++
	}
	if len(condition.Any) > 0 {
		groupCount++
	}
	if condition.Not != nil {
		groupCount++
	}
	if strings.TrimSpace(condition.Expr) != "" {
		groupCount++
	}
	if condition.Field != "" || condition.Op != "" {
		groupCount++
	}
	if groupCount != 1 {
		*issues = append(*issues, bo.ValidationIssue{Path: path, Message: "condition node must define exactly one of all/any/not/expr/predicate"})
		return
	}

	if strings.TrimSpace(condition.Expr) != "" {
		if _, err := compileCELCondition(condition.Expr); err != nil {
			*issues = append(*issues, bo.ValidationIssue{Path: path + ".expr", Message: err.Error()})
		}
	}
	if condition.Field != "" || condition.Op != "" {
		if strings.TrimSpace(condition.Field) == "" {
			*issues = append(*issues, bo.ValidationIssue{Path: path + ".field", Message: "field is required"})
		}
		if !isSupportedOperator(condition.Op) {
			*issues = append(*issues, bo.ValidationIssue{Path: path + ".op", Message: "unsupported operator"})
		}
		if _, err := parsePath(condition.Field); err != nil {
			*issues = append(*issues, bo.ValidationIssue{Path: path + ".field", Message: err.Error()})
		}
		if strings.TrimSpace(protocol) != "" {
			if _, ok := mockprotocol.FieldForPath(protocol, condition.Field); !ok {
				*issues = append(*issues, bo.ValidationIssue{Path: path + ".field", Message: "field is not registered for protocol"})
			} else if !mockprotocol.OperatorAllowed(protocol, condition.Field, condition.Op) {
				*issues = append(*issues, bo.ValidationIssue{Path: path + ".op", Message: "operator is not allowed for field"})
			}
		}
		if condition.Op == eo.OperatorRegex {
			validateRegexCondition(condition, path, issues)
		}
	}
	for i, child := range condition.All {
		validateCondition(protocol, child, fmt.Sprintf("%s.all[%d]", path, i), issues)
	}
	for i, child := range condition.Any {
		validateCondition(protocol, child, fmt.Sprintf("%s.any[%d]", path, i), issues)
	}
	if condition.Not != nil {
		validateCondition(protocol, *condition.Not, path+".not", issues)
	}
}

func validateRegexCondition(condition bo.Condition, path string, issues *[]bo.ValidationIssue) {
	pattern, ok := condition.Value.(string)
	if !ok {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".value", Message: "regex value must be a string"})
		return
	}
	if len(pattern) > maxRegexPatternLength {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".value", Message: fmt.Sprintf("regex pattern must be at most %d characters", maxRegexPatternLength)})
		return
	}
	if _, err := regexp.Compile(pattern); err != nil {
		*issues = append(*issues, bo.ValidationIssue{Path: path + ".value", Message: "invalid regex pattern: " + err.Error()})
	}
}

func isSupportedOperator(operator string) bool {
	switch operator {
	case eo.OperatorEQ, eo.OperatorNE, eo.OperatorIn, eo.OperatorNotIn, eo.OperatorContains,
		eo.OperatorNotContain, eo.OperatorExists, eo.OperatorNotExists, eo.OperatorIsNull,
		eo.OperatorIsNotNull, eo.OperatorPrefix,
		eo.OperatorSuffix, eo.OperatorRegex, eo.OperatorGT, eo.OperatorGTE,
		eo.OperatorLT, eo.OperatorLTE:
		return true
	default:
		return false
	}
}

func specificity(condition bo.Condition) int {
	if len(condition.All) > 0 {
		score := 0
		for _, child := range condition.All {
			score += specificity(child)
		}
		return score
	}
	if len(condition.Any) > 0 {
		score := 0
		for _, child := range condition.Any {
			score += specificity(child)
		}
		return score
	}
	if condition.Not != nil {
		return specificity(*condition.Not)
	}
	if strings.TrimSpace(condition.Expr) != "" {
		return 1
	}

	score := 1
	switch condition.Op {
	case eo.OperatorEQ:
		score += 4
	case eo.OperatorPrefix, eo.OperatorSuffix:
		score += 3
	case eo.OperatorContains, eo.OperatorIn:
		score += 2
	case eo.OperatorRegex:
		score += 1
	}

	if strings.Contains(condition.Field, "headers.") || strings.Contains(condition.Field, "query.") {
		score += 1
	}
	if strings.Contains(condition.Field, "body.") {
		score += 2
	}
	return score
}

func collectConditionExpressions(condition bo.Condition) []string {
	result := make([]string, 0)
	if strings.TrimSpace(condition.Expr) != "" {
		result = append(result, condition.Expr)
	}
	for _, child := range condition.All {
		result = append(result, collectConditionExpressions(child)...)
	}
	for _, child := range condition.Any {
		result = append(result, collectConditionExpressions(child)...)
	}
	if condition.Not != nil {
		result = append(result, collectConditionExpressions(*condition.Not)...)
	}
	return result
}

func indexCompiledRules(compiled *CompiledRuleSet) {
	for index, rule := range compiled.Compiled {
		if rule.ExactMethod == "" {
			compiled.MethodWildcardRules = append(compiled.MethodWildcardRules, index)
		} else {
			compiled.MethodIndex[rule.ExactMethod] = append(compiled.MethodIndex[rule.ExactMethod], index)
		}
		if rule.ExactHost == "" {
			compiled.HostWildcardRules = append(compiled.HostWildcardRules, index)
		} else {
			compiled.HostIndex[rule.ExactHost] = append(compiled.HostIndex[rule.ExactHost], index)
		}
		if rule.ExactPath == "" {
			compiled.PathWildcardRules = append(compiled.PathWildcardRules, index)
		} else {
			compiled.ExactPath[rule.ExactPath] = append(compiled.ExactPath[rule.ExactPath], index)
		}
	}
}

func extractEquality(condition bo.Condition, field string) (string, bool) {
	if len(condition.All) > 0 {
		for _, child := range condition.All {
			if value, ok := extractEquality(child, field); ok {
				return value, true
			}
		}
		return "", false
	}
	if strings.TrimSpace(condition.Expr) != "" {
		return "", false
	}
	if condition.Field == field && condition.Op == eo.OperatorEQ {
		value, ok := condition.Value.(string)
		return value, ok
	}
	return "", false
}

func newRuleTemplate(name, body string) (*template.Template, error) {
	return template.New(name).Funcs(templateFuncMap()).Parse(body)
}
