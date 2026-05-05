package admin

import (
	"bytes"
	"encoding/json"
	"strings"
)

func normalizeJSON(raw json.RawMessage, invalidErr error) (string, error) {
	if !json.Valid(raw) {
		return "", invalidErr
	}

	var buffer bytes.Buffer
	if err := json.Compact(&buffer, raw); err != nil {
		return "", invalidErr
	}

	return buffer.String(), nil
}

func normalizeJSONObject(raw json.RawMessage, invalidErr error) (string, error) {
	normalized, err := normalizeJSON(raw, invalidErr)
	if err != nil {
		return "", err
	}

	var value map[string]any
	if err := json.Unmarshal([]byte(normalized), &value); err != nil {
		return "", invalidErr
	}

	return normalized, nil
}

func marshalJSONObjectString(raw, fallbackKey string) json.RawMessage {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return json.RawMessage(`{}`)
	}

	normalized, err := normalizeJSONObject(json.RawMessage(trimmed), errInvalidDescription)
	if err == nil {
		return json.RawMessage(normalized)
	}

	legacy, marshalErr := json.Marshal(map[string]string{fallbackKey: raw})
	if marshalErr != nil {
		return json.RawMessage(`{}`)
	}

	return json.RawMessage(legacy)
}
