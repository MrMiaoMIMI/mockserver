package bo

type Namespace struct {
	DBID        uint64                     `json:"-"`
	ID          string                     `json:"id"`
	Name        string                     `json:"name,omitempty"`
	Description string                     `json:"description,omitempty"`
	Policies    map[string]NamespacePolicy `json:"policies"`
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
		ID:   id,
		Name: id,
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
