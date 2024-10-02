package cmd

import (
	"os"

	"fmt"

	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	cfgFile        string
	showVersion    bool
	debug          bool
	welcomePage    bool
	connectionPage bool
	rootCmd        = &cobra.Command{
		Use:   "vi-mongo",
		Short: "MongoDB TUI client",
		Long:  `A Terminal User Interface (TUI) client for MongoDB`,
		Run:   runApp,
	}

	version = "v0.0.0"
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
	rootCmd.Flags().BoolVar(&showVersion, "version", false, "Show version")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug mode")
	rootCmd.Flags().BoolVar(&welcomePage, "welcome-page", false, "Show welcome page on startup")
	rootCmd.Flags().BoolVar(&connectionPage, "connection-page", false, "Show connection page on startup")
}

func runApp(cmd *cobra.Command, args []string) {
	if showVersion {
		greenColor := "\033[32m"
		resetColor := "\033[0m"
		fmt.Printf("%s\n", greenColor)
		fmt.Print(`
 __      ___   __  __                       
 \ \    / (_) |  \/  |                      
  \ \  / / _  | \  / | ___  _ __   __ _  ___
   \ \/ / | | | |\/| |/ _ \| '_ \ / _` + "`" + ` |/ _ \
    \  /  | | | |  | | (_) | | | | (_| | (_) |
     \/   |_| |_|  |_|\___/|_| |_|\__, |\___/
                                   __/ |     
                                  |___/      
`)
		fmt.Printf("Version %s%s\n", version, resetColor)
		os.Exit(0)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
		os.Exit(1)
	}

	debug := false

	cmd.Flags().Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "version":
			showVersion = true
		case "debug":
			debug = true
		case "welcome-page":
			cfg.ShowWelcomePage = welcomePage
		case "connection-page":
			cfg.ShowConnectionPage = connectionPage
		}
	})

	logLevel := zerolog.InfoLevel
	if debug {
		logLevel = zerolog.DebugLevel
	}

	logFile := logging(cfg.Log.Path, logLevel, cfg.Log.PrettyPrint)
	defer func() {
		err := logFile.Close()
		if err != nil {
			log.Fatal().Err(err).Msg("Error closing log file")
		}
	}()

	if debug {
		log.Debug().Msg("Debug mode enabled")
	}
	log.Info().Msg("Mongo UI started")

	if os.Getenv("ENV") == "vi-dev" {
		log.Info().Msg("Dev mode enabled, keys and styles will be loaded from default values")
	}

	app := tui.NewApp(cfg)
	err = app.Init()
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing app")
	}
	app.Render()
	err = app.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("Error running app")
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
