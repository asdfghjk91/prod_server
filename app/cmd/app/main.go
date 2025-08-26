package main

import (
	"app/internal/app"
	"app/internal/config"
	"app/pkg/common/logging"
	"context"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load() // подхватит .env, если есть

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.GetLogger(ctx)

	logger.Info("config initializing")
	cfg := config.GetConfig()

	ctx = logging.ContextWithLogger(ctx, logger)

	a, err := app.NewApp(ctx, cfg)
	if err != nil {
		logger.Fatalln(err)
	}

	logger.Info("Running application")
	a.Run(ctx)
}
