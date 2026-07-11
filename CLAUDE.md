# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this is

`heheserver` is a read-only HTTP file server written in Go. Its distinguishing features are `.heheignore` support (gitignore-style file omission that also makes ignored files un-openable) and an optional embedded image/video gallery with thumbnailing. It's meant for quick sharing on trusted local networks — there is deliberately no authentication.

## Commands

```bash
go build                       # build the binary (module: github.com/wsand02/heheserver)
go build -v ./...              # build everything (matches CI)
go test -v -race ./...         # run all tests with the race detector (matches CI)
go test -race ./internal/fs    # test a single package
go test -run TestGetIgnoreForPathRace ./internal/ignore   # run a single test

./heheserver -g -r ./somedir   # run locally with gallery + resize enabled
```

CI (`.github/workflows/go.yml`) runs `go build -v ./...` and `go test -v -race ./...` on push/PR to `master`. The race detector matters here — cache access is concurrent, and there are dedicated race tests.

## Architecture

Entry point `main.go` → `config.ParseFromFlags()` → `server.NewServer(cfg)` → `s.Start()`. Everything else lives under `internal/`.

**Request flow depends on the `-gallery` flag** (`internal/server/server.go` `setupRoutes`):
- Gallery **off**: a single `http.FileServer(hfs)` serves everything.
- Gallery **on**: `/fs/` serves raw files, `/` renders the directory listing, `/post/` renders a single-item view, and (with `-resize`) `/resize/` and `/vidthumb/` serve generated thumbnails.

**Handlers receive state via a closure injector.** Gallery handlers don't match the standard `http.HandlerFunc` signature — they take `(w, r, path, *fs.HeheFS, *config.Config)`. `Server.makeHfsInjector` wraps them, pulling the target path from the `?path=` query parameter (defaulting to `/`). When editing routes/handlers, keep this signature and register through `makeHfsInjector`.

**`fs.HeheFS` is the core abstraction** (`internal/fs/fs.go`). It wraps `http.Dir` so it can be passed straight to `http.FileServer`, but overrides `Open` and `Readdir` to consult `.heheignore` rules. An ignored file returns `fs.ErrNotExist` from `Open` (so it appears not to exist, not merely hidden from listings) and is skipped in `Readdir`. `HeheFS.Root` is carried around because heheignore resolution needs the absolute root for directory traversal.

**Ignore resolution** (`internal/ignore/ignore.go`): `GetIgnoreForPath` walks from the target directory up to `Root`, collecting compiled `.heheignore` files at each level (reversed so parent rules apply first), backed by the ristretto ignore cache. Compiled ignore files are cached forever — **changing a `.heheignore` after it's been read requires a server restart** (documented behavior, not a bug). Matching uses `github.com/wsand02/go-gitignore`.

**Caches** (`internal/cache/cache.go`) are package-level singletons backed by `dgraph-io/ristretto/v2`: `ignoreCache`, `resizeCache`, `vidThumbCache`. They must be initialized (`NewXCache`) before use or the `GetXCache` accessors `log.Fatal`. `Server.initCache` sets them up conditionally based on flags (resize/vidthumb only when those features are on). Cache sizes come from CLI flags in MB; `sizeToNCMB` derives the NumCounters heuristically.

**Thumbnailing** (`internal/resize`, `internal/vidthumb`, plus the matching handlers): prefers shelling out to `ffmpeg` when it's on PATH (`utils.FFmpegExists()` is checked once at config time into `Config.FFmpegExists`). Image resize falls back to a pure-Go `golang.org/x/image/draw` path when ffmpeg is absent; video thumbnails require ffmpeg and the route is only registered when it exists.

**View layer**: `internal/models/gallery.go` (`GalleryItem` with type predicates like `IsImage`/`IsVideo` and URL builders) feeds `internal/templates` (Go `text/template` files embedded via `//go:embed *.html`). Custom template funcs for pagination/arithmetic live in `templates.go`.

## Conventions and gotchas

- **Path handling is security-sensitive.** URL/query path escaping is centralized in `models.escapeQueryPath`/`escapeURLPath`; the fs and vidthumb handlers replicate Go stdlib's `Open` localization logic (`path.Clean`, `filepath.Localize`) to avoid traversal. Preserve this when touching path construction — emoji/special-character filenames have already caused 404 regressions.
- Templates use `text/template`, not `html/template` — output is not auto-escaped.
- New file-type support usually means updating the predicate methods on `GalleryItem` and the relevant template.
- Start bug fixes with a failing test derived from the issue (see the race tests in `internal/ignore` and `internal/fs` for the pattern).
