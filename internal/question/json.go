package question

import (
	"bytes"
	"encoding/json"
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
