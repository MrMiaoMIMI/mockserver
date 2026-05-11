package dao

import "go.uber.org/fx"

var Module = fx.Module("dao",
	fx.Provide(
		NewDB,
		provideRuleSetRepository,
		provideNamespaceRepository,
	),
)

func provideRuleSetRepository(db DB) RuleSetRepository {
	return db.GetRuleSetRepository()
}

func provideNamespaceRepository(db DB) NamespaceRepository {
	return db.GetNamespaceRepository()
}
