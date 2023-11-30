package main

import (
	"flag"
	"mongo-ui/component"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	debug := flag.Bool("debug", false, "sets app in debug mode")
	flag.Parse()

	logLevel := zerolog.InfoLevel
	if *debug {
		logLevel = zerolog.DebugLevel
	}

	defer logging(logLevel).Close()

	app := component.NewApp()
	err := app.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing app")
	}
}

func logging(logLevel zerolog.Level) *os.File {
	LOG_FILE := "./log.txt"

	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal().Err(err).Msg("Error opening log file")
	}

	zerolog.SetGlobalLevel(logLevel)

	return logFile
}
