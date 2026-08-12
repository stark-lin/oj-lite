// Embeds frontend pages and shared static assets for direct server responses.

package app

import "embed"

var (
	//go:embed html/*.html html/*.js html/pages/*.css html/ui/*.css html/ui/*.js html/vendor/*.js
	htmlFS embed.FS
)

func readEmbeddedHTML(name string) ([]byte, error) {
	return htmlFS.ReadFile("html/" + name)
}

func readEmbeddedAsset(name string) ([]byte, error) {
	return htmlFS.ReadFile("html/" + name)
}
