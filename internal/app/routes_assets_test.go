package app

import (
	"bytes"
	"net/http"
	"testing"
)

func TestEmbeddedAssetRoutes(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	cssResponse := performRequest(t, app, http.MethodGet, "/assets/app.css", nil, nil)
	if cssResponse.Code != http.StatusOK {
		t.Fatalf("GET /assets/app.css status = %d, want %d body=%s", cssResponse.Code, http.StatusOK, cssResponse.Body.String())
	}
	if !bytes.Contains(cssResponse.Body.Bytes(), []byte("--accent")) {
		t.Fatalf("GET /assets/app.css body missing expected design token")
	}

	jsResponse := performRequest(t, app, http.MethodGet, "/assets/app.js", nil, nil)
	if jsResponse.Code != http.StatusOK {
		t.Fatalf("GET /assets/app.js status = %d, want %d body=%s", jsResponse.Code, http.StatusOK, jsResponse.Body.String())
	}
	if !bytes.Contains(jsResponse.Body.Bytes(), []byte("window.OJLite")) {
		t.Fatalf("GET /assets/app.js body missing expected namespace")
	}
}
