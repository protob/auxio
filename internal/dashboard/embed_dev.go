//go:build !release

package dashboard

import (
	"io/fs"
	"testing/fstest"
)

// Embedded reports whether this binary carries the dashboard.
const Embedded = false

const devNotice = `<!doctype html>
<meta charset="utf-8">
<title>Auxio dashboard - development build</title>
<h1>Dashboard not embedded</h1>
<p>This binary was built without <code>-tags release</code>.</p>
<p>In development the dashboard is served by Vite at
<a href="http://localhost:5173/dashboard/">http://localhost:5173/dashboard/</a>.
Build the real one with <code>just build</code>.</p>
`

// The default build embeds nothing, so a fresh clone compiles with no
// internal/dashboard/dist. Vite owns /dashboard in development - the vite proxy
// deliberately excludes it - so this page only answers someone who points a
// browser at the backend port directly.
//
// fstest.MapFS is the standard library's only ready-made in-memory fs.FS and it
// drags in no test machinery: `go list -deps testing/fstest` does not include
// package testing.
var devFS = fstest.MapFS{
	"index.html": &fstest.MapFile{Data: []byte(devNotice)},
}

// GetStaticFS returns a one-page filesystem explaining that this is a dev build.
func GetStaticFS() (fs.FS, error) {
	return devFS, nil
}
