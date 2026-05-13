package router

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/MrMiaoMIMI/mockserver/internal/config"
	"github.com/MrMiaoMIMI/mockserver/internal/controller"
)

var Module = fx.Module("router",
	fx.Provide(
		newAdminAuthConfig,
		newEngine,
	),
)

func newAdminAuthConfig(cfg config.Config) AdminAuthConfig {
	return AdminAuthConfig{
		AdminToken:   cfg.AdminToken,
		ReadToken:    cfg.AdminReadToken,
		WriteToken:   cfg.AdminWriteToken,
		PublishToken: cfg.AdminPublishToken,
	}
}

type engineParams struct {
	fx.In

	AdminController   *controller.AdminController
	RuntimeController *controller.RuntimeController
	MetricsController *controller.MetricsController
	TrafficController *controller.TrafficController
	AuthConfig        AdminAuthConfig
}

func newEngine(p engineParams) *gin.Engine {
	return NewWithTraffic(p.AdminController, p.RuntimeController, p.AuthConfig, p.MetricsController, p.TrafficController)
}
