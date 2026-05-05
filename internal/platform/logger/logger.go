// Initializes the shared logger and exposes common project logging helpers.

package logger

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	domain string
	base   *log.Logger
}

func NewLogger(domain string) *Logger {
	return &Logger{
		domain: domain,
		base:   log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds),
	}
}

func (logger *Logger) PrintLog(content string) {
	logger.Infof("%s", content)
}

func (logger *Logger) Debugf(format string, args ...any) {
	logger.logf("DEBUG", format, args...)
}

func (logger *Logger) Infof(format string, args ...any) {
	logger.logf("INFO", format, args...)
}

func (logger *Logger) Warnf(format string, args ...any) {
	logger.logf("WARN", format, args...)
}

func (logger *Logger) Errorf(format string, args ...any) {
	logger.logf("ERROR", format, args...)
}

func (logger *Logger) logf(level, format string, args ...any) {
	logger.base.Printf("[%s] [%s] %s", logger.domain, level, fmt.Sprintf(format, args...))
}
