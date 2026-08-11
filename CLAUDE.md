# CLAUDE.md

Guidance for Claude Code when working in this repository.

## What this is

Auxio is a single-binary S3-compatible object server with an on-the-fly image
processor and an embedded Vue dashboard. Go backend, filesystem storage, SQLite
index. Module path is `github.com/protob/auxio`.

## Commands

Recipes live in the `justfile`; `just` with no argument lists them.

| Task | Command |
|---|---|
| Build the frontend into the embedded dist | `just frontend` |
| Build the whole binary | `just build` |
| Run frontend + backend together | `just dev` |
| Backend only | `just dev-server` |
| Frontend only (Vite on :5173) | `just dev-frontend` |
| Go tests | `just test` |
| Rebuild the SQLite index from disk | `just rebuild-index` |
| Prove `just dev` works from a fresh clone | `just smoke` |
| Regenerate the gomod2nix lock | `just nix-lock` |

**The dashboard is embedded only under `-tags release`.** `internal/dashboard`
is split across `embed_release.go` and `embed_dev.go`; the default build embeds
nothing and serves a notice page at `/dashboard`, which is what lets a fresh
clone `go build ./...` and `go test ./...` with no `internal/dashboard/dist`.
`just build` and the nix derivation both pass the tag, and any binary built
without it logs a warning at startup. Anything else that produces a shipped
binary has to pass it too.

`just dev` assumes only that `just`, `go` and `bun` are on PATH. It runs
`bun install --frozen-lockfile` itself, then Vite and `go tool wgo` as a pair;
if either exits the other is torn down and the recipe exits non-zero. In its
`cleanup`, children get a plain `kill "$pid"` and only then a group kill as a
backstop - wgo runs the server it builds in a process group of its own and reaps
it on TERM, but a group kill takes wgo out first and orphans the server still
holding `:9000`.

### libvips

`internal/imaging` uses govips, which binds libvips over CGo. Anything that
compiles that package - including `go vet ./...` and `go test ./tests/...` -
needs libvips on `PKG_CONFIG_PATH`. If a build fails with
`Package vips was not found`, run it inside the dev shell:

```
nix develop --command go test ./...
```

The frontend toolchain is `bun` (see `bun.lock`), not npm. `bun` is deliberately
**not** in the flake devShell - it is assumed present on the host, and so is
`node`. The devShell carries what the host does not have.

## Layout

```
cmd/auxio/          main.go: config, storage, imaging, router wiring, signals
internal/config/    AUXIO_* env vars, all defaults
internal/s3/        S3 API: SigV4 auth, routes, handlers, multipart, XML, chunked
internal/storage/   filesystem store, SQLite index, sidecars, path/name validation
internal/dashboard/ dashboard API (huma) + sessions + embedded frontend
internal/imaging/   libvips processor, on-disk cache, transform middleware
frontend/           Vue 3 + TypeScript + Vite dashboard, served under /dashboard/
nix/                NixOS module and a Caddy reverse-proxy example
tests/integration/  black-box tests driven through aws-sdk-go-v2
```

## Architecture notes

**Two routers, one mux.** `cmd/auxio/main.go` hand-builds the chi router: an S3
group behind `s3.AuthMiddleware`, a dashboard group behind
`dashboard.DashboardAuthMiddleware`, then `/dashboard/*` for the embedded SPA.
`s3.NewRouter` builds the same S3 shape for tests. The two must be kept in sync.

The S3 auth middleware reads chi route params, so it only works inside the
`Group` that registers those routes. Mounting it at mux level leaves the params
empty and a private bucket will serve anonymous GETs. There is a regression test
for exactly this in `tests/integration/semantics_test.go`.

**Storage is the filesystem; SQLite is an index.** Objects live at
`dataDir/bucket/key` with a `<key>.meta.json` sidecar; bucket settings live in
`.bucket.json`; in-progress multipart uploads live under `.uploads/<id>`. The
sidecars are authoritative and `index.db` is derived - `--rebuild-index` walks
the tree and regenerates it. Keep it that way: never make the index the only
home for a fact.

`SetMaxOpenConns(1)` serializes all queries through one connection, and pragmas
go through the `_pragma=name(value)` DSN form that `modernc.org/sqlite` expects
(not mattn's `_busy_timeout=`).

**Dashboard-only concepts.** Bucket `group` and `pinned` are dashboard
affordances. They must stay invisible to the S3 API; `tests/integration/grouping_test.go`
asserts that.

**Sessions are in-memory.** A restart forces re-login. The frontend treats that
as routine: any 401 calls `clearAuth()` and drops back to the login form.

**Config is env-only.** Everything is `AUXIO_*`, read once in
`internal/config`. There is no config file and no config endpoint - the
dashboard Settings page is a read-only mockup of the values. On a non-loopback
bind, Auxio refuses to start while the S3 keys are still the built-in defaults.

## Conventions

**Comments.** The bar is: does a competent reader of this file learn something
the code does not already say? If not, it goes. Specifics:

- Say WHY, not WHAT. Names cover the WHAT.
- No history. Comments describe the code as it is now; git holds the rest.
  Never write "previously", "the old cast", "refactored from", or a phase or
  session number.
- No marketing register: `robust`, `seamlessly`, `ensures`, `fails loudly`,
  `single source of truth`, `production-ready`, `best practice`.
- Plain ASCII hyphen `-` for a dash-shaped pause, not `—`.
- Declaration-level comments get a capital and a full stop; comments inside a
  function body are lowercase with no terminal punctuation.
- Prefer naming the concrete failure over describing it in the abstract: "you
  get a zod error at the fetch, not NaN in a table three files later".
- Three lines is a lot. Ten lines on an internal function is always wrong.
- Never touch directives (`//go:embed`, `//nolint`, `// @ts-expect-error`, …)
  or the description strings in `t.Run`/`describe`.

**Go.** Standard `gofmt`. Errors from `internal/storage` are sentinel values in
`types.go` compared with `errors.Is`. The dashboard maps them to huma errors.

**Frontend.** Every API response is parsed through a zod schema in
`api/schemas.ts`; TypeScript types are inferred from those schemas rather than
declared separately, so the runtime and static contracts cannot drift.
`reset.d.ts` makes `res.json()` return `unknown`, which means a new endpoint
physically cannot skip its schema.

The UI is **Protobiont UI**, a copy-source kit (the shadcn model - no runtime
package). `prt add <Name>` copies a component into `src/components/ui/` and
records its file hashes in `frontend/components.json`; the kit source is
external and read-only from here. `prt diff <Name>` reports every file that
has drifted from its template. Anything that differs from the kit on purpose
should stay documented somewhere findable - an undocumented difference is
indistinguishable from an accident six months later.

**No safelist, ever.** Every utility class exists as a complete literal string in
source. Joining whole literals is fine; assembling one from fragments is not. If
a style goes missing the component is wrong, never the config. Two corollaries
that bite:

- `cx()` has no tailwind-merge and neither does a Vue `:class`. Two utilities
  setting the same property at equal specificity are resolved by generated-CSS
  order, which is not a contract. Never pass a property through `class` that the
  component's variants already set - use its own props, a wrapper `div`, or put
  the alternatives in one binding so they can never coexist.
- The reset is `@unocss/reset/tailwind-compat.css`, which leaves a button its
  native background. Every chrome-less `<button>` must say `bg-transparent
  cursor-pointer` for itself.

`styles/` is `tokens.css` and a 40-line `base.css`, and that is all the
hand-written CSS in the app. `tokens.css` is the kit's rack - primitives,
semantics, the ten-pad seed rack, and an `oklch(from …)` derivation layer; light
mode is `[data-mode='light']`, not `.light`. Colours come from those custom
properties, never hard-coded hex or a literal `oklch()` at a call site, and
`oklch(from …)` is never written outside that file. **`em` is banned** -
including through named `tracking-*` utilities, which emit it by construction.
`rem` is fine. A scoped `<style>` block is a last resort: four survive, each for
a reason stated in the file.

Anything in `base.css` that a utility is meant to be able to override goes in
`:where()`. A bare element or pseudo-class rule ties with a single-class utility
and wins on order, because the app's stylesheet loads after `virtual:uno.css` -
which is how `:focus-visible` beat the `outline-none` every kit field carries in
order to opt out.

State is Pinia (`stores/`), HTTP is ofetch through `api/client.ts`, and browser
behaviour comes from VueUse before it is hand-rolled. `api/client.ts` never
exports its ofetch instance: every call goes through `request(path, schema)`, so
a new endpoint cannot return `any`. Colour mode and the toast queue are the two
exceptions to Pinia, and deliberately so: both are kit-owned module singletons
(`PrtModeToggle`, `useToast()` + `PrtToaster`). `alert` and `confirm` are not
used - `ConfirmHost` over `PrtDialog` is. A control that is unavailable for a
reason worth reading (delete on a non-empty bucket) uses `aria-disabled`, not
`disabled`, so it stays focusable and the reason is reachable; `btn/variants.ts`
carries a deliberate drift from the kit to style it.

The folder layering, which is checkable by eye in a diff:

- `ui/` is prt-installed and imports nothing from the app. Every other folder
  imports from it freely.
- `base/` is the app's own primitives - `BaseCheckbox`, `SearchInput`,
  `ConfirmHost`. It imports `lib/`, `ui/`, and (for `ConfirmHost` only)
  `stores/confirm`. Never a feature folder.
- `layout/`, `buckets/`, `objects/`, `auth/`, `settings/` import `base/`, `ui/`,
  `lib/`, `composables/`, `stores/`. Never each other.
- `views/` compose their own feature folder plus `base/`, and are the only place
  route props may appear.
- `stores/` import `api/` and `lib/`. Never a component.
- `api/` imports `lib/` and `schemas.ts`. Never a store - `stores/auth.ts`
  registers itself with `onUnauthorized()` instead, which is what keeps that
  edge pointing one way.

Tables are ordinary app components in `buckets/` and `objects/`, built from divs
with a complete ARIA role set (`table`, `row`, and a role on every direct child
of a row - a partial set is worse than none). There is no kit table component and
no `<table>` element anywhere. Rows live in their own SFCs and read their grid
from a `--cols` custom property that the card sets with `[--cols:…]` and the row
reads with `grid-cols-[var(--cols)]`. Custom properties inherit through a
component boundary and scoped-CSS ids do not, which is the whole reason that
works.

Targets are **Chromium and Firefox equally, plus their forks; Safari is not a
target.** Recent stable of either engine, no version matrix - relative colour
syntax, `<dialog closedby>`, `popover` and CSS anchor positioning are all used
freely.

**Nix.** `description =` fields in `nix/module.nix` are user-facing option
documentation, a different contract from a code comment. Format with `nixfmt`.

## Testing

`tests/integration` drives a real server over HTTP with `aws-sdk-go-v2`, so it
covers wire-level behavior that unit tests miss: presigned URLs, `aws-chunked`
framing, multipart lifecycles, anonymous access to public buckets. When
changing S3 semantics, add the test there rather than mocking the handler.

Unit tests sit next to their packages. `internal/storage/validation_test.go`
covers bucket and group name rules; `internal/s3/auth_test.go` covers signing.

## Gotchas

- The Vite dev proxy forwards `/{bucket}/{key}` to the backend but excludes
  `/@`, `/src`, `/node_modules` and `/dashboard`. Without that lookahead the
  proxy swallows Vite's own dev paths, and `/dashboard` comes back as the stale
  embedded build.
- The frontend build (`vue-tsc && vite build`) type-checks. A type error fails
  the build, so `just build` is the real check, not just `go build`.
