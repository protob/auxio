# Auxio reference

Everything beyond the overview in the [README](../README.md): build, configuration, the S3
API, image transforms, the dashboard, internals, deployment and development.

## Contents

- [Quick start](#quick-start)
- [Configuration](#configuration)
- [S3 API](#s3-api)
- [Image processing](#image-processing)
- [Dashboard](#dashboard)
- [How it works](#how-it-works)
- [NixOS](#nixos)
- [Development](#development)

## Quick start

Build orchestration is defined in the [justfile](../justfile). `just` with no argument
lists every recipe.

```sh
just build        # frontend -> embed -> go build -> ./auxio-bin
./auxio-bin
```

The dashboard is embedded only under the `release` build tag, which `just build` passes.
The default build embeds nothing and serves a short notice page at `/dashboard` instead, so
a fresh clone can `go build ./...`, `go vet ./...` and `go test ./...` before the frontend
has ever been built. Every path that ships a binary has to pass the tag; a binary that was
built without it says so with a warning at startup.

Without credentials, Auxio starts on `127.0.0.1:9000` with the default keys and logs a
warning. The dashboard is at <http://localhost:9000/dashboard/>, default login
`admin` / `password`.

### With Nix

```sh
nix run .
```

The derivation passes `tags = [ "release" ]`, so it needs `internal/dashboard/dist` to
exist - run `just frontend` first. It does not build the frontend itself; nixpkgs has no
bun builder.

### Platform and build dependencies

NixOS is the development and test platform. `nix develop` provides Go, gopls, libvips,
pkg-config, gomod2nix, `just`, the `aws` CLI and `mc`. **bun** is not in the dev shell -
the frontend toolchain is expected on the host.

`internal/imaging` binds libvips over CGo, so `CGO_ENABLED=1` and the libvips headers are
required to compile or test anything that imports it, `go test ./...` included. Other Linux
distributions should work but are untested (Debian/Ubuntu: `apt install libvips-dev`).

## Configuration

Config is environment variables only. There is no config file and no config endpoint.

| Variable                     | Default            | Meaning                                            |
| ---------------------------- | ------------------ | -------------------------------------------------- |
| `AUXIO_ACCESS_KEY`           | `auxio`            | S3 access key. Set this.                           |
| `AUXIO_SECRET_KEY`           | `auxio-secret-key` | S3 secret key. Set this.                           |
| `AUXIO_USERNAME`             | `admin`            | Dashboard login                                    |
| `AUXIO_PASSWORD`             | `password`         | Dashboard password                                 |
| `AUXIO_DATA_DIR`             | `./data`           | Location of the buckets and the index              |
| `AUXIO_BIND`                 | `127.0.0.1`        | Bind address                                       |
| `AUXIO_HTTP_PORT`            | `9000`             | Listen port                                        |
| `AUXIO_REGION`               | `eu-north-1`       | Region reported in responses                       |
| `AUXIO_IMAGING_ENABLED`      | `true`             | Enable the transform middleware (requires libvips) |
| `AUXIO_IMAGING_MAX_WIDTH`    | `4096`             | Largest accepted `w`                               |
| `AUXIO_IMAGING_MAX_HEIGHT`   | `4096`             | Largest accepted `h`                               |
| `AUXIO_UPLOAD_CLEANUP_HOURS` | `24`               | Interval for removing abandoned multipart uploads  |
| `AUXIO_LOG_LEVEL`            | `info`             | slog level                                         |

An invalid value for a boolean or integer variable produces a warning, and the default is
used instead.

**Bind address.** A non-loopback bind combined with the default S3 keys aborts startup.
Loopback means `127.0.0.1`, `::1` or `localhost`. Default dashboard credentials never abort
startup, they only produce a warning. The guard inspects the bind address only: a reverse
proxy that publishes Auxio to the internet is invisible to it.

**Region.** Signature verification uses the region from the client's credential scope, not
`AUXIO_REGION`, so any region a client signs with is accepted as long as the client is
consistent.

**Flags.** `--version` prints version and commit. `--rebuild-index` regenerates `index.db`
from the sidecar files and exits.

## S3 API

Signature Version 4 in header form and in query-string (presigned) form.

| Operation               | Method   | Path                                        |
| ----------------------- | -------- | ------------------------------------------- |
| ListBuckets             | `GET`    | `/`                                         |
| CreateBucket            | `PUT`    | `/{bucket}`                                 |
| HeadBucket              | `HEAD`   | `/{bucket}`                                 |
| DeleteBucket            | `DELETE` | `/{bucket}`                                 |
| ListObjects V1/V2       | `GET`    | `/{bucket}` , `/{bucket}?list-type=2`       |
| PutObject               | `PUT`    | `/{bucket}/{key}`                           |
| GetObject               | `GET`    | `/{bucket}/{key}`                           |
| HeadObject              | `HEAD`   | `/{bucket}/{key}`                           |
| DeleteObject            | `DELETE` | `/{bucket}/{key}`                           |
| DeleteObjects           | `POST`   | `/{bucket}?delete`                          |
| CreateMultipartUpload   | `POST`   | `/{bucket}/{key}?uploads`                   |
| UploadPart              | `PUT`    | `/{bucket}/{key}?partNumber=N&uploadId=...` |
| CompleteMultipartUpload | `POST`   | `/{bucket}/{key}?uploadId=...`              |
| AbortMultipartUpload    | `DELETE` | `/{bucket}/{key}?uploadId=...`              |
| ListParts               | `GET`    | `/{bucket}/{key}?uploadId=...`              |

```sh
export AWS_ACCESS_KEY_ID=auxio
export AWS_SECRET_ACCESS_KEY=auxio-secret-key
export AWS_DEFAULT_REGION=us-east-1

aws --endpoint-url http://localhost:9000 s3 mb s3://my-bucket
aws --endpoint-url http://localhost:9000 s3 cp photo.jpg s3://my-bucket/
aws --endpoint-url http://localhost:9000 s3 ls s3://my-bucket/
```

**Addressing.** Use path-style. The `aws` CLI selects it automatically with
`--endpoint-url`; SDKs usually require an explicit `UsePathStyle` / `forcePathStyle` flag.
Virtual-host style is not supported.

**Presigned URLs.** Query-string SigV4 is verified, so any standard SDK generates presigned
**GET** and **PUT** URLs against Auxio: temporary anonymous reads of private objects, and
direct uploads from the browser, without an additional endpoint. The S3 maximum expiry of
7 days is enforced, and a timestamp in the future is rejected.

**Public buckets.** `x-amz-acl: public-read` at bucket creation is the only ACL header
Auxio honors, and the dashboard toggle sets the same flag. Anonymous `GET /{bucket}/{key}`
then succeeds. Listing always requires credentials, on public buckets as well.

Behavioral notes:

- **CopyObject is not implemented.** Clients that rename an object through a server-side
  copy (rclone, for example) revert to download plus re-upload.
- **`x-amz-meta-*` on a multipart upload is accepted but not persisted** to the finished
  object. A plain PUT preserves it.
- **Keys ending in `/` are rejected** with `InvalidObjectName`. Folders are virtual and
  materialize when the first object is uploaded under the prefix - a filesystem store
  cannot contain both a file `docs` and a directory `docs/`.
- **Payload bodies are not re-hashed.** `x-amz-content-sha256` is covered by the signature,
  but the server does not recompute the hash of the body.
- **`aws-chunked` transfer encoding is decoded**, which is what the AWS SDKs send for a
  streaming PutObject.

## Image processing

Any `GET` on an object accepts transform parameters. The middleware intercepts the request,
transforms the image with libvips, caches the bytes and serves them. A request without any
of these parameters passes through untouched.

| Parameter | Values                           | Default       | Meaning                      |
| --------- | -------------------------------- | ------------- | ---------------------------- |
| `w`       | `1` - `4096`                     | -             | Output width, px             |
| `h`       | `1` - `4096`                     | -             | Output height, px            |
| `fmt`     | `jpeg` `jpg` `png` `webp` `avif` | source format | Output format                |
| `q`       | `1` - `100`                      | `80`          | Quality, JPEG/WebP/AVIF only |
| `fit`     | `contain` `cover`                | `contain`     | Resize mode                  |

The `w` and `h` maxima come from `AUXIO_IMAGING_MAX_WIDTH` and `AUXIO_IMAGING_MAX_HEIGHT`.
`contain` scales the image to fit inside `w` x `h`; `cover` fills `w` x `h` and crops the
center. A value outside the accepted range is an error, not a clamp.

```
GET /my-bucket/photo.jpg?w=800&fmt=webp&q=85     # 800px wide WebP
GET /my-bucket/photo.jpg?w=400&h=400             # fit in 400x400, keep format
GET /my-bucket/photo.jpg?w=300&h=300&fit=cover&fmt=avif
GET /my-bucket/photo.jpg?fmt=webp                # convert only
```

Transformed bytes are written to
`{AUXIO_DATA_DIR}/.imgcache/{bucket}/{sha256(key)}/{sha256(params)}.data`, with a `.meta`
file beside it for the content type and the timestamp. One directory per object means every
variant is purged in a single operation when the object is replaced or deleted. A daily
sweep removes entries older than 30 days.

Responses include `X-Auxio-Cache: HIT` or `MISS`, and
`Cache-Control: public, max-age=31536000, immutable`.

Two limitations: the source image is decoded entirely into memory, so this assumes
image-sized files, and animated GIFs are flattened to a single frame. The stored object
keeps its animation, only the transformed variant loses it.

## Dashboard

Served from the binary at `/dashboard/`. It manages buckets (create, delete,
public/private, grouping, pinning), provides a file browser with search, sorting,
thumbnails and a details panel, accepts uploads of single files or entire folders with
relative paths preserved, including drag-and-drop, and offers copy URL / download / view /
delete per object.

Bucket `group` and `pinned` are dashboard concepts. They live in `.bucket.json` and are
invisible to the S3 API.

Keyboard shortcuts, when focus is not in a text field:

| Key   | Action                          |
| ----- | ------------------------------- |
| `u`   | Upload dialog                   |
| `n`   | New bucket (on the bucket list) |
| `/`   | Focus search                    |
| `Esc` | Close panel or dialog           |

Sessions are kept in memory, so a restart invalidates all of them. The frontend interprets
any 401 as a request to authenticate again. The Settings page displays the effective
configuration and is read-only.

## How it works

```
chi router
  Recoverer - LoggingMiddleware - StripSlashes
    |
    +-- S3 group          /{bucket}/...
    |     AuthMiddleware      verify SigV4 (header or query), allow anonymous
    |     |                   GET-by-key on public buckets
    |     ImagingMiddleware   GET with ?w/h/fmt/q/fit -> cache hit, or
    |     |                   GetObject -> govips -> cache
    |     S3 handlers         CRUD, listing, multipart
    |
    +-- Dashboard group   /api/...
    |     DashboardAuthMiddleware   Bearer token
    |     POST   /api/login              GET    /api/health
    |     GET    /api/stats              GET    /api/buckets
    |     POST   /api/buckets            PATCH  /api/buckets/{bucket}
    |     DELETE /api/buckets/{bucket}   GET    /api/buckets/{bucket}/objects
    |     DELETE /api/buckets/{bucket}/objects
    |     POST   /api/buckets/{bucket}/upload
    |     POST   /api/export/{bucket}    (streams a tar.gz)
    |
    +-- /dashboard/*      embedded Vue SPA
```

The S3 auth middleware reads chi route parameters, so it only functions inside the group
that registers those routes. Writes pass through the store, which updates the SQLite index
and then triggers a mutation hook that purges every cached variant of the object.

### On disk

```
{AUXIO_DATA_DIR}/
  index.db                       SQLite index (WAL), derived - rebuildable
  index.db-shm, index.db-wal
  .imgcache/
    {bucket}/{sha256(key)}/{sha256(params)}.data   transformed bytes
                                             .meta content type + timestamp
  .uploads/{uploadId}/           parts of in-progress multipart uploads
  .tmp/                          scratch space for atomic writes
  {bucket}/
    .bucket.json                 created_at, region, public, group, pinned
    {key}                        the raw bytes
    {key}.meta.json              etag, size, content_type, last_modified, user_metadata
```

The sidecars are authoritative, `index.db` is derived from them. After a manual edit of the
tree, execute `--rebuild-index`.

### Packages

| Package              | Responsibility                                                                     |
| -------------------- | ---------------------------------------------------------------------------------- |
| `cmd/auxio`          | Flags, wiring, background sweeps, graceful shutdown                                |
| `internal/config`    | `AUXIO_*` loading, defaults, the insecure-config guard                             |
| `internal/storage`   | `ObjectStore` interface, filesystem store, SQLite index, sidecars, name validation |
| `internal/s3`        | Routes, handlers, SigV4, XML, multipart, `aws-chunked` decoding                    |
| `internal/imaging`   | Parameter parsing, govips pipeline, disk cache, middleware                         |
| `internal/dashboard` | huma REST API, sessions, export, embedded frontend                                 |
| `frontend`           | Vue 3 + TypeScript + Vite                                                          |

Built on [govips](https://github.com/davidbyttow/govips) (libvips),
[chi](https://github.com/go-chi/chi), [huma](https://github.com/danielgtaylor/huma),
[modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite),
[google/uuid](https://github.com/google/uuid),
[go-humanize](https://github.com/dustin/go-humanize), and
[aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2) in the integration tests.

## NixOS

```nix
# flake.nix
inputs.auxio.url = "github:protob/auxio";

# configuration.nix
imports = [ auxio.nixosModules.default ];

services.auxio = {
  enable = true;
  bind = "127.0.0.1";       # expose through a reverse proxy, not directly
  port = 9000;
  dataDir = "/var/lib/auxio";

  # File holding AUXIO_ACCESS_KEY, AUXIO_SECRET_KEY,
  # AUXIO_USERNAME and AUXIO_PASSWORD.
  environmentFile = "/run/secrets/auxio";

  imaging = {
    enable = true;
    maxWidth = 4096;
    maxHeight = 4096;
  };
};
```

The unit runs under `DynamicUser` with `ProtectSystem = "strict"`, `ProtectHome`,
`NoNewPrivileges` and similar hardening options. `nix/caddy-example.nix` contains a reverse
proxy configuration to adapt.

## Development

```sh
just dev            # Vite on :5173 + Go on :9000, both reload
just dev-frontend
just dev-server
just smoke          # clone into a temp dir, `just dev`, assert both answer
just test           # requires libvips: nix develop --command just test
just rebuild-index
just nix-lock       # regenerate gomod2nix.toml
just clean
```

`just dev` needs nothing prepared. It runs `bun install --frozen-lockfile` first, which is
a few milliseconds once the lockfile is satisfied, then starts Vite and
[wgo](https://github.com/bokwoon95/wgo) as a pair. wgo rebuilds and restarts the backend on
any `.go` change and comes from the `tool` directive in `go.mod`, so there is nothing to
install. If either process exits the other is torn down and `just dev` exits non-zero -
otherwise a dead backend leaves Vite answering every `/api` call with a proxy error.

In development the dashboard is at <http://localhost:5173/dashboard/>. The Vite dev server
proxies `/api/*` and `/{bucket}/{key}` to `:9000` but deliberately not `/dashboard`, so the
backend's own `/dashboard` is never what the browser sees.

`tests/integration` drives a real server over HTTP with `aws-sdk-go-v2`. That is where
wire-level behavior is verified: presigned URLs, `aws-chunked` framing, multipart
lifecycles, anonymous access to public buckets. Unit tests reside beside their packages.
