package main

import (
	"flag"
	"log"
	"os"

	"github.com/nsr888/lifecalendar/internal/app"
	"github.com/nsr888/lifecalendar/internal/config"
	"github.com/nsr888/lifecalendar/internal/storage"
)

func main() {
	logger := log.New(os.Stdout, "APP: ", log.LstdFlags)

	var jsonPlan bool
	var aiReview bool
	flag.BoolVar(&jsonPlan, "json-plan", false, "Output vacation plans as JSON with weekend/holiday data")
	flag.BoolVar(&aiReview, "ai-review", false, "Review vacation plans with AI and print analysis")
	flag.Parse()

	var appConfig *config.Config
	var err error

	// Get config file path from remaining arguments
	args := flag.Args()
	if len(args) > 0 {
		appConfig, err = config.Load(args[0])
	} else {
		appConfig, err = config.LoadDefault()
	}

	if err != nil {
		logger.Fatalf("Failed to load app config: %v", err)
	}

	// Store flags in config
	appConfig.JSONPlan = jsonPlan
	appConfig.AIReview = aiReview

	csvStorage := storage.NewCSVStorage(appConfig.GetDataFolderWithFallback())
	appService := app.NewService(csvStorage, logger)

	if aiReview {
		if runErr := appService.RunAIReview(appConfig); runErr != nil {
			logger.Fatalf("Application failed: %v", runErr)
		}
	} else if jsonPlan {
		if runErr := appService.RunJSONPlan(appConfig); runErr != nil {
			logger.Fatalf("Application failed: %v", runErr)
		}
	} else {
		if runErr := appService.Run(appConfig); runErr != nil {
			logger.Fatalf("Application failed: %v", runErr)
		}
	}
}
