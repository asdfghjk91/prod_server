package main

import (
	"app/internal/app"
	"app/internal/config"
	"log"
)

func main() {
	log.Print("config initializing")
	cfg := config.GetConfig()

	log.Print("logger initializing but does not exist in the project")

	a, err := app.NewApp(cfg)
	if err != nil {
		log.Fatal((err))
	}

	log.Println("Running application")
	a.Run()
}
