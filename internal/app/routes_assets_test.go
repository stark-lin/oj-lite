package app

import (
	"bytes"
	"encoding/json"
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

	pageStyles := []struct {
		path   string
		marker []byte
	}{
		{path: "/assets/pages/admin.css", marker: []byte(".admin-page")},
		{path: "/assets/pages/login.css", marker: []byte(".login-page")},
		{path: "/assets/pages/student.css", marker: []byte(".student-page")},
		{path: "/assets/pages/teacher.css", marker: []byte(".teacher-page")},
	}
	for _, asset := range pageStyles {
		response := performRequest(t, app, http.MethodGet, asset.path, nil, nil)
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d", asset.path, response.Code, http.StatusOK)
		}
		if contentType := response.Header().Get("Content-Type"); contentType != "text/css; charset=utf-8" {
			t.Fatalf("GET %s content type = %q", asset.path, contentType)
		}
		if !bytes.Contains(response.Body.Bytes(), asset.marker) {
			t.Fatalf("GET %s body missing marker %q", asset.path, asset.marker)
		}
	}

	javascriptAssets := []struct {
		path   string
		marker []byte
	}{
		{path: "/assets/app.js", marker: []byte("window.OJLite")},
		{path: "/assets/components.js", marker: []byte("bindResizablePane")},
		{path: "/assets/admin.js", marker: []byte("handleCreateTeacher")},
		{path: "/assets/login.js", marker: []byte("loginForm")},
		{path: "/assets/student.js", marker: []byte("submitCurrentCode")},
		{path: "/assets/teacher.js", marker: []byte("createClassAction")},
	}
	for _, asset := range javascriptAssets {
		response := performRequest(t, app, http.MethodGet, asset.path, nil, nil)
		if response.Code != http.StatusOK {
			t.Fatalf("GET %s status = %d, want %d body=%s", asset.path, response.Code, http.StatusOK, response.Body.String())
		}
		if contentType := response.Header().Get("Content-Type"); contentType != "application/javascript; charset=utf-8" {
			t.Fatalf("GET %s content type = %q", asset.path, contentType)
		}
		if !bytes.Contains(response.Body.Bytes(), asset.marker) {
			t.Fatalf("GET %s body missing marker %q", asset.path, asset.marker)
		}
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

func TestPagesLoadDedicatedScriptsAfterSharedApp(t *testing.T) {
	pages := map[string]struct {
		style  string
		script string
	}{
		"admin.html":   {style: "/assets/pages/admin.css", script: "/assets/admin.js"},
		"login.html":   {style: "/assets/pages/login.css", script: "/assets/login.js"},
		"student.html": {style: "/assets/pages/student.css", script: "/assets/student.js"},
		"teacher.html": {style: "/assets/pages/teacher.css", script: "/assets/teacher.js"},
	}

	for pageName, assets := range pages {
		page, err := readEmbeddedHTML(pageName)
		if err != nil {
			t.Fatalf("read %s: %v", pageName, err)
		}

		sharedStyleIndex := bytes.Index(page, []byte(`/assets/app.css`))
		pageStyleIndex := bytes.Index(page, []byte(assets.style))
		appIndex := bytes.Index(page, []byte(`/assets/app.js`))
		componentsIndex := bytes.Index(page, []byte(`/assets/components.js`))
		pageScriptIndex := bytes.Index(page, []byte(assets.script))
		if sharedStyleIndex < 0 || pageStyleIndex < sharedStyleIndex {
			t.Fatalf("%s must load %s after app.css", pageName, assets.style)
		}
		if appIndex < 0 || componentsIndex < appIndex || pageScriptIndex < componentsIndex {
			t.Fatalf("%s must load components.js between app.js and %s", pageName, assets.script)
		}
		if bytes.Contains(page, []byte("<script>")) {
			t.Fatalf("%s must not contain inline JavaScript", pageName)
		}
		if bytes.Contains(page, []byte("<style>")) || bytes.Contains(page, []byte(" style=")) {
			t.Fatalf("%s must not contain inline CSS", pageName)
		}
	}
}

func TestSharedComponentsStaySeparatedFromAppRuntime(t *testing.T) {
	appJS, err := readEmbeddedAsset("app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	componentsJS, err := readEmbeddedAsset("ui/components.js")
	if err != nil {
		t.Fatalf("read ui/components.js: %v", err)
	}

	for _, marker := range [][]byte{[]byte("function listItem"), []byte("function createColumnResizer")} {
		if bytes.Contains(appJS, marker) {
			t.Fatalf("app.js unexpectedly contains UI component marker %q", marker)
		}
		if !bytes.Contains(componentsJS, marker) {
			t.Fatalf("ui/components.js missing UI component marker %q", marker)
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

func TestPagesLoadConfiguredAppNameFromHealthz(t *testing.T) {
	app := newTestApp(t)
	defer shutdownTestApp(t, app)

	healthResponse := performRequest(t, app, http.MethodGet, "/healthz", nil, nil)
	if healthResponse.Code != http.StatusOK {
		t.Fatalf("GET /healthz status = %d, want %d body=%s", healthResponse.Code, http.StatusOK, healthResponse.Body.String())
	}

	var healthPayload struct {
		Data struct {
			Service string `json:"service"`
		} `json:"data"`
	}
	if err := json.Unmarshal(healthResponse.Body.Bytes(), &healthPayload); err != nil {
		t.Fatalf("decode GET /healthz response: %v", err)
	}
	if healthPayload.Data.Service != app.Config().App.Name {
		t.Fatalf("GET /healthz service = %q, want %q", healthPayload.Data.Service, app.Config().App.Name)
	}

	pages := map[string]string{
		"login.html":   "Login",
		"admin.html":   "Admin Workspace",
		"teacher.html": "Teacher Workspace",
		"student.html": "Student Workspace",
	}
	for name, title := range pages {
		page, err := readEmbeddedHTML(name)
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		if bytes.Contains(page, []byte("OJ Lite")) {
			t.Fatalf("%s must not contain a hard-coded app name", name)
		}
		if !bytes.Contains(page, []byte(`data-app-title="`+title+`"`)) {
			t.Fatalf("%s body missing runtime app title metadata", name)
		}
	}

	loginPage, err := readEmbeddedHTML("login.html")
	if err != nil {
		t.Fatalf("read login.html: %v", err)
	}
	if !bytes.Contains(loginPage, []byte("data-app-name hidden")) {
		t.Fatal("login.html app name must remain hidden until /healthz succeeds")
	}

	appJS, err := readEmbeddedAsset("app.js")
	if err != nil {
		t.Fatalf("read app.js: %v", err)
	}
	for _, marker := range [][]byte{
		[]byte("fetch('/healthz'"),
		[]byte("payload?.data?.service"),
		[]byte("element.hidden = !appName"),
	} {
		if !bytes.Contains(appJS, marker) {
			t.Fatalf("app.js missing runtime app name marker %q", marker)
		}
	}
}
