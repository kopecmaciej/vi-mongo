package cmd

import (
	"os"
	"strings"

	"fmt"

	"github.com/kopecmaciej/vi-mongo/internal/config"
	"github.com/kopecmaciej/vi-mongo/internal/tui"
	"github.com/kopecmaciej/vi-mongo/internal/util"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	cfgFile           string
	showVersion       bool
	debug             bool
	welcomePage       bool
	connectionPage    bool
	connectionName    string
	listConnections   bool
	encryptionKeyPath string
	jumpInto          string
	rootCmd           = &cobra.Command{
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
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.config/vi-mongo/config.yaml)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Show version")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.Flags().BoolVarP(&welcomePage, "welcome-page", "w", false, "Show welcome page on startup")
	rootCmd.Flags().BoolVarP(&connectionPage, "connection-page", "p", false, "Show connection page on startup")
	rootCmd.Flags().StringVarP(&connectionName, "connection-name", "n", "", "Connect to a specific MongoDB connection by name")
	rootCmd.Flags().BoolVarP(&listConnections, "connection-list", "l", false, "List all available connections")
	rootCmd.Flags().StringVar(&encryptionKeyPath, "key-path", "", "Path to the encryption key file")
	rootCmd.Flags().Bool("gen-key", false, "Generate valid encryption key")
	rootCmd.Flags().StringVarP(&jumpInto, "jump", "j", "", "Jump directly to database/collection (format: db-name/collection-name)")
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
	cfg, err := config.LoadConfigWithVersion(version)
	if err != nil {
		fatalf("loading config: %v", err)
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
		case "connection-list":
			listAvailableConnections(cfg)
			os.Exit(0)
		case "connection-name":
			found := false
			for _, conn := range cfg.Connections {
				if conn.Name == connectionName {
					found = true
					cfg.CurrentConnection = connectionName
					cfg.ShowConnectionPage = false
					break
				}
			}
			if !found {
				fatalf("Connection '%s' not found. Use --list or -l to see available connections.", connectionName)
			}
		case "gen-key":
			util.PrintEncryptionKeyInstructions()
			os.Exit(0)
		case "key-path":
			if encryptionKeyPath != "" {
				if _, err := os.ReadFile(encryptionKeyPath); err != nil {
					fatalf("reading encryption key from %s: %v", encryptionKeyPath, err)
				}
				cfg.EncryptionKeyPath = &encryptionKeyPath
				if err := cfg.UpdateConfig(); err != nil {
					fatalf("saving path to config file: %v", err)
				}
				fmt.Println("Encryption key file path saved successfully")
			}
			os.Exit(0)
		case "jump":
			if jumpInto != "" {
				if err := validateDirectNavigateFormat(jumpInto); err != nil {
					fatalf("invalid jump format: %v", err)
				}
				cfg.JumpInto = jumpInto
				cfg.ShowConnectionPage = false
				cfg.ShowWelcomePage = false
			} else {
				fatalf("jump value cannot be empty")
			}
		}
	})

	if err := cfg.LoadEncryptionKey(); err != nil {
		fatalf("loading encryption key: %v", err)
	}

	logLevel := zerolog.InfoLevel
	if debug {
		logLevel = zerolog.DebugLevel
	}

	logFile := logging(cfg.Log.Path, logLevel, cfg.Log.PrettyPrint)
	defer func() {
		err := logFile.Close()
		if err != nil {
			fmt.Printf("\nError closing log file %s, error: %s", cfg.Log.Path, err)
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

func listAvailableConnections(cfg *config.Config) {
	if len(cfg.Connections) == 0 {
		fmt.Println("No connections available. Use the app to add connections.")
		return
	}

	maxNameLength := 4
	for _, conn := range cfg.Connections {
		if len(conn.Name) > maxNameLength {
			maxNameLength = len(conn.Name)
		}
	}

	maxNameLength += 2

	fmt.Println("Available connections:")
	fmt.Printf("%-2s %-*s %s\n", "", maxNameLength, "NAME", "URL")
	fmt.Printf("%-2s %-*s %s\n", "", maxNameLength, "----", "---")

	for _, conn := range cfg.Connections {
		currentMark := " "
		if cfg.CurrentConnection == conn.Name {
			currentMark = "*"
		}
		fmt.Printf("%s %-*s %s\n", currentMark, maxNameLength, conn.Name, conn.GetSafeUri())
	}

	fmt.Println("\n* Current connection")
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

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

func validateDirectNavigateFormat(format string) error {
	parts := strings.Split(format, "/")
	if len(parts) != 2 {
		return fmt.Errorf("format should be db-name/collection-name")
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return fmt.Errorf("both db-name and collection-name are required")
	}
	return nil
}
