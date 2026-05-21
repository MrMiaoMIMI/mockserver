package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"

	"github.com/MrMiaoMIMI/goshared/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"

	"github.com/MrMiaoMIMI/mockserver/internal/config"
	"github.com/MrMiaoMIMI/mockserver/internal/controller"
	"github.com/MrMiaoMIMI/mockserver/internal/dao"
	"github.com/MrMiaoMIMI/mockserver/internal/observability"
	"github.com/MrMiaoMIMI/mockserver/internal/router"
	"github.com/MrMiaoMIMI/mockserver/internal/service"
	"github.com/MrMiaoMIMI/mockserver/internal/view"
)

func Run(ctx context.Context, cfg config.Config) error {
	logger.Init(logger.DefaultConfig())
	fxApp := newFxApp(ctx, cfg, fx.Invoke(registerHTTPServer))
	if err := fxApp.Err(); err != nil {
		return err
	}
	fxApp.Run()
	return nil
}

func newFxApp(ctx context.Context, cfg config.Config, opts ...fx.Option) *fx.App {
	if ctx == nil {
		ctx = context.Background()
	}
	base := []fx.Option{
		fx.Supply(
			fx.Annotate(ctx, fx.As(new(context.Context))),
			cfg,
		),
		serverModule,
	}
	base = append(base, opts...)
	return fx.New(base...)
}

var serverModule = fx.Module("mockserver",
	dao.Module,
	service.Module,
	view.Module,
	observability.Module,
	controller.Module,
	router.Module,
	fx.Provide(
		newHTTPServer,
	),
	fx.Invoke(initializeRuleSetData),
)

type initializerParams struct {
	fx.In

	Context          context.Context
	NamespaceService service.NamespaceService
}

func initializeRuleSetData(p initializerParams) error {
	if _, err := p.NamespaceService.EnsureDefaultNamespace(p.Context); err != nil {
		return fmt.Errorf("bootstrap default namespace failed: %w", err)
	}
	return nil
}

func newHTTPServer(cfg config.Config, engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    cfg.Address,
		Handler: engine,
	}
}

func registerHTTPServer(lc fx.Lifecycle, server *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := net.Listen("tcp", server.Addr)
			if err != nil {
				return fmt.Errorf("listen %s: %w", server.Addr, err)
			}
			logger.Info(ctx, "Mockserver listening", logger.String("address", server.Addr))
			go func() {
				if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Error(context.Background(), "Mockserver HTTP server stopped unexpectedly", logger.Err(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Info(ctx, "Mockserver stopping", logger.String("address", server.Addr))
			return server.Shutdown(ctx)
		},
	})
}
