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

	file := logging(config.Log.Path, logLevel, config.Log.PrettyPrint)
	defer func() {
		err := file.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("Error closing log file")
		}
	}()

	log.Info().Msg("Mongo UI started")
	app := component.NewApp(config)
	err = app.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing app")
	}
}

func logging(path string, logLevel zerolog.Level, pretty bool) *os.File {
	logFile, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
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
