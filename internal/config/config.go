package config

import (
	"fmt"
	"net/url"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Database struct {
	DSN  string `mapstructure:"dsn" validate:"omitempty,url"`
	Name string `mapstructure:"name" validate:""`
}

type Nats struct {
	URL string `mapstructure:"url" validate:"required,url"`
}

type Log struct {
	Level   string `mapstructure:"level" validate:"oneof=trace debug info warn error"`
	Console bool   `mapstructure:"console"`
}

type Handler struct {
	Port int `mapstructure:"port" validate:"omitempty,gte=80,lte=65535"`
}

type Processor struct {
	Port int `mapstructure:"port" validate:"omitempty,gte=80,lte=65535"`
}

type View struct {
	Port int `mapstructure:"port" validate:"omitempty,gte=80,lte=65535"`
}

type Telemetry struct {
	Name     string `mapstructure:"name" validate:"required"`
	Exporter string `mapstructure:"exporter" validate:"oneof=jaeger xray"`
	Endpoint string `mapstructure:"endpoint" validate:"required"`
}

type Config struct {
	Database  Database  `mapstructure:"database"`
	Nats      Nats      `mapstructure:"nats"`
	Log       Log       `mapstructure:"log"`
	Handler   Handler   `mapstructure:"handler"`
	Processor Processor `mapstructure:"processor"`
	View      View      `mapstructure:"view"`
	Telemetry Telemetry `mapstructure:"telemetry"`
}

func (c Config) MarshalZerologObject(e *zerolog.Event) {
	u, _ := url.Parse(c.Database.DSN)
	if u.User != nil {
		username := u.User.Username()
		u.User = url.UserPassword(username, "xxxxxx")
	}

	e.Dict("database", e.CreateDict().Str("dsn", u.String())).
		Dict("log", e.CreateDict().Str("level", c.Log.Level).Bool("console", c.Log.Console)).
		Dict("nats", e.CreateDict().Str("url", c.Nats.URL)).
		Dict("handler", e.CreateDict().Int("port", c.Handler.Port)).
		Dict("processor", e.CreateDict().Int("port", c.Processor.Port)).
		Dict("view", e.CreateDict().Int("port", c.View.Port)).
		Dict("telemetry", e.CreateDict().Str("name", c.Telemetry.Name).
			Str("exporter", c.Telemetry.Exporter).Str("endpoint", c.Telemetry.Endpoint))
}

func LoadConfig(v *viper.Viper) (*Config, error) {
	var config Config

	// Final Step: Unmarshal into the struct
	// Viper will prioritize: Env Var > .env File > Defaults
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validation Step
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

var Module = fx.Module("config",
	fx.Provide(func(v *viper.Viper, zl zerolog.Logger) (*Config, error) {
		config, err := LoadConfig(v)
		if err != nil {
			zl.Error().Err(err).Msg("failed to load config")
			return nil, err
		}
		zl.Info().Interface("config", config).Msg("loaded config")
		return config, nil
	}),
)
