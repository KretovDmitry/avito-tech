package config

import (
	"os"
	"time"

	"github.com/KretovDmitry/avito-tech/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/qiangxue/go-env"
	"gopkg.in/yaml.v3"
)

const (
	defaultServerPort = 8080
	// HUGE expiration for Postman to test with always valid token
	defaultJWTExpiration      = 480 * time.Hour
	defaultShutdownTimeout    = 30 * time.Second
	defaultCacheExpiration    = 5 * time.Minute
	defaultBannerBufferLength = 5
)

// Config represents an application configuration.
type Config struct {
	// the server port. Defaults to 8080
	ServerPort int `yaml:"server_port" env:"SERVER_PORT"`
	// the data source name (DSN) for connecting to the database. required.
	DSN string `yaml:"dsn" env:"DSN,secret"`
	// JWT signing key. required.
	JWTSigningKey string `yaml:"jwt_signing_key" env:"JWT_SIGNING_KEY,secret"`
	// JWT expiration in hours. Defaults to 12-24 hours maybe
	JWTExpiration time.Duration
	// Shutdown timeout in seconds. Defaults to 30 seconds
	ShutdownTimeout time.Duration
	// Live mode for development instant reload. Local default true
	LiveMode bool `yaml:"live_reload" env:"LIVE_RELOAD"`
	// Length of buffer for async banners deleting. Defaults to 5
	BannerBufferLength int `yaml:"banner_buffer_length" env:"BANNER_BUFFER_LENGTH"`
	// Cache expiration time. Defaults to 5 minutes
	CacheExpiration time.Duration `yaml:"cache_expiration" env:"CACHE_EXPIRATION"`
}

// Validate validates the application configuration.
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DSN, validation.Required),
		validation.Field(&c.JWTSigningKey, validation.Required),
	)
}

// Load returns an application configuration which is populated
// from the given configuration file and environment variables.
func Load(file string, logger log.Logger) (*Config, error) {
	// default config
	c := Config{
		ServerPort:         defaultServerPort,
		JWTExpiration:      defaultJWTExpiration,
		ShutdownTimeout:    defaultShutdownTimeout,
		CacheExpiration:    defaultCacheExpiration,
		LiveMode:           false,
		BannerBufferLength: defaultBannerBufferLength,
	}

	// load from YAML config file
	bytes, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(bytes, &c); err != nil {
		return nil, err
	}

	// load from environment variables prefixed with "APP_"
	if err = env.New("APP_", logger.Infof).Load(&c); err != nil {
		return nil, err
	}

	// validation
	if err = c.Validate(); err != nil {
		return nil, err
	}

	return &c, err
}
