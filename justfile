version := `git describe --tags --abbrev=0 2>/dev/null || cat flake.nix 2>/dev/null | grep -oP 'version.*=.*' | head -1 | cut -d'"' -f2 | cut -d'"' -f1 || echo "0.1.0"`
commit := `git rev-parse --short HEAD 2>/dev/null || echo "unknown"`

# List available recipes
default:
    @just --list

# Fail with one actionable line instead of a wall of cgo errors.
_require-vips:
    #!/usr/bin/env bash
    if ! pkg-config --exists vips 2>/dev/null; then
        echo "libvips not found on PKG_CONFIG_PATH. internal/imaging binds it over CGo." >&2
        echo "Run inside the dev shell:  nix develop --command just <recipe>" >&2
        exit 1
    fi

# Build frontend and copy into the embedded dashboard dist
frontend:
    #!/usr/bin/env bash
    set -euo pipefail
    (cd frontend && bun install --frozen-lockfile && bun run build)
    # stage then rename: a failed copy must not leave a half populated dist
    # that -tags release would happily embed
    rm -rf internal/dashboard/.dist.next internal/dashboard/dist
    cp -r frontend/dist internal/dashboard/.dist.next
    mv internal/dashboard/.dist.next internal/dashboard/dist

# Build the release binary with the dashboard embedded
build: _require-vips frontend
    go build -tags release -ldflags "-s -w -X main.version={{version}} -X main.commit={{commit}}" -o auxio-bin ./cmd/auxio/

# Start frontend (vite) and backend (wgo live reload) together
dev: _require-vips
    #!/usr/bin/env bash
    set -euo pipefail
    # --frozen-lockfile is a 4ms no-op once satisfied, so there is nothing to
    # guard it with and a fresh clone needs no separate install step
    (cd frontend && bun install --frozen-lockfile)

    # each child gets its own process group, so a Ctrl-C aimed at the foreground
    # group reaches just and this shell but not the servers - cleanup owns them
    set -m
    pids=()
    cleanup() {
        local status=$? pid
        trap - EXIT INT TERM
        # plain TERM per pid, never kill -- -pid here: wgo runs the server it
        # builds in a process group of its own and reaps it on TERM, but a group
        # kill takes wgo out first and orphans the server still holding :9000
        for pid in "${pids[@]}"; do kill "$pid" 2>/dev/null || true; done
        sleep 1
        # backstop for anything that slept through the TERM
        for pid in "${pids[@]}"; do kill -- -"$pid" 2>/dev/null || true; done
        wait "${pids[@]}" 2>/dev/null || true
        exit "$status"
    }
    trap cleanup EXIT INT TERM

    (cd frontend && exec bun run dev) &
    pids+=($!)
    (exec env AUXIO_ACCESS_KEY=test AUXIO_SECRET_KEY=secret AUXIO_IMAGING_ENABLED=true \
        go tool wgo run -exit -xdir 'frontend|data|internal/dashboard/dist|0_notes|docs|docs-full' ./cmd/auxio/) &
    pids+=($!)

    # wait -n returns on the first child to exit, so a dead backend takes vite
    # down with it instead of leaving the proxy answering ECONNREFUSED
    set +e
    wait -n "${pids[@]}"
    status=$?
    set -e
    echo "just dev: a process exited (status $status), shutting down the other" >&2
    exit "$status"

# Run the frontend dev server only (vite)
dev-frontend:
    cd frontend && bun run dev

# Run the backend dev server only (go)
dev-server: _require-vips
    AUXIO_ACCESS_KEY=test AUXIO_SECRET_KEY=secret AUXIO_IMAGING_ENABLED=true go run ./cmd/auxio/

# Prove `just dev` works from a fresh clone
smoke:
    ./scripts/smoke-dev.sh

# Remove build artifacts and node_modules
clean:
    rm -f auxio-bin
    rm -rf frontend/dist internal/dashboard/dist internal/dashboard/.dist.next
    rm -rf frontend/node_modules

# Rebuild the search index
rebuild-index:
    go run ./cmd/auxio/ --rebuild-index

# Run Go tests
test:
    go test ./...

# Regenerate the gomod2nix lock file
nix-lock:
    gomod2nix generate
