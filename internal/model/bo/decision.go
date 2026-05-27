package bo

type RuntimeDecision struct {
	Kind     string            `json:"kind"`
	Matched  bool              `json:"matched"`
	Fallback bool              `json:"fallback,omitempty"`
	Protocol string            `json:"protocol,omitempty"`
	Trace    MatchTrace        `json:"trace"`
	Response *ProtocolResponse `json:"response,omitempty"`
	Forward  *ForwardDecision  `json:"forward,omitempty"`
	Meta     DecisionMeta      `json:"meta,omitempty"`

	Diagnostics *DecisionDiagnostics `json:"-"`
}

type ForwardDecision struct {
	TimeoutMS int `json:"timeout_ms,omitempty"`
}

type DecisionMeta struct {
	TraceID    string `json:"trace_id,omitempty"`
	ScenarioID string `json:"scenario_id,omitempty"`
}

type DecisionDiagnostics struct {
	RuleSetSelection RuleSetSelectionDiagnostics `json:"ruleset_selection,omitempty"`
	RuleSelection    RuleSelectionDiagnostics    `json:"rule_selection,omitempty"`
}

type RuleSetSelectionDiagnostics struct {
	WinnerRuleSetID  string             `json:"winner_ruleset_id,omitempty"`
	WinnerSnapshotID string             `json:"winner_snapshot_id,omitempty"`
	Candidates       []RuleSetCandidate `json:"candidates,omitempty"`
}

type RuleSelectionDiagnostics struct {
	CandidateRuleIDs []string `json:"candidate_rule_ids,omitempty"`
	WinnerRuleID     string   `json:"winner_rule_id,omitempty"`
}
