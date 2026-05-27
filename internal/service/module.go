package service

import "go.uber.org/fx"

var Module = fx.Module("service",
	fx.Provide(
		NewRuleSetService,
		NewNamespaceService,
		NewScenarioService,
		NewRuntimeServiceWithScenarios,
		NewTrafficService,
	),
)
