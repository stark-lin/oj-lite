// Reads and parses file configuration into the application config structure.

package config

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"
)

const DefaultPath = "config.json"

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

type fileConfig struct {
	App       fileAppConfig       `json:"app"`
	HTTP      fileHTTPConfig      `json:"http"`
	DB        fileDBConfig        `json:"db"`
	Scheduler fileSchedulerConfig `json:"scheduler"`
}

type fileAppConfig struct {
	Name string `json:"name"`
	Env  string `json:"env"`
}

type fileHTTPConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	GinMode         string `json:"gin_mode"`
	ReadTimeout     string `json:"read_timeout"`
	WriteTimeout    string `json:"write_timeout"`
	IdleTimeout     string `json:"idle_timeout"`
	ShutdownTimeout string `json:"shutdown_timeout"`
}

type fileDBConfig struct {
	Path        string `json:"path"`
	BusyTimeout string `json:"busy_timeout"`
}

type fileSchedulerConfig struct {
	Concurrency    int    `json:"concurrency"`
	FetchBatchSize int    `json:"fetch_batch_size"`
	IdleSleep      string `json:"idle_sleep"`
}

func Load() (Config, error) {
	return LoadFile(DefaultPath)
}

func LoadFile(path string) (Config, error) {
	if _, err := os.Stat(path); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("stat config file %q: %w", path, err)
		}

		if err := writeDefaultFile(path); err != nil {
			return Config{}, err
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read config file %q: %w", path, err)
	}

	raw := defaultFileConfig()
	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return Config{}, fmt.Errorf("decode config file %q: %w", path, err)
	}
	if err := ensureNoTrailingJSON(decoder); err != nil {
		return Config{}, fmt.Errorf("decode config file %q: %w", path, err)
	}

	cfg, err := raw.toConfig()
	if err != nil {
		return Config{}, fmt.Errorf("parse config file %q: %w", path, err)
	}

	return cfg, nil
}

func (cfg HTTPConfig) Address() string {
	return net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))
}

func ensureNoTrailingJSON(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err == nil {
		return errors.New("unexpected extra JSON content")
	} else if !errors.Is(err, io.EOF) {
		return err
	}

	return nil
}

func writeDefaultFile(path string) error {
	content, err := json.MarshalIndent(defaultFileConfig(), "", "  ")
	if err != nil {
		return fmt.Errorf("marshal default config: %w", err)
	}
	content = append(content, '\n')

	if err := os.WriteFile(path, content, 0o644); err != nil {
		return fmt.Errorf("write default config file %q: %w", path, err)
	}

	return nil
}

func defaultFileConfig() fileConfig {
	return fileConfig{
		App: fileAppConfig{
			Name: "oj-lite",
			Env:  "local",
		},
		HTTP: fileHTTPConfig{
			Host:            "0.0.0.0",
			Port:            8080,
			GinMode:         "debug",
			ReadTimeout:     "5s",
			WriteTimeout:    "10s",
			IdleTimeout:     "60s",
			ShutdownTimeout: "10s",
		},
		DB: fileDBConfig{
			Path:        "oj-lite.db",
			BusyTimeout: "5s",
		},
		Scheduler: fileSchedulerConfig{
			Concurrency:    4,
			FetchBatchSize: 4,
			IdleSleep:      "1s",
		},
	}
}

func (raw fileConfig) toConfig() (Config, error) {
	readTimeout, err := parseDuration("http.read_timeout", raw.HTTP.ReadTimeout)
	if err != nil {
		return Config{}, err
	}
	writeTimeout, err := parseDuration("http.write_timeout", raw.HTTP.WriteTimeout)
	if err != nil {
		return Config{}, err
	}
	idleTimeout, err := parseDuration("http.idle_timeout", raw.HTTP.IdleTimeout)
	if err != nil {
		return Config{}, err
	}
	shutdownTimeout, err := parseDuration("http.shutdown_timeout", raw.HTTP.ShutdownTimeout)
	if err != nil {
		return Config{}, err
	}
	dbBusyTimeout, err := parseDuration("db.busy_timeout", raw.DB.BusyTimeout)
	if err != nil {
		return Config{}, err
	}
	schedulerIdleSleep, err := parseDuration("scheduler.idle_sleep", raw.Scheduler.IdleSleep)
	if err != nil {
		return Config{}, err
	}

	schedulerConcurrency := raw.Scheduler.Concurrency
	schedulerFetchBatchSize := raw.Scheduler.FetchBatchSize
	if schedulerConcurrency <= 0 {
		schedulerConcurrency = 1
	}
	if schedulerFetchBatchSize <= 0 {
		schedulerFetchBatchSize = schedulerConcurrency
	}

	return Config{
		App: AppConfig{
			Name: raw.App.Name,
			Env:  raw.App.Env,
		},
		HTTP: HTTPConfig{
			Host:            raw.HTTP.Host,
			Port:            raw.HTTP.Port,
			GinMode:         raw.HTTP.GinMode,
			ReadTimeout:     readTimeout,
			WriteTimeout:    writeTimeout,
			IdleTimeout:     idleTimeout,
			ShutdownTimeout: shutdownTimeout,
		},
		DB: DBConfig{
			Path:        raw.DB.Path,
			BusyTimeout: dbBusyTimeout,
		},
		Scheduler: SchedulerConfig{
			Concurrency:    schedulerConcurrency,
			FetchBatchSize: schedulerFetchBatchSize,
			IdleSleep:      schedulerIdleSleep,
		},
	}, nil
}

func parseDuration(name, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}

	return duration, nil
}
