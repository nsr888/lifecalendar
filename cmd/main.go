package main

import (
	"log"
	"os"

	"github.com/nsr888/lifecalendar/internal/app"
	"github.com/nsr888/lifecalendar/internal/config"
	"github.com/nsr888/lifecalendar/internal/storage"
)

func main() {
	logger := log.New(os.Stdout, "APP: ", log.LstdFlags)

	var appConfig *config.Config
	var err error

	if len(os.Args) > 1 {
		appConfig, err = config.Load(os.Args[1])
	} else {
		appConfig, err = config.LoadDefault()
	}

	if err != nil {
		logger.Fatalf("Failed to load app config: %v", err)
	}

	csvStorage := storage.NewCSVStorage(appConfig.GetDataFolderWithFallback())
	appService := app.NewService(csvStorage, logger)

	if runErr := appService.Run(appConfig); runErr != nil {
		logger.Fatalf("Application failed: %v", runErr)
	}
}
