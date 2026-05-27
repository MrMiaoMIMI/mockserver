package bo

const (
	ScenarioStatusActive = "active"
)

type Scenario struct {
	DBID        uint64 `json:"-"`
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status"`
	ExpireTime  uint64 `json:"expire_time"`
	Version     int    `json:"version"`
}

type ScenarioQuery struct {
	Status         string
	IncludeExpired bool
	Now            uint64
	Limit          int
	Offset         int
}

type ScenarioUpdate struct {
	Name        *string
	Description *string
	ExpireTime  *uint64
}

type ScenarioRule struct {
	DBID         uint64 `json:"-"`
	ScenarioID   string `json:"scenario_id"`
	ScenarioDBID uint64 `json:"-"`
	RuleID       string `json:"rule_id"`
	Name         string `json:"name,omitempty"`
	Protocol     string `json:"protocol"`
	Namespace    string `json:"namespace"`
	Enabled      bool   `json:"enabled"`
	Priority     int    `json:"priority"`
	ExpireTime   uint64 `json:"expire_time"`
	Rule         Rule   `json:"rule"`
	Version      int    `json:"version"`
}

type ScenarioRuleQuery struct {
	ScenarioID  string
	Protocol    string
	Namespace   string
	Now         uint64
	EnabledOnly bool
}

type HTTPQuickRule struct {
	RuleID    string            `json:"rule_id,omitempty"`
	Name      string            `json:"name,omitempty"`
	Namespace string            `json:"namespace,omitempty"`
	Enabled   *bool             `json:"enabled,omitempty"`
	Priority  int               `json:"priority,omitempty"`
	Match     HTTPQuickMatch    `json:"match"`
	Respond   HTTPQuickResponse `json:"respond"`
}

type HTTPQuickMatch struct {
	Method  string         `json:"method,omitempty"`
	Host    string         `json:"host,omitempty"`
	Path    string         `json:"path,omitempty"`
	Query   map[string]any `json:"query,omitempty"`
	Headers map[string]any `json:"headers,omitempty"`
	Body    map[string]any `json:"body,omitempty"`
}

type HTTPQuickResponse struct {
	Status  int                 `json:"status,omitempty"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    any                 `json:"body,omitempty"`
}
