package main

import (
	"mongo-ui/component"
	"mongo-ui/config"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	config, err := config.LoadAppConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}

	logLevel := zerolog.InfoLevel
	if config.Debug {
		logLevel = zerolog.DebugLevel
	}

	file := logging(logLevel, false)
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("Error closing log file")
		}
	}()

	app := component.NewApp(config)
	err = app.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing app")
	}

	log.Info().Msg("Mongo UI started")
}

func logging(logLevel zerolog.Level, pretty bool) *os.File {
	LOG_FILE := "mongo-ui.log"

	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening log file")
	}

	zerolog.SetGlobalLevel(logLevel)

	log.Logger = log.Output(logFile)
	if pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: logFile})
	}

	log.Logger = log.With().Caller().Logger()

	return logFile
}
