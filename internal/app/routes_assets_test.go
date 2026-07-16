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

	vendorAssets := []struct {
		path   string
		marker []byte
	}{
		{path: "/assets/vendor/marked.min.js", marker: []byte("marked v15.0.12")},
		{path: "/assets/vendor/purify.min.js", marker: []byte("DOMPurify 3.4.12")},
	}
	for _, asset := range vendorAssets {
		response := performRequest(t, app, http.MethodGet, asset.path, nil, nil)
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", asset.path, response.Code, http.StatusOK)
		}
		if contentType := response.Header().Get("Content-Type"); contentType != "application/javascript; charset=utf-8" {
			t.Fatalf("GET %s content type = %q", asset.path, contentType)
		}
		if !bytes.Contains(response.Body.Bytes(), asset.marker) {
			t.Fatalf("GET %s body missing version marker %q", asset.path, asset.marker)
		}
	}
}

func TestMarkdownPagesLoadLocalDependenciesBeforeApp(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	for _, name := range []string{"student.html", "teacher.html"} {
		page, err := readEmbeddedHTML(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}

		markedIndex := bytes.Index(page, []byte(`/assets/vendor/marked.min.js`))
		purifyIndex := bytes.Index(page, []byte(`/assets/vendor/purify.min.js`))
		appIndex := bytes.Index(page, []byte(`/assets/app.js`))
		if markedIndex < 0 || purifyIndex < 0 || appIndex < 0 || markedIndex > appIndex || purifyIndex > appIndex {
			t.Fatalf("%s must load local marked and DOMPurify before app.js", name)
		}
		if bytes.Contains(page, []byte("cdn.jsdelivr.net")) {
			t.Fatalf("%s unexpectedly references jsDelivr", name)
		}
	}

	appJS, err := readEmbeddedAsset("app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	parseIndex := bytes.Index(appJS, []byte("parse(source"))
	sanitizeIndex := bytes.Index(appJS, []byte("sanitize(rendered"))
	if parseIndex < 0 || sanitizeIndex < parseIndex {
		t.Fatal("app.js must sanitize rendered Markdown before returning it")
	}
}
