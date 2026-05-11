package bo

import "time"

type RuleSet struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Enabled   bool     `json:"enabled"`
	Protocol  string   `json:"protocol"`
	Namespace string   `json:"namespace"`
	Selector  Selector `json:"selector"`
	Rules     []Rule   `json:"rules"`
	Version   int      `json:"version"`
}

type Selector struct {
	All []Condition `json:"all,omitempty"`
}

type Rule struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Enabled  bool      `json:"enabled"`
	Priority int       `json:"priority"`
	When     Condition `json:"when"`
	Action   Action    `json:"action"`
}

type Condition struct {
	All   []Condition `json:"all,omitempty"`
	Any   []Condition `json:"any,omitempty"`
	Not   *Condition  `json:"not,omitempty"`
	Expr  string      `json:"expr,omitempty"`
	Field string      `json:"field,omitempty"`
	Op    string      `json:"op,omitempty"`
	Value any         `json:"value,omitempty"`
}

type Action struct {
	Type             string              `json:"type"`
	Status           int                 `json:"status,omitempty"`
	Headers          map[string][]string `json:"headers,omitempty"`
	Body             any                 `json:"body,omitempty"`
	BodyTemplate     string              `json:"body_template,omitempty"`
	BodyExpression   string              `json:"body_expression,omitempty"`
	Sequence         []SequenceStep      `json:"sequence,omitempty"`
	SequenceStrategy string              `json:"sequence_strategy,omitempty"`
	Webhook          *WebhookConfig      `json:"webhook,omitempty"`
}

type SequenceStep struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    any                 `json:"body,omitempty"`
}

type WebhookConfig struct {
	URL       string              `json:"url"`
	Method    string              `json:"method,omitempty"`
	Headers   map[string][]string `json:"headers,omitempty"`
	TimeoutMS int                 `json:"timeout_ms,omitempty"`
}

type ValidationIssue struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Issues   []ValidationIssue `json:"issues,omitempty"`
	Warnings []ValidationIssue `json:"warnings,omitempty"`
}

type MatchTrace struct {
	RulesetID      string `json:"ruleset_id,omitempty"`
	RuleID         string `json:"rule_id,omitempty"`
	FallbackReason string `json:"fallback_reason,omitempty"`
}

type AuditInfo struct {
	Action           string `json:"action,omitempty"`
	Operator         string `json:"operator,omitempty"`
	Reason           string `json:"reason,omitempty"`
	TraceID          string `json:"trace_id,omitempty"`
	SourceSnapshotID string `json:"source_snapshot_id,omitempty"`
}

type ConditionExplanation struct {
	Kind     string                 `json:"kind"`
	Matched  bool                   `json:"matched"`
	Field    string                 `json:"field,omitempty"`
	Operator string                 `json:"operator,omitempty"`
	Expected any                    `json:"expected,omitempty"`
	Actual   []any                  `json:"actual,omitempty"`
	Expr     string                 `json:"expr,omitempty"`
	Message  string                 `json:"message,omitempty"`
	Children []ConditionExplanation `json:"children,omitempty"`
}

type ActionExplanation struct {
	Type           string `json:"type"`
	Template       string `json:"template,omitempty"`
	Expression     string `json:"expression,omitempty"`
	RenderedResult any    `json:"rendered_result,omitempty"`
	Message        string `json:"message,omitempty"`
}

type RuleExplanation struct {
	RuleID     string               `json:"rule_id"`
	Priority   int                  `json:"priority"`
	Matched    bool                 `json:"matched"`
	Condition  ConditionExplanation `json:"condition"`
	ActionInfo *ActionExplanation   `json:"action_info,omitempty"`
}

type SelectorCheckExplanation struct {
	Name     string `json:"name"`
	Matched  bool   `json:"matched"`
	Expected any    `json:"expected,omitempty"`
	Actual   any    `json:"actual,omitempty"`
	Message  string `json:"message,omitempty"`
}

type RuleSetExplanation struct {
	RuleSetID        string                     `json:"ruleset_id"`
	Matched          bool                       `json:"matched"`
	Message          string                     `json:"message,omitempty"`
	SelectorChecks   []SelectorCheckExplanation `json:"selector_checks,omitempty"`
	CandidateRules   []string                   `json:"candidate_rules,omitempty"`
	RuleExplanations []RuleExplanation          `json:"rule_explanations,omitempty"`
}

type MatchExplanation struct {
	RuleSetID           string               `json:"ruleset_id,omitempty"`
	CandidateRules      []string             `json:"candidate_rules,omitempty"`
	RuleExplanations    []RuleExplanation    `json:"rule_explanations,omitempty"`
	RuleSetExplanations []RuleSetExplanation `json:"rule_set_explanations,omitempty"`
}

type RuleDiffSummary struct {
	RuleID           string             `json:"rule_id"`
	ChangeType       string             `json:"change_type"`
	ConditionChanged bool               `json:"condition_changed,omitempty"`
	ActionChanged    bool               `json:"action_changed,omitempty"`
	Message          string             `json:"message,omitempty"`
	FieldDiffs       []FieldDiffSummary `json:"field_diffs,omitempty"`
}

type FieldDiffSummary struct {
	Path    string `json:"path"`
	Message string `json:"message,omitempty"`
	Current any    `json:"current,omitempty"`
	Target  any    `json:"target,omitempty"`
}

type RollbackDiffSummary struct {
	CurrentSnapshotID string             `json:"current_snapshot_id,omitempty"`
	TargetSnapshotID  string             `json:"target_snapshot_id,omitempty"`
	Changed           bool               `json:"changed"`
	Messages          []string           `json:"messages,omitempty"`
	RuleSetFieldDiffs []FieldDiffSummary `json:"ruleset_field_diffs,omitempty"`
	RuleDiffs         []RuleDiffSummary  `json:"rule_diffs,omitempty"`
}

type ActionExecution struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    any                 `json:"body,omitempty"`
}

type SimulationResult struct {
	Matched    bool             `json:"matched"`
	Fallback   bool             `json:"fallback,omitempty"`
	Trace      MatchTrace       `json:"trace"`
	Candidates []string         `json:"candidates,omitempty"`
	Response   ActionExecution  `json:"response,omitempty"`
	Explain    MatchExplanation `json:"explain,omitempty"`
}

type PublishedRuleSetSnapshot struct {
	SnapshotID  string     `json:"snapshot_id"`
	PublishedAt time.Time  `json:"published_at"`
	RuleSet     RuleSet    `json:"ruleset"`
	Audit       *AuditInfo `json:"audit,omitempty"`
}

type RollbackPreviewResult struct {
	Snapshot   PublishedRuleSetSnapshot `json:"snapshot"`
	Valid      bool                     `json:"valid"`
	Validation ValidationResult         `json:"validation"`
	Diff       RollbackDiffSummary      `json:"diff"`
	Simulation *SimulationResult        `json:"simulation,omitempty"`
}
