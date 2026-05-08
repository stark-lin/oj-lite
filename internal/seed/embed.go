// Embeds bundled lesson seed files for runtime database initialization.

package seed

import "embed"

var (
	//go:embed lessons/*.json
	lessonFS embed.FS
)

func readEmbeddedLesson(name string) ([]byte, error) {
	return lessonFS.ReadFile("lessons/" + name)
}
