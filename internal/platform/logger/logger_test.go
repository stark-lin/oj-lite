package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestLoggerWritesStructuredJSON(t *testing.T) {
	var output bytes.Buffer
	log := newLogger("test-domain", &output)

	log.Error("request failed", "request_id", "req-123", "user_id", int64(42), "err", errors.New("boom"))

	var entry map[string]any
	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("decode log entry: %v", err)
	}

	assertLogField(t, entry, "level", "ERROR")
	assertLogField(t, entry, "msg", "request failed")
	assertLogField(t, entry, "domain", "test-domain")
	assertLogField(t, entry, "request_id", "req-123")
	assertLogField(t, entry, "user_id", float64(42))
	assertLogField(t, entry, "err", "boom")
	if _, ok := entry["time"]; !ok {
		t.Fatal("log entry is missing time field")
	}
}

func assertLogField(t *testing.T, entry map[string]any, name string, want any) {
	t.Helper()
	if got := entry[name]; got != want {
		t.Fatalf("log field %q = %#v, want %#v", name, got, want)
	}
}
