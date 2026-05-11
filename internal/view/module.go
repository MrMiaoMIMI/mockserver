package view

import "go.uber.org/fx"

var Module = fx.Module("view",
	fx.Provide(
		NewRuleSetView,
		NewNamespaceView,
		NewRuntimeView,
	),
)
