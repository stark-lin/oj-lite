// Reads and parses environment configuration into the application config structure.

package config

import (
	"net"
	"os"
	"strconv"
	"time"
)

type Config struct {
	App       AppConfig
	DB        DBConfig
	HTTP      HTTPConfig
	Scheduler SchedulerConfig
}

type AppConfig struct {
	Name string
	Env  string
}

type HTTPConfig struct {
	Host            string
	Port            int
	GinMode         string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type DBConfig struct {
	Path        string
	BusyTimeout time.Duration
}

type SchedulerConfig struct {
	Concurrency    int
	FetchBatchSize int
	IdleSleep      time.Duration
}

func Load() Config {
	schedulerConcurrency := envInt("SCHEDULER_CONCURRENCY", 4)
	schedulerFetchBatchSize := envInt("SCHEDULER_FETCH_BATCH_SIZE", schedulerConcurrency)
	if schedulerConcurrency <= 0 {
		schedulerConcurrency = 1
	}
	if schedulerFetchBatchSize <= 0 {
		schedulerFetchBatchSize = schedulerConcurrency
	}

	return Config{
		App: AppConfig{
			Name: envString("APP_NAME", "oj-lite"),
			Env:  envString("APP_ENV", "local"),
		},
		HTTP: HTTPConfig{
			Host:            envString("HTTP_HOST", "0.0.0.0"),
			Port:            envInt("HTTP_PORT", 8080),
			GinMode:         envString("GIN_MODE", "debug"),
			ReadTimeout:     envDuration("HTTP_READ_TIMEOUT", 5*time.Second),
			WriteTimeout:    envDuration("HTTP_WRITE_TIMEOUT", 10*time.Second),
			IdleTimeout:     envDuration("HTTP_IDLE_TIMEOUT", 60*time.Second),
			ShutdownTimeout: envDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second),
		},
		DB: DBConfig{
			Path:        envString("DB_PATH", "data/oj-lite.db"),
			BusyTimeout: envDuration("DB_BUSY_TIMEOUT", 5*time.Second),
		},
		Scheduler: SchedulerConfig{
			Concurrency:    schedulerConcurrency,
			FetchBatchSize: schedulerFetchBatchSize,
			IdleSleep:      envDuration("SCHEDULER_IDLE_SLEEP", time.Second),
		},
	}
}

func (cfg HTTPConfig) Address() string {
	return net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
