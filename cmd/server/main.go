package main

import (
	"context"

	"github.com/MrMiaoMIMI/goshared/logger"

	"github.com/MrMiaoMIMI/mockserver/internal/config"
	"github.com/MrMiaoMIMI/mockserver/internal/server"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal(ctx, "Load config failed", logger.Err(err))
	}
	if err := server.Run(ctx, cfg); err != nil {
		logger.Fatal(ctx, "Mockserver stopped", logger.Err(err))
	}
}
