package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadFileCreatesDefaultConfigWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("load missing config: %v", err)
	}

	if cfg.App.Name != "oj-lite" {
		t.Fatalf("app name = %q, want %q", cfg.App.Name, "oj-lite")
	}
	if cfg.DB.Path != "oj-lite.db" {
		t.Fatalf("db path = %q, want %q", cfg.DB.Path, "oj-lite.db")
	}
	if cfg.HTTP.ReadTimeout != 5*time.Second {
		t.Fatalf("http read timeout = %s, want 5s", cfg.HTTP.ReadTimeout)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read created config: %v", err)
	}
	if !strings.Contains(string(content), `"oj-lite"`) || !strings.Contains(string(content), `"read_timeout"`) {
		t.Fatalf("created config does not look like default JSON: %s", string(content))
	}
}

func TestLoadFileUsesExistingConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{
  "app": {
    "name": "custom-oj",
    "env": "test"
  },
  "http": {
    "host": "127.0.0.1",
    "port": 9090,
    "gin_mode": "test",
    "read_timeout": "2s",
    "write_timeout": "3s",
    "idle_timeout": "4s",
    "shutdown_timeout": "5s"
  },
  "db": {
    "path": "custom.db",
    "busy_timeout": "6s"
  },
  "scheduler": {
    "concurrency": 7,
    "fetch_batch_size": 8,
    "idle_sleep": "9s"
  }
}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("load existing config: %v", err)
	}

	if cfg.App.Name != "custom-oj" || cfg.App.Env != "test" {
		t.Fatalf("app config = %#v", cfg.App)
	}
	if cfg.HTTP.Address() != "127.0.0.1:9090" {
		t.Fatalf("http address = %q, want 127.0.0.1:9090", cfg.HTTP.Address())
	}
	if cfg.HTTP.GinMode != "test" || cfg.HTTP.ReadTimeout != 2*time.Second || cfg.HTTP.ShutdownTimeout != 5*time.Second {
		t.Fatalf("http config = %#v", cfg.HTTP)
	}
	if cfg.DB.Path != "custom.db" || cfg.DB.BusyTimeout != 6*time.Second {
		t.Fatalf("db config = %#v", cfg.DB)
	}
	if cfg.Scheduler.Concurrency != 7 || cfg.Scheduler.FetchBatchSize != 8 || cfg.Scheduler.IdleSleep != 9*time.Second {
		t.Fatalf("scheduler config = %#v", cfg.Scheduler)
	}
}

func TestLoadFileFailsInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"app":`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("load invalid JSON succeeded")
	}
	if !strings.Contains(err.Error(), "decode config file") {
		t.Fatalf("error = %v, want decode config file", err)
	}
}

func TestLoadFileFailsInvalidDuration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{
  "http": {
    "read_timeout": "soon"
  }
}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	_, err := LoadFile(path)
	if err == nil {
		t.Fatal("load invalid duration succeeded")
	}
	if !strings.Contains(err.Error(), "http.read_timeout") {
		t.Fatalf("error = %v, want http.read_timeout", err)
	}
}

func TestLoadFileNormalizesSchedulerLimits(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{
  "scheduler": {
    "concurrency": 0,
    "fetch_batch_size": 0
  }
}`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Scheduler.Concurrency != 1 || cfg.Scheduler.FetchBatchSize != 1 {
		t.Fatalf("scheduler config = %#v, want concurrency/fetch batch size normalized to 1", cfg.Scheduler)
	}
}
