package main

import (
	"context"

	"github.com/MrMiaoMIMI/goshared/logger"

	"mockserver/internal/config"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal(ctx, "Load config failed", logger.Err(err))
	}
	if err := run(ctx, cfg); err != nil {
		logger.Fatal(ctx, "Mockserver stopped", logger.Err(err))
	}
}
