# Auxio

## About


A single-binary S3-compatible object server with on-demand image processing and a web
dashboard.

Auxio is a personal replacement for MinIO made with prototyping in mind. It implements the subset of the S3 API. No enterprise focus - only minimal set of features. Every object stays on a filesystem.



- **S3 REST API** with AWS Signature Version 4, including presigned GET and PUT URLs
- **Image transforms from the URL** - `?w=800&fmt=webp` resizes and converts on request
  (libvips), then caches the result on disk
- **Files on the filesystem.** Objects are plain files with a JSON sidecar next to them.
  SQLite is only an index and can be rebuilt from disk.
- **Embedded Vue dashboard** at `/dashboard/` for buckets, browsing, and uploads
- **NixOS module** with a hardened systemd unit

## Scope and non-goals

Auxio is built for a **single operator** running it **behind a TLS-terminating reverse
proxy** such as Caddy. It is not multi-tenant and has no IAM, STS, or policy engine.

- **Loopback by default.** `AUXIO_BIND` is `127.0.0.1`. Use a reverse proxy.
- **The startup guard only inspects the bind address.** If your proxy exposes Auxio to the
  internet, the guard cannot detect it - configure real credentials.
- **Payload bodies are not re-hashed.** `x-amz-content-sha256` is covered by the SigV4
  signature, but the server does not recompute the body hash.


**Public buckets.** A public bucket serves its objects by key: a client that
knows the key needs no credentials, which is what a prototype needs to serve media
directly. Upload an object, use its URL. The object list is not published. **Enumerating a bucket always requires credentials**, public or not.

**Not implemented, not in scope by design:** object versioning, lifecycle rules, ACL XML and bucket
policies, IAM/STS, cross-region replication, bucket notifications, 
object lock, CopyObject.

## Quick start

`just`, `go` and `bun` are the only prerequisites, plus libvips - the image processor
binds it over CGo, so `nix develop` is the shortest way to get it.

```sh
just dev                    # Vite on :5173, Go on :9000, both reload
just build && ./auxio-bin   # single binary with the dashboard embedded
```

`just dev` installs frontend dependencies itself and runs the two servers. The dashboard is at
<http://localhost:5173/dashboard/> in development and
<http://localhost:9000/dashboard/> from the built binary. Default login is
`admin` / `password`; the default S3 keys are `auxio` / `auxio-secret-key` and produce a
warning at startup.

Only `just build` embeds the dashboard, via `-tags release`. A plain `go build` produces a
working server whose `/dashboard` is a placeholder.

## Documentation

[docs/reference.md](docs/reference.md) - build and quick start, configuration, the S3 API,
image transforms, the dashboard, internals, NixOS, development.

## License

MIT. See [LICENSE](LICENSE).
