package bo

import "github.com/MrMiaoMIMI/mockserver/internal/model/eo"

type Namespace struct {
	DBID              uint64                  `json:"-"`
	ID                string                  `json:"id"`
	Name              string                  `json:"name,omitempty"`
	Description       string                  `json:"description,omitempty"`
	RulesetMissAction NamespaceFallbackAction `json:"ruleset_miss_action"`
	RuleMissAction    NamespaceFallbackAction `json:"rule_miss_action"`
}

type NamespaceFallbackAction struct {
	Type     string                     `json:"type"`
	Response *NamespaceResponseFallback `json:"response,omitempty"`
	Forward  *NamespaceForwardFallback  `json:"forward,omitempty"`
}

type NamespaceResponseFallback struct {
	Status  int                 `json:"status"`
	Headers map[string][]string `json:"headers,omitempty"`
	Body    any                 `json:"body,omitempty"`
}

type NamespaceForwardFallback struct {
	TimeoutMS int `json:"timeout_ms,omitempty"`
}

const DefaultNamespaceForwardTimeoutMS = 5000

func DefaultNamespaceFallbackAction() NamespaceFallbackAction {
	return NamespaceFallbackAction{
		Type:    eo.NamespaceFallbackTypeForward,
		Forward: &NamespaceForwardFallback{TimeoutMS: DefaultNamespaceForwardTimeoutMS},
	}
}

func DefaultNamespace(id string) Namespace {
	return Namespace{
		ID:                id,
		Name:              id,
		RulesetMissAction: DefaultNamespaceFallbackAction(),
		RuleMissAction:    DefaultNamespaceFallbackAction(),
	}
}
