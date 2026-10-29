# Project guidance

heheserver is a read-only Go HTTP file server for trusted local networks. `.heheignore`
rules hide entries and deny direct access. Preserve path escaping and ignore enforcement.

## Commands

- `go build ./...`
- `go test -race ./...`
- `go run . -g -r /path/to/media`
- In `web/`: `pnpm install --frozen-lockfile`, `pnpm typecheck`, `pnpm test`, `pnpm build`, `pnpm test:browser`

## Architecture

`main.go` parses config and starts `internal/server`. Plain mode uses `http.FileServer`.
Gallery mode serves the React application at `/` and `/post/`, committed assets at `/assets/`,
raw files at `/fs/`, read-only JSON at `/api/gallery` and `/api/media`, and optional `/resize/`
and `/vidthumb/` endpoints. Unknown API, file, and asset routes must never return the UI shell.

React/TypeScript source lives in `web/`. Compiled output in `internal/frontend/dist/` is
committed and embedded in Go, so `go install` needs no frontend tooling. Rebuild and commit
the bundle whenever changing frontend source. Node and pnpm versions are pinned in `web/`.

Both APIs share ignore-aware listing, filtering before pagination, and deterministic filename
ordering. Media navigation includes only matching image/video/audio siblings, across pages.
The React overlay keeps the gallery mounted; history and focus restoration preserve browsing
state. Do not add autoplay or media players to grid tiles. Release media resources on navigation.

Caches are package-level ristretto instances initialized by `Server.initCache`. Ignore files
are cached until restart. Thumbnailing prefers ffmpeg and falls back to pure Go for images.
Path handling and special-character filename URL encoding are security-sensitive. Add a
regression test for bugs; run race checks when modifying caches or filesystem behavior.
