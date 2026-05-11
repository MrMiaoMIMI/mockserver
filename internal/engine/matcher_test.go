package engine

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/MrMiaoMIMI/mockserver/internal/model/bo"
	"github.com/MrMiaoMIMI/mockserver/internal/model/eo"
)

func TestCompileAndMatchTemplateRule(t *testing.T) {
	ruleSet := bo.RuleSet{
		ID:        "http-default",
		Name:      "http default",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  httpSelector(httpHostSelector("demo.com"), httpPathPrefixSelector("/api/")),
		Rules: []bo.Rule{
			{
				ID:       "debug-api",
				Name:     "Debug API response",
				Enabled:  true,
				Priority: 100,
				When: bo.Condition{
					All: []bo.Condition{
						{Field: "request.method", Op: eo.OperatorEQ, Value: "GET"},
						{Field: "request.path", Op: eo.OperatorEQ, Value: "/api/v1/debug"},
						{Field: "request.query.q1[*]", Op: eo.OperatorContains, Value: "qv1"},
					},
				},
				Action: bo.Action{
					Type:         eo.ActionTypeTemplateResponse,
					Status:       200,
					Headers:      map[string][]string{"content-type": {"application/json"}},
					BodyTemplate: `{"message":"hello {{ query . "q1" }}","path":"{{ field . "request.path" }}"}`,
				},
			},
		},
	}

	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	event := bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"host":   "demo.com",
			"path":   "/api/v1/debug",
			"query":  map[string][]string{"q1": {"qv1"}},
		},
	}

	result, err := Match([]CompiledRuleSet{compiled}, event)
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if !result.Matched {
		t.Fatalf("expected result to match")
	}
	if result.Trace.RuleID != "debug-api" {
		t.Fatalf("unexpected rule id: %s", result.Trace.RuleID)
	}
	if result.Response.Status != 200 {
		t.Fatalf("unexpected status: %d", result.Response.Status)
	}
	body, ok := result.Response.Body.(string)
	if !ok || body != `{"message":"hello qv1","path":"/api/v1/debug"}` {
		t.Fatalf("unexpected body: %#v", result.Response.Body)
	}
}

func TestCompileAndMatchSequenceResponseRule(t *testing.T) {
	ruleSet := testSingleActionRuleSet("sequence-case", bo.Action{
		Type:             eo.ActionTypeSequenceResponse,
		SequenceStrategy: eo.SequenceStrategyLast,
		Sequence: []bo.SequenceStep{
			{Status: 202, Body: map[string]any{"state": "pending"}},
			{Status: 200, Body: map[string]any{"state": "done"}},
		},
	})
	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	first, err := Match([]CompiledRuleSet{compiled}, testHTTPEvent("/api/v1/action"))
	if err != nil {
		t.Fatalf("first Match() error = %v", err)
	}
	second, err := Match([]CompiledRuleSet{compiled}, testHTTPEvent("/api/v1/action"))
	if err != nil {
		t.Fatalf("second Match() error = %v", err)
	}
	third, err := Match([]CompiledRuleSet{compiled}, testHTTPEvent("/api/v1/action"))
	if err != nil {
		t.Fatalf("third Match() error = %v", err)
	}

	if first.Response.Status != 202 || first.Response.Body.(map[string]any)["state"] != "pending" {
		t.Fatalf("unexpected first response: %#v", first.Response)
	}
	if second.Response.Status != 200 || second.Response.Body.(map[string]any)["state"] != "done" {
		t.Fatalf("unexpected second response: %#v", second.Response)
	}
	if third.Response.Status != 200 || third.Response.Body.(map[string]any)["state"] != "done" {
		t.Fatalf("unexpected third response: %#v", third.Response)
	}
}

func TestCompileAndMatchWebhookResponseRule(t *testing.T) {
	oldFactory := newWebhookHTTPClient
	defer func() { newWebhookHTTPClient = oldFactory }()
	newWebhookHTTPClient = func(timeout time.Duration) *http.Client {
		return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected webhook method: %s", r.Method)
			}
			return &http.Response{
				StatusCode: http.StatusCreated,
				Header: http.Header{
					"Content-Type":     {"application/json"},
					"X-Webhook-Result": {"ok"},
				},
				Body: io.NopCloser(strings.NewReader(`{"from_webhook":true}`)),
			}, nil
		})}
	}

	ruleSet := testSingleActionRuleSet("webhook-case", bo.Action{
		Type: eo.ActionTypeWebhookResponse,
		Webhook: &bo.WebhookConfig{
			URL:       "http://webhook.example/mock",
			Method:    http.MethodPost,
			TimeoutMS: 1000,
		},
	})
	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	result, err := Match([]CompiledRuleSet{compiled}, testHTTPEvent("/api/v1/action"))
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if result.Response.Status != http.StatusCreated {
		t.Fatalf("unexpected status: %d", result.Response.Status)
	}
	if values := result.Response.Headers["X-Webhook-Result"]; len(values) != 1 || values[0] != "ok" {
		t.Fatalf("unexpected webhook header: %#v", result.Response.Headers)
	}
	body, ok := result.Response.Body.(map[string]any)
	if !ok || body["from_webhook"] != true {
		t.Fatalf("unexpected body: %#v", result.Response.Body)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func httpSelector(conditions ...bo.Condition) bo.Selector {
	return bo.Selector{All: conditions}
}

func httpHostSelector(host string) bo.Condition {
	return bo.Condition{Field: "request.host", Op: eo.OperatorEQ, Value: host}
}

func httpPathPrefixSelector(prefix string) bo.Condition {
	return bo.Condition{Field: "request.path", Op: eo.OperatorPrefix, Value: prefix}
}

func testSingleActionRuleSet(id string, action bo.Action) bo.RuleSet {
	return bo.RuleSet{
		ID:        id,
		Name:      id,
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  httpSelector(httpPathPrefixSelector("/api/")),
		Rules: []bo.Rule{
			{
				ID:       "action-rule",
				Name:     "Action response",
				Enabled:  true,
				Priority: 100,
				When: bo.Condition{
					All: []bo.Condition{
						{Field: "request.method", Op: eo.OperatorEQ, Value: "GET"},
						{Field: "request.path", Op: eo.OperatorEQ, Value: "/api/v1/action"},
					},
				},
				Action: action,
			},
		},
	}
}

func testHTTPEvent(path string) bo.Event {
	return bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"host":   "demo.com",
			"path":   path,
		},
	}
}

func TestCompileAndMatchTemplateRuleWithBodyHelpers(t *testing.T) {
	ruleSet := bo.RuleSet{
		ID:        "http-template-body",
		Name:      "http template body",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  httpSelector(httpPathPrefixSelector("/api/")),
		Rules: []bo.Rule{
			{
				ID:       "template-body",
				Name:     "Template body response",
				Enabled:  true,
				Priority: 100,
				When: bo.Condition{
					All: []bo.Condition{
						{Field: "request.path", Op: eo.OperatorEQ, Value: "/api/v1/template"},
						{Field: "request.body.user.id", Op: eo.OperatorEQ, Value: "u-1"},
					},
				},
				Action: bo.Action{
					Type:         eo.ActionTypeTemplateResponse,
					Status:       200,
					BodyTemplate: `{"user":"{{ body . "user.id" }}","tag":"{{ first (queryAll . "tag") }}","payload":{{ toJSON (body . "user") }}}`,
				},
			},
		},
	}

	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	event := bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "POST",
			"host":   "demo.com",
			"path":   "/api/v1/template",
			"query":  map[string][]string{"tag": {"t1", "t2"}},
			"body": map[string]any{
				"user": map[string]any{
					"id":   "u-1",
					"name": "tester",
				},
			},
		},
	}

	result, err := Match([]CompiledRuleSet{compiled}, event)
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	body, ok := result.Response.Body.(string)
	if !ok {
		t.Fatalf("unexpected body type: %T", result.Response.Body)
	}
	if body != `{"user":"u-1","tag":"t1","payload":{"id":"u-1","name":"tester"}}` {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestCompileAndMatchCELRule(t *testing.T) {
	ruleSet := bo.RuleSet{
		ID:        "http-cel",
		Name:      "http cel",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  httpSelector(httpPathPrefixSelector("/api/")),
		Rules: []bo.Rule{
			{
				ID:       "cel-rule",
				Name:     "CEL score response",
				Enabled:  true,
				Priority: 100,
				When: bo.Condition{
					All: []bo.Condition{
						{Field: "request.body.user.id", Op: eo.OperatorEQ, Value: "u-1"},
						{Expr: `request.headers["x-env"][0] == "test" && request.body.score >= 90`},
					},
				},
				Action: bo.Action{
					Type:           eo.ActionTypeCELResponse,
					Status:         201,
					Headers:        map[string][]string{"content-type": {"application/json"}},
					BodyExpression: `{"user_id": request.body.user.id, "score": request.body.score, "path": request.path}`,
				},
			},
		},
	}

	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	event := bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "POST",
			"host":   "demo.com",
			"path":   "/api/v1/cel",
			"headers": map[string][]string{
				"x-env": {"test"},
			},
			"body": map[string]any{
				"user": map[string]any{
					"id": "u-1",
				},
				"score": 95,
			},
		},
	}

	result, err := Match([]CompiledRuleSet{compiled}, event)
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if !result.Matched {
		t.Fatalf("expected result to match")
	}
	if result.Response.Status != 201 {
		t.Fatalf("unexpected status: %d", result.Response.Status)
	}

	body, ok := result.Response.Body.(map[string]any)
	if !ok {
		t.Fatalf("unexpected body type: %T", result.Response.Body)
	}
	if body["user_id"] != "u-1" {
		t.Fatalf("unexpected user_id: %#v", body["user_id"])
	}
	if body["path"] != "/api/v1/cel" {
		t.Fatalf("unexpected path: %#v", body["path"])
	}
}

func TestMatchUsesCompiledRuleIndexesWithoutDroppingWildcardRules(t *testing.T) {
	ruleSet := bo.RuleSet{
		ID:        "indexed-case",
		Name:      "indexed case",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  httpSelector(httpPathPrefixSelector("/api/")),
		Rules: []bo.Rule{
			{
				ID:       "wrong-exact-path",
				Name:     "Wrong exact path response",
				Enabled:  true,
				Priority: 200,
				When: bo.Condition{
					All: []bo.Condition{
						{Field: "request.method", Op: eo.OperatorEQ, Value: "GET"},
						{Field: "request.path", Op: eo.OperatorEQ, Value: "/api/v1/wrong"},
					},
				},
				Action: bo.Action{
					Type:   eo.ActionTypeStaticResponse,
					Status: 200,
					Body:   map[string]any{"rule": "wrong"},
				},
			},
			{
				ID:       "wildcard-path",
				Name:     "Wildcard path response",
				Enabled:  true,
				Priority: 100,
				When: bo.Condition{
					Field: "request.method",
					Op:    eo.OperatorEQ,
					Value: "GET",
				},
				Action: bo.Action{
					Type:   eo.ActionTypeStaticResponse,
					Status: 200,
					Body:   map[string]any{"rule": "wildcard"},
				},
			},
		},
	}

	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	result, err := Match([]CompiledRuleSet{compiled}, bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"path":   "/api/v1/actual",
		},
	})
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if !result.Matched {
		t.Fatalf("expected result to match wildcard rule")
	}
	if result.Trace.RuleID != "wildcard-path" {
		t.Fatalf("unexpected matched rule: %s", result.Trace.RuleID)
	}
	if len(result.Candidates) != 1 || result.Candidates[0] != "wildcard-path" {
		t.Fatalf("unexpected indexed candidates: %#v", result.Candidates)
	}
}

func TestHTTPPathPrefixSelectorMatchesPathSegments(t *testing.T) {
	ruleSet := bo.RuleSet{
		ID:        "path-prefix-case",
		Name:      "path prefix case",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  httpSelector(httpPathPrefixSelector("/api")),
		Rules: []bo.Rule{
			{
				ID:       "method-only-rule",
				Name:     "Method only response",
				Enabled:  true,
				Priority: 100,
				When: bo.Condition{
					Field: "request.method",
					Op:    eo.OperatorEQ,
					Value: "GET",
				},
				Action: bo.Action{
					Type:   eo.ActionTypeStaticResponse,
					Status: 200,
					Body:   map[string]any{"ok": true},
				},
			},
		},
	}

	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	matchedSegment, err := Match([]CompiledRuleSet{compiled}, bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"path":   "/api/v1/debug",
		},
	})
	if err != nil {
		t.Fatalf("Match() segment path error = %v", err)
	}
	if !matchedSegment.Matched {
		t.Fatalf("expected /api/v1/debug to match /api prefix")
	}

	siblingSegment, err := Match([]CompiledRuleSet{compiled}, bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"path":   "/api2/v1/debug",
		},
	})
	if err != nil {
		t.Fatalf("Match() sibling path error = %v", err)
	}
	if siblingSegment.Matched {
		t.Fatalf("expected /api2/v1/debug not to match /api prefix")
	}
	if !hasSelectorCheck(siblingSegment.Explain, "request.path", false) {
		t.Fatalf("expected request.path selector check to miss, got %#v", siblingSegment.Explain.RuleSetExplanations)
	}
}

func TestMatchExplainsIndexedOutRulesWhenRuleSetMatches(t *testing.T) {
	ruleSet := bo.RuleSet{
		ID:        "rule-miss-case",
		Name:      "rule miss case",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  httpSelector(httpPathPrefixSelector("/api")),
		Rules: []bo.Rule{
			{
				ID:       "debug-v2-rule",
				Name:     "Debug v2 response",
				Enabled:  true,
				Priority: 100,
				When: bo.Condition{
					All: []bo.Condition{
						{Field: "request.method", Op: eo.OperatorEQ, Value: "GET"},
						{Field: "request.path", Op: eo.OperatorEQ, Value: "/api/v2/debug"},
					},
				},
				Action: bo.Action{
					Type:   eo.ActionTypeStaticResponse,
					Status: 200,
					Body:   map[string]any{"ok": true},
				},
			},
		},
	}

	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	result, err := Match([]CompiledRuleSet{compiled}, bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"path":   "/api/v3/debug",
		},
	})
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if result.Matched {
		t.Fatalf("expected rule miss")
	}
	if len(result.Candidates) != 0 {
		t.Fatalf("expected no indexed candidates, got %#v", result.Candidates)
	}
	if !hasRuleConditionIssue(result.Explain, "debug-v2-rule", "request.path", "/api/v2/debug", "/api/v3/debug") {
		t.Fatalf("expected indexed-out rule explanation, got %#v", result.Explain.RuleExplanations)
	}
}

func TestValidateRuleSetRejectsInvalidCEL(t *testing.T) {
	ruleSet := bo.RuleSet{
		ID:        "invalid-cel",
		Name:      "invalid cel",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Rules: []bo.Rule{
			{
				ID:       "broken",
				Name:     "Broken CEL response",
				Enabled:  true,
				Priority: 1,
				When: bo.Condition{
					Expr: `request.path == `,
				},
				Action: bo.Action{
					Type:   eo.ActionTypeStaticResponse,
					Status: 200,
					Body:   map[string]any{"ok": true},
				},
			},
		},
	}

	validation := ValidateRuleSet(ruleSet)
	if validation.Valid {
		t.Fatalf("expected validation to fail")
	}
}

func TestValidateRuleSetRejectsUnsafeOrAmbiguousRules(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(*bo.RuleSet)
		wantPath   string
		wantReason string
	}{
		{
			name: "empty rules",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules = nil
			},
			wantPath:   "rules",
			wantReason: "at least one rule",
		},
		{
			name: "missing rule name",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules[0].Name = ""
			},
			wantPath:   "rules[0].name",
			wantReason: "rule name is required",
		},
		{
			name: "duplicate rule id",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules = append(ruleSet.Rules, ruleSet.Rules[0])
			},
			wantPath:   "rules[1].id",
			wantReason: "duplicate rule id",
		},
		{
			name: "invalid status",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules[0].Action.Status = 99
			},
			wantPath:   "rules[0].action.status",
			wantReason: "between 100 and 599",
		},
		{
			name: "invalid header name",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules[0].Action.Headers = map[string][]string{"bad header": {"value"}}
			},
			wantPath:   "rules[0].action.headers.bad header",
			wantReason: "invalid header name",
		},
		{
			name: "unsafe header value",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules[0].Action.Headers = map[string][]string{"x-good": {"bad\r\nvalue"}}
			},
			wantPath:   "rules[0].action.headers.x-good[0]",
			wantReason: "must not contain CR or LF",
		},
		{
			name: "invalid selector field",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Selector.All = []bo.Condition{{Field: "request.body.status", Op: eo.OperatorEQ, Value: "doing"}}
			},
			wantPath:   "selector.all[0].field",
			wantReason: "field is not registered as selector",
		},
		{
			name: "invalid selector operator",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Selector.All = []bo.Condition{{Field: "request.path", Op: eo.OperatorRegex, Value: "^/api/"}}
			},
			wantPath:   "selector.all[0].op",
			wantReason: "operator is not allowed for selector",
		},
		{
			name: "invalid regex",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules[0].When = bo.Condition{Field: "request.path", Op: eo.OperatorRegex, Value: "["}
			},
			wantPath:   "rules[0].when.value",
			wantReason: "invalid regex pattern",
		},
		{
			name: "regex too long",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules[0].When = bo.Condition{Field: "request.path", Op: eo.OperatorRegex, Value: strings.Repeat("a", maxRegexPatternLength+1)}
			},
			wantPath:   "rules[0].when.value",
			wantReason: "at most",
		},
		{
			name: "unknown protocol field",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules[0].When = bo.Condition{Field: "request.cache_key", Op: eo.OperatorEQ, Value: "user:1"}
			},
			wantPath:   "rules[0].when.field",
			wantReason: "field is not registered",
		},
		{
			name: "operator not allowed by field",
			mutate: func(ruleSet *bo.RuleSet) {
				ruleSet.Rules[0].When = bo.Condition{Field: "request.method", Op: eo.OperatorRegex, Value: "GET|POST"}
			},
			wantPath:   "rules[0].when.op",
			wantReason: "operator is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ruleSet := validRuleSetForValidation()
			tt.mutate(&ruleSet)
			validation := ValidateRuleSet(ruleSet)
			if validation.Valid {
				t.Fatalf("expected validation to fail")
			}
			if !hasValidationIssue(validation, tt.wantPath, tt.wantReason) {
				t.Fatalf("expected issue path=%q reason containing=%q, got %#v", tt.wantPath, tt.wantReason, validation.Issues)
			}
		})
	}
}

func TestValidateRuleSetAllowsDynamicJSONOperators(t *testing.T) {
	ruleSet := validRuleSetForValidation()
	ruleSet.Rules[0].When = bo.Condition{
		All: []bo.Condition{
			{Field: "request.body.score", Op: eo.OperatorGTE, Value: 90},
			{Field: "request.headers.x-env[0]", Op: eo.OperatorRegex, Value: "test|beta"},
		},
	}
	validation := ValidateRuleSet(ruleSet)
	if !validation.Valid {
		t.Fatalf("expected validation to pass, got %#v", validation.Issues)
	}
}

func TestValidateRuleSetWarnsUnreachableRule(t *testing.T) {
	ruleSet := validRuleSetForValidation()
	ruleSet.Selector = httpSelector(bo.Condition{
		Field: "request.path",
		Op:    eo.OperatorEQ,
		Value: "/api",
	})
	ruleSet.Rules[0].When = bo.Condition{
		Field: "request.path",
		Op:    eo.OperatorEQ,
		Value: "/api/v1/validation",
	}

	validation := ValidateRuleSet(ruleSet)
	if !validation.Valid {
		t.Fatalf("expected validation to pass, got %#v", validation.Issues)
	}
	if len(validation.Warnings) != 1 {
		t.Fatalf("expected one reachability warning, got %#v", validation.Warnings)
	}
	if validation.Warnings[0].Path != "rules[0].when" || !strings.Contains(validation.Warnings[0].Message, "may be unreachable") {
		t.Fatalf("unexpected warning: %#v", validation.Warnings[0])
	}
}

func TestMatchNullOperators(t *testing.T) {
	ruleSet := validRuleSetForValidation()
	ruleSet.Rules[0].When = bo.Condition{
		All: []bo.Condition{
			{Field: "request.body.result_list", Op: eo.OperatorIsNull},
			{Field: "request.body.status", Op: eo.OperatorIsNotNull},
			{Field: "request.headers.x-missing[0]", Op: eo.OperatorIsNull},
		},
	}
	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	event := testHTTPEvent("/api/v1/validation")
	event.Request["body"] = map[string]any{
		"result_list": nil,
		"status":      "doing",
	}
	result, err := Match([]CompiledRuleSet{compiled}, event)
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if !result.Matched {
		t.Fatalf("expected null-aware rule to match")
	}
}

func validRuleSetForValidation() bo.RuleSet {
	return bo.RuleSet{
		ID:        "validation-case",
		Name:      "validation case",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  httpSelector(httpHostSelector("demo.com"), httpPathPrefixSelector("/api/")),
		Rules: []bo.Rule{
			{
				ID:       "validation-rule",
				Name:     "Validation response",
				Enabled:  true,
				Priority: 100,
				When: bo.Condition{
					Field: "request.path",
					Op:    eo.OperatorEQ,
					Value: "/api/v1/validation",
				},
				Action: bo.Action{
					Type:   eo.ActionTypeStaticResponse,
					Status: 200,
					Body:   map[string]any{"ok": true},
				},
			},
		},
	}
}

func hasValidationIssue(validation bo.ValidationResult, path string, reason string) bool {
	for _, issue := range validation.Issues {
		if issue.Path == path && strings.Contains(issue.Message, reason) {
			return true
		}
	}
	return false
}

func hasSelectorCheck(explain bo.MatchExplanation, name string, matched bool) bool {
	for _, ruleSet := range explain.RuleSetExplanations {
		for _, check := range ruleSet.SelectorChecks {
			if check.Name == name && check.Matched == matched {
				return true
			}
		}
	}
	return false
}

func hasRuleConditionIssue(explain bo.MatchExplanation, ruleID string, field string, expected any, actual any) bool {
	for _, rule := range explain.RuleExplanations {
		if rule.RuleID != ruleID {
			continue
		}
		if conditionExplanationContains(rule.Condition, field, expected, actual) {
			return true
		}
	}
	return false
}

func conditionExplanationContains(explanation bo.ConditionExplanation, field string, expected any, actual any) bool {
	if explanation.Field == field && explanation.Expected == expected && anyValue(explanation.Actual, func(value any) bool {
		return value == actual
	}) {
		return true
	}
	for _, child := range explanation.Children {
		if conditionExplanationContains(child, field, expected, actual) {
			return true
		}
	}
	return false
}

func TestMatchWithOptionsTrimExplainDepth(t *testing.T) {
	ruleSet := bo.RuleSet{
		ID:        "depth-case",
		Name:      "depth case",
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Rules: []bo.Rule{
			{
				ID:       "nested-rule",
				Name:     "Nested condition response",
				Enabled:  true,
				Priority: 1,
				When: bo.Condition{
					All: []bo.Condition{
						{
							Any: []bo.Condition{
								{Field: "request.method", Op: eo.OperatorEQ, Value: "GET"},
							},
						},
					},
				},
				Action: bo.Action{
					Type:   eo.ActionTypeStaticResponse,
					Status: 200,
					Body:   map[string]any{"ok": true},
				},
			},
		},
	}

	compiled, err := CompileRuleSet(ruleSet)
	if err != nil {
		t.Fatalf("CompileRuleSet() error = %v", err)
	}

	result, err := MatchWithOptions([]CompiledRuleSet{compiled}, bo.Event{
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Request: bo.EventRequest{
			"method": "GET",
			"path":   "/",
		},
	}, MatchOptions{
		ExplainMaxDepth: 1,
	})
	if err != nil {
		t.Fatalf("MatchWithOptions() error = %v", err)
	}

	if len(result.Explain.RuleExplanations) == 0 {
		t.Fatalf("expected rule explanations")
	}
	root := result.Explain.RuleExplanations[0].Condition
	if len(root.Children) != 1 {
		t.Fatalf("expected one first-level child, got %d", len(root.Children))
	}
	if len(root.Children[0].Children) != 0 {
		t.Fatalf("expected nested children to be trimmed at max depth")
	}
}

func TestMatchSelectsMostSpecificRuleSetBySelector(t *testing.T) {
	event := testHTTPEvent("/api/order/status")
	broad := testStaticRuleSetWithSelector("z-broad", httpSelector(httpPathPrefixSelector("/api/")))
	specific := testStaticRuleSetWithSelector("a-specific", httpSelector(httpPathPrefixSelector("/api/order/")))

	result := mustMatchRuleSets(t, event, broad, specific)
	if result.Trace.RulesetID != "a-specific" {
		t.Fatalf("expected more specific path selector to win, got %s", result.Trace.RulesetID)
	}
}

func TestMatchSelectorHostSpecificityWinsBeforePath(t *testing.T) {
	event := testHTTPEvent("/api/order/status")
	hostSpecific := testStaticRuleSetWithSelector("a-host", httpSelector(httpHostSelector("demo.com"), httpPathPrefixSelector("/api/")))
	pathSpecific := testStaticRuleSetWithSelector("z-path", httpSelector(httpPathPrefixSelector("/api/order/status")))

	result := mustMatchRuleSets(t, event, pathSpecific, hostSpecific)
	if result.Trace.RulesetID != "a-host" {
		t.Fatalf("expected host selector to win before path selector, got %s", result.Trace.RulesetID)
	}
}

func TestMatchSelectorTieBreaksByRuleSetIDDesc(t *testing.T) {
	event := testHTTPEvent("/api/order/status")
	left := testStaticRuleSetWithSelector("a-ruleset", httpSelector(httpPathPrefixSelector("/api/")))
	right := testStaticRuleSetWithSelector("z-ruleset", httpSelector(httpPathPrefixSelector("/api/")))

	result := mustMatchRuleSets(t, event, left, right)
	if result.Trace.RulesetID != "z-ruleset" {
		t.Fatalf("expected ruleset id desc tiebreaker to win, got %s", result.Trace.RulesetID)
	}
}

func TestMatchDoesNotFallbackAfterMostSpecificRuleSetSelected(t *testing.T) {
	event := testHTTPEvent("/api/order/status")
	broad := testStaticRuleSetWithSelector("broad", httpSelector(httpPathPrefixSelector("/api/")))
	specific := testStaticRuleSetWithSelector("specific", httpSelector(httpPathPrefixSelector("/api/order/")))
	specific.Rules[0].When = bo.Condition{Field: "request.path", Op: eo.OperatorEQ, Value: "/api/order/not-match"}

	result := mustMatchRuleSets(t, event, broad, specific)
	if result.Matched {
		t.Fatalf("expected no fallback after the most specific ruleset has no matching rule, got %#v", result.Trace)
	}
}

func mustMatchRuleSets(t *testing.T, event bo.Event, ruleSets ...bo.RuleSet) bo.SimulationResult {
	t.Helper()
	compiled := make([]CompiledRuleSet, 0, len(ruleSets))
	for _, ruleSet := range ruleSets {
		item, err := CompileRuleSet(ruleSet)
		if err != nil {
			t.Fatalf("CompileRuleSet(%s) error = %v", ruleSet.ID, err)
		}
		compiled = append(compiled, item)
	}
	result, err := Match(compiled, event)
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	return result
}

func testStaticRuleSetWithSelector(id string, selector bo.Selector) bo.RuleSet {
	return bo.RuleSet{
		ID:        id,
		Name:      id,
		Enabled:   true,
		Protocol:  eo.ProtocolHTTP,
		Namespace: "default",
		Selector:  selector,
		Rules: []bo.Rule{
			{
				ID:       "match",
				Name:     "Selector match response",
				Enabled:  true,
				Priority: 100,
				When: bo.Condition{
					All: []bo.Condition{
						{Field: "request.method", Op: eo.OperatorEQ, Value: "GET"},
						{Field: "request.path", Op: eo.OperatorEQ, Value: "/api/order/status"},
					},
				},
				Action: bo.Action{
					Type:   eo.ActionTypeStaticResponse,
					Status: 200,
					Body:   map[string]any{"ruleset": id},
				},
			},
		},
	}
}
