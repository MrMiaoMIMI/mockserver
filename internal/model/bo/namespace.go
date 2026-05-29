package bo

type Namespace struct {
	DBID        uint64                     `json:"-"`
	Name        string                     `json:"name"`
	Description string                     `json:"description,omitempty"`
	Policies    map[string]NamespacePolicy `json:"policies"`
	Version     int                        `json:"version"`
}

type NamespacePolicy struct {
	RulesetMissAction Action `json:"ruleset_miss_action"`
	RuleMissAction    Action `json:"rule_miss_action"`
}

type NamespaceForwardFallback struct {
	TimeoutMS int `json:"timeout_ms,omitempty"`
}

const DefaultNamespaceForwardTimeoutMS = 5000

func DefaultNamespaceForwardAction() Action {
	return Action{
		Type:    "forward",
		Forward: &NamespaceForwardFallback{TimeoutMS: DefaultNamespaceForwardTimeoutMS},
	}
}

func DefaultNamespace(id string) Namespace {
	return Namespace{
		Name:    id,
		Version: 1,
		Policies: map[string]NamespacePolicy{
			"http": {
				RulesetMissAction: DefaultNamespaceForwardAction(),
				RuleMissAction:    DefaultNamespaceForwardAction(),
			},
			"spex": {
				RulesetMissAction: DefaultNamespaceForwardAction(),
				RuleMissAction:    DefaultNamespaceForwardAction(),
			},
			"cache": {
				RulesetMissAction: DefaultNamespaceForwardAction(),
				RuleMissAction:    DefaultNamespaceForwardAction(),
			},
		},
	}
}
