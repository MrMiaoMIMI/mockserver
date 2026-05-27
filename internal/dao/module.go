package dao

import "go.uber.org/fx"

var Module = fx.Module("dao",
	fx.Provide(
		NewDB,
		provideRuleSetRepository,
		provideNamespaceRepository,
		provideTrafficRepository,
		provideScenarioRepository,
	),
)

func provideRuleSetRepository(db DB) RuleSetRepository {
	return db.GetRuleSetRepository()
}

func provideNamespaceRepository(db DB) NamespaceRepository {
	return db.GetNamespaceRepository()
}

func provideTrafficRepository(db DB) TrafficRepository {
	return db.GetTrafficRepository()
}

func provideScenarioRepository(db DB) ScenarioRepository {
	return db.GetScenarioRepository()
}
