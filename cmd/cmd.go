package cmd

import (
	"fmt"
	"os"

	"github.com/kopecmaciej/mongui/component"
	"github.com/kopecmaciej/mongui/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	cfgFile            string
	debug              bool
	showWelcomePage    bool
	showConnectionPage bool
	rootCmd            = &cobra.Command{
		Use:   "mongui",
		Short: "MongoDB TUI client",
		Long:  `A Terminal User Interface (TUI) client for MongoDB`,
		Run:   runApp,
	}
)

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/mongui/config.yaml)")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode")
	rootCmd.Flags().BoolVar(&showWelcomePage, "show-welcome-page", false, "Show welcome page on startup")
	rootCmd.Flags().BoolVar(&showConnectionPage, "show-connection-page", false, "Show connection page on startup")
}

func runApp(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// Override config values with CLI flags
	cfg.Debug = debug
	cfg.ShowWelcomePage = showWelcomePage
	cfg.ShowConnectionPage = showConnectionPage

	logLevel := zerolog.InfoLevel
	if cfg.Debug {
		logLevel = zerolog.DebugLevel
	}

	logFile := logging(cfg.Log.Path, logLevel, cfg.Log.PrettyPrint)
	defer func() {
		err := logFile.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("Error closing log file")
		}
	}()

	log.Info().Msg("Mongo UI started")
	app := component.NewApp(cfg)
	err = app.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing app")
	}
}

func logging(path string, logLevel zerolog.Level, pretty bool) *os.File {
	logFile, err := os.OpenFile(path, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		if os.IsNotExist(err) {
			logFile, err = os.Create(path)
			if err != nil {
				log.Fatal().Err(err).Msg("Error creating log file")
			}
		} else {
			log.Fatal().Err(err).Msg("Error opening log file")
		}
	}

	zerolog.SetGlobalLevel(logLevel)

	log.Logger = log.Output(logFile)
	if pretty {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: logFile})
	}

	log.Logger = log.With().Caller().Logger()

	return logFile
}
