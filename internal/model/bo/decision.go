package bo

type RuntimeDecision struct {
	Kind     string           `json:"kind"`
	Matched  bool             `json:"matched"`
	Fallback bool             `json:"fallback,omitempty"`
	Trace    MatchTrace       `json:"trace"`
	Response *ActionExecution `json:"response,omitempty"`
	Forward  *ForwardDecision `json:"forward,omitempty"`
	Meta     DecisionMeta     `json:"meta,omitempty"`
}

type ForwardDecision struct {
	TimeoutMS int `json:"timeout_ms,omitempty"`
}

type DecisionMeta struct {
	TraceID string `json:"trace_id,omitempty"`
}
