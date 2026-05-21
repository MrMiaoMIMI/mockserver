package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	authlib "github.com/MrMiaoMIMI/mockserver/internal/auth"
	"github.com/MrMiaoMIMI/mockserver/internal/config"
	"github.com/MrMiaoMIMI/mockserver/internal/controller"
)

var Module = fx.Module("router",
	fx.Provide(
		newAuthConfig,
		newRouteConfig,
		newEngine,
	),
)

func newAuthConfig(cfg config.Config) AuthConfig {
	return AuthConfig{
		JWT: authlib.Config{
			JWTSecret:         cfg.AuthJWTSecret,
			DebugLoginEnabled: cfg.AuthDebugLoginEnabled,
		},
	}
}

func newRouteConfig(cfg config.Config) RouteConfig {
	return RouteConfig{
		APIPrefix: cfg.APIPrefix,
	}
}

type engineParams struct {
	fx.In

	AdminController   *controller.AdminController
	RuntimeController *controller.RuntimeController
	MetricsController *controller.MetricsController
	TrafficController *controller.TrafficController
	AuthConfig        AuthConfig
	RouteConfig       RouteConfig
}

func newEngine(p engineParams) *gin.Engine {
	return NewWithConfig(p.AdminController, p.RuntimeController, p.AuthConfig, p.RouteConfig, p.MetricsController, p.TrafficController)
}
