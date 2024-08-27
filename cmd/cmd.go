package cmd

import (
	"os"

	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	cfgFile            string
	debug              bool
	showWelcomePage    bool
	showConnectionPage bool
	rootCmd            = &cobra.Command{
		Use:   "vi-mongo",
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
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.config/vi-mongo/config.yaml)")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode")
	rootCmd.Flags().BoolVar(&showWelcomePage, "show-welcome-page", false, "Show welcome page on startup")
	rootCmd.Flags().BoolVar(&showConnectionPage, "show-connection-page", false, "Show connection page on startup")
}

func runApp(cmd *cobra.Command, args []string) {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
		os.Exit(1)
	}

	cmd.Flags().Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "debug":
			cfg.Debug = debug
		case "show-welcome-page":
			cfg.ShowWelcomePage = showWelcomePage
		case "show-connection-page":
			cfg.ShowConnectionPage = showConnectionPage
		}
	})

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

	if cfg.Debug {
		log.Info().Msg("Debug mode enabled")
	}
	log.Info().Msg("Mongo UI started")
	app := tui.NewApp(cfg)
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
