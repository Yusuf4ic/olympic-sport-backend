package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig      `mapstructure:",squash"`
	Server   ServerConfig   `mapstructure:",squash"`
	Postgres PostgresConfig `mapstructure:",squash"`
	Redis    RedisConfig    `mapstructure:",squash"`
	JWT      JWTConfig      `mapstructure:",squash"`
	HMAC     HMACConfig     `mapstructure:",squash"`
}

type AppConfig struct {
	Env      string `mapstructure:"APP_ENV"`
	LogLevel string `mapstructure:"LOG_LEVEL"`
}

type ServerConfig struct {
	Port            string        `mapstructure:"SERVER_PORT"`
	Mode            string        `mapstructure:"SERVER_MODE"`
	ShutdownTimeout time.Duration `mapstructure:"SERVER_SHUTDOWN_TIMEOUT"`
}

type JWTConfig struct {
	Secret     string        `mapstructure:"JWT_SECRET"`
	Expiration time.Duration `mapstructure:"JWT_EXPIRATION"`
}

type HMACConfig struct {
	Secret string `mapstructure:"HMAC_SECRET"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"POSTGRES_HOST"`
	Port            string        `mapstructure:"POSTGRES_PORT"`
	User            string        `mapstructure:"POSTGRES_USER"`
	Password        string        `mapstructure:"POSTGRES_PASSWORD"`
	DBName          string        `mapstructure:"POSTGRES_DB"`
	SSLMode         string        `mapstructure:"POSTGRES_SSLMODE"`
	MaxConns        int32         `mapstructure:"POSTGRES_MAX_CONNS"`
	MinConns        int32         `mapstructure:"POSTGRES_MIN_CONNS"`
	MaxConnIdleTime time.Duration `mapstructure:"POSTGRES_MAX_CONN_IDLE_TIME"`
	MaxConnLifetime time.Duration `mapstructure:"POSTGRES_MAX_CONN_LIFETIME"`
}

func (p *PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}

type RedisConfig struct {
	Host     string `mapstructure:"REDIS_HOST"`
	Port     string `mapstructure:"REDIS_PORT"`
	Password string `mapstructure:"REDIS_PASSWORD"`
	DB       int    `mapstructure:"REDIS_DB"`
	PoolSize int    `mapstructure:"REDIS_POOL_SIZE"`
	TLS      bool   `mapstructure:"REDIS_TLS"`
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig(path string) (*Config, error) {
	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("SERVER_PORT", "8080")
	viper.SetDefault("SERVER_MODE", "debug")
	viper.SetDefault("SERVER_SHUTDOWN_TIMEOUT", "10s")
	viper.SetDefault("POSTGRES_HOST", "localhost")
	viper.SetDefault("POSTGRES_PORT", "5432")
	viper.SetDefault("POSTGRES_USER", "postgres")
	viper.SetDefault("POSTGRES_PASSWORD", "postgres")
	viper.SetDefault("POSTGRES_DB", "olympic_sport_db")
	viper.SetDefault("POSTGRES_SSLMODE", "disable")
	viper.SetDefault("POSTGRES_MAX_CONNS", 50)
	viper.SetDefault("POSTGRES_MIN_CONNS", 10)
	viper.SetDefault("POSTGRES_MAX_CONN_IDLE_TIME", "15m")
	viper.SetDefault("POSTGRES_MAX_CONN_LIFETIME", "1h")
	viper.SetDefault("REDIS_HOST", "localhost")
	viper.SetDefault("REDIS_PORT", "6379")
	viper.SetDefault("REDIS_PASSWORD", "")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("REDIS_POOL_SIZE", 100)
	viper.SetDefault("REDIS_TLS", false)
	viper.SetDefault("JWT_SECRET", "supersecretjwtkeyforolympicsports")
	viper.SetDefault("JWT_EXPIRATION", "24h")
	viper.SetDefault("HMAC_SECRET", "supersecrethmacsigningkeyforqrcodes")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg.App); err != nil {
		return nil, fmt.Errorf("error unmarshaling app config: %w", err)
	}
	if err := viper.Unmarshal(&cfg.Server); err != nil {
		return nil, fmt.Errorf("error unmarshaling server config: %w", err)
	}
	if err := viper.Unmarshal(&cfg.Postgres); err != nil {
		return nil, fmt.Errorf("error unmarshaling postgres config: %w", err)
	}
	if err := viper.Unmarshal(&cfg.Redis); err != nil {
		return nil, fmt.Errorf("error unmarshaling redis config: %w", err)
	}
	if err := viper.Unmarshal(&cfg.JWT); err != nil {
		return nil, fmt.Errorf("error unmarshaling jwt config: %w", err)
	}
	if err := viper.Unmarshal(&cfg.HMAC); err != nil {
		return nil, fmt.Errorf("error unmarshaling hmac config: %w", err)
	}

	return &cfg, nil
}
