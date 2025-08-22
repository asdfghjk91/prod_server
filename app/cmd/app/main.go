package main

import (
	"app/internal/app"
	"app/internal/config"
	"app/pkg/common/logging"
	"log"
)

func main() {
	log.Print("config initializing")
	cfg := config.GetConfig()

	log.Print("logger initializing but does not exist in the project")
	logger := logging.GetLogger(cfg.AppConfig.LogLevel)

	a, err := app.NewApp(cfg, &logger)
	if err != nil {
		logger.Fatal((err))
	}

	logger.Println("Running application")
	a.Run()
}
