package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

const (
	defaultLogLevel = zerolog.InfoLevel
	version         = "v0.0.1"
)

var cfgFile string
var v = viper.New()
var logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Level(defaultLogLevel)

var rootCmd = &cobra.Command{
	Use:   "simdrone",
	Short: "drone army simulator",
	Long: `
A drone army simulator based on github.com/cloudnativego/drone-* to help
learn some fundamental Go concepts, programming patterns, and AWS usage.

Uses environment variables - prefixed DRONE_ - for configuration.`,
	PersistentPreRunE: globalPreRun,
	Version:           version,
	SilenceErrors:     true,
	SilenceUsage:      true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal().Err(err).Msg("exiting application...")
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func globalPreRun(_ *cobra.Command, _ []string) error {
	if v.GetBool("log.console") {
		logger = logger.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	level := v.GetString("log.level")
	lvl, err := zerolog.ParseLevel(level)
	if err == nil && lvl != logger.GetLevel() {
		logger = logger.Level(lvl)
	}
	logger.Info().Err(err).Str("min_level", logger.GetLevel().String()).Msg("minimum logging level")

	log.Logger = logger
	return nil
}

func initConfig() {
	// 1. Set default values
	v.SetDefault("database.dsn", "")
	v.SetDefault("database.name", "example")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.console", false)
	v.SetDefault("nats.url", "nats://localhost:4222")
	v.SetDefault("handler.port", 1313)
	v.SetDefault("processor.port", 1314)
	v.SetDefault("view.port", 1315)

	// 2. Load any .env settings into the environment
	if cfgFile != "" {
		// Use config file from the flag.
		err := godotenv.Load(cfgFile)
		if err != nil {
			logger.Error().Err(err).Str("cfgFile", cfgFile).Msg("failed to load config file")
		}
	}
	err := godotenv.Load()
	if err != nil {
		logger.Error().Err(err).Msg("failed to load ./.env file; skipping")
	}

	// 3. Setup Environment Variable Logic
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetEnvPrefix("DRONE") // Prepends "DRONE_" to all env lookups
	v.AutomaticEnv()

	//logger.Info().Str("cfgFile", cfgFile).Msg("initialized config")
}

// commonModule returns a set of common fx.Options to be used across multiple cobra commands.
func commonModule(cmd *cobra.Command) fx.Option {
	return fx.Module("common",
		fx.Supply(v),
		fx.Supply(logger),
		fx.Supply(
			fx.Annotate(
				cmd.Context(),
				fx.As(new(context.Context)),
			),
		),
	)
}
