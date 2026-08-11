//go:build release

package dashboard

import (
	"embed"
	"io/fs"
)

// Embedded reports whether this binary carries the dashboard.
const Embedded = true

// A missing dist is a compile error here, and that is the point: `just build`
// runs the frontend first, so a release binary cannot ship without a dashboard.
//
//go:embed dist
var staticFS embed.FS

// GetStaticFS returns the dashboard built by `just frontend`.
func GetStaticFS() (fs.FS, error) {
	return fs.Sub(staticFS, "dist")
}
