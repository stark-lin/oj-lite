// Embeds frontend pages and shared static assets for direct server responses.

package app

import "embed"

var (
	//go:embed html/*.css html/*.html html/*.js
	htmlFS embed.FS
)

func readEmbeddedHTML(name string) ([]byte, error) {
	return htmlFS.ReadFile("html/" + name)
}

func readEmbeddedAsset(name string) ([]byte, error) {
	return htmlFS.ReadFile("html/" + name)
}
