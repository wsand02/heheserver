# heheserver
[![Go](https://github.com/wsand02/heheserver/actions/workflows/go.yml/badge.svg)](https://github.com/wsand02/heheserver/actions/workflows/go.yml)
![GitHub Tag](https://img.shields.io/github/v/tag/wsand02/heheserver)

A readonly http file server with file omission using heheignore files.
Written in Go.

Made it because the built-in Python web server didn't properly support serving videos,
and I didn't want to use the bloated http-server in the node ecosystem.

Primarily intended for quick file sharing on trusted local networks. It does not implement
authentication or other security features required for safe exposure to the public internet.

## Features
- Lightweight, read-only HTTP file server
- Single static binary
- `.heheignore` file omission — matching files are hidden from every directory index *and* made un-openable, as if they don't exist
- Optional embedded gallery for images, video, and audio, with thumbnails and a single-item post view
- Filter the gallery by file type, filename, or extension — updates in realtime, with shareable filter URLs
- Thumbnail-only video tiles in the gallery — playback happens in the post view, so long browsing sessions stay light
- Infinite scroll for large directories, with an option to force pagination
- Optional on-the-fly image resizing (pure-Go fallback, accelerated by ffmpeg when present)
- Video thumbnails when ffmpeg is available

## Requirements

Go is all you need to build and run heheserver.

[ffmpeg](https://ffmpeg.org/) is an **optional** dependency: if it's on your `PATH`, it's used
to accelerate image resizing and to generate video thumbnails. Without it, image resizing still
works via a pure-Go fallback and video thumbnails are simply disabled.

## Installation

### Install from source

```
git clone https://github.com/wsand02/heheserver
cd heheserver
go build
```

Then copy the generated executable to a directory of your choosing.

If you want to skip those steps and you aren't in a directory with a go.mod file, you can use:

```
go install github.com/wsand02/heheserver@latest
```

## Usage

```
heheserver [options] [path]
```

## Available options

`-p` or `-port` The port the server will run on. (default 3400)

`-h` or `-host` The host the server will run on. (default 0.0.0.0)

`-g` or `-gallery` Enables the embedded gallery page. (default omitted => false)

`-r` or `-resize` Enables the image resizing endpoint. Uses ffmpeg if it's on your `PATH`, otherwise a pure-Go fallback. When ffmpeg is present this also enables video thumbnails. (default omitted => false)

`-s` or `-split` Max items per page for gallery pagination. (default 48)

`-no-infinite-scroll` Disable infinite scroll in the gallery; always use paginated navigation even with JavaScript enabled. (default omitted => false)

`-igncache` Size of ignore cache in megabytes, approximate. (default 16)

`-rescache` Size of resize cache in megabytes, approximate. (default 1000)

`-vidtcache` Size of video thumbnail cache in megabytes, approximate. (default 1000)

## Startup output

On launch heheserver prints its version and the URLs it's reachable at — a `localhost` link
plus, when bound to all interfaces, a LAN link — so you can click straight through from the
terminal:

```
heheserver vX.X.X
Serving ./
  http://localhost:3400
  http://192.168.1.42:3400
```

## Heheignore
Basically just gitignore, but omits matching files from all directory indexes while also making the files appear as if they don't exist when you try to access them.

Files are only read once so if you change a heheignore file that has already been read you will need to restart the server.

Supports subdirectory heheignore files.

## Gallery view

Enable the React gallery with `-g`. JavaScript is required. The gallery and its dependencies
are embedded in the executable and work without a CDN or external network access. Plain
file-server mode (without `-g`) remains available.

Add `-r` for resized image previews and, when ffmpeg is available, video thumbnails.
Browse using breadcrumbs and filename, type, and extension filters. Filters apply to the
entire directory before pagination. Results use deterministic filename ordering.

Click an image, video, or audio tile to open the viewer. Previous/next controls and keyboard
arrows move through matching media in the same directory, including items on other pages.
Escape, Close, or browser Back returns to the gallery without losing loaded tiles or scroll
position. Direct `/post/?path=...` links open a standalone viewer. Playback never starts
automatically, and leaving an item releases its media resources.

Directories load more items as you scroll. Use `-no-infinite-scroll` for explicit pagination;
`-s` controls page size. Theme selection supports system, light, and dark appearances.

Supported images: `.jpg`, `.jpeg`, `.png`, `.webp`, `.svg`, `.gif` (including animation); video: `.mov`, `.mp4`, `.m4v`,
`.webm`; audio: `.mp3`, `.wav`, `.ogg`, `.m4a`. Browser codec support determines playback.

## Frontend development

Go builds and `go install` use committed production assets in `internal/frontend/dist/`.
They do not invoke Node or pnpm. Frontend contributors need the Node version in
`web/.node-version` and the pnpm version pinned in `web/package.json`.

```sh
cd web
pnpm install --frozen-lockfile
pnpm dev
```

Run the Go server separately with `go run . -g -r /path/to/media` on port 3400. Vite proxies
API, raw-file, and thumbnail requests to that server. Its default dev URL is localhost:5173.

```sh
pnpm typecheck
pnpm test
pnpm build
pnpm exec playwright install chromium
pnpm test:browser
```

Commit frontend source, `web/pnpm-lock.yaml`, and rebuilt `internal/frontend/dist/` together.
CI rebuilds and checks for differences, then runs browser tests on desktop and phone-sized
viewports. Frontend output is bundled locally; source maps are not shipped.

Dependency-only PRs in this repository (Dependabot or manual updates to `web/package.json`
and `web/pnpm-lock.yaml`) automatically rebuild changed assets. A read-only build uploads
the generated bundle; a separate workflow commits only `internal/frontend/dist/` back to
the PR and explicitly starts CI on that commit. No personal access token is needed.
Source changes and fork PRs still require contributors to include their own rebuilt assets.

The publishing workflow becomes active after these workflows are merged into `master`.
For dependency PRs already open at that point, run **Rebuild frontend bundle** from the
Actions tab and enter the PR number. Older branches are updated from `master` when needed
to inherit dispatchable CI, then rebuilt. Unchanged bundles produce no extra commit.

Read-only JSON endpoints `/api/gallery` and `/api/media` use the existing `path`, `type`,
`q`, `ext`, and `p` query parameters. Directory listings include pagination and media URLs;
media responses include adjacent viewer URLs and a gallery-return URL. File access and
listings share `.heheignore` enforcement.

## License

This software is licensed under the MIT License. See [LICENSE](LICENSE) for details.
