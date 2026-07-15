// Initializes the shared logger and exposes common project logging helpers.

package logger

import (
	"io"
	"log/slog"
	"os"
)

type Logger = slog.Logger

func NewLogger(domain string) *Logger {
	return newLogger(domain, os.Stdout)
}

func newLogger(domain string, output io.Writer) *Logger {
	return slog.New(slog.NewJSONHandler(output, nil)).With("domain", domain)
}
