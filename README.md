# heheserver
[![Go](https://github.com/wsand02/heheserver/actions/workflows/go.yml/badge.svg)](https://github.com/wsand02/heheserver/actions/workflows/go.yml)

A ~~simple barebones~~ readonly http file server with file omission using heheignore files.
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
- Pagination for large directories
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

`-s` or `-split` Max items per page for gallery pagination. (default 64)

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

Enable the gallery with `-g`. Instead of the plain directory listing, heheserver renders a
thumbnail grid of the current directory with breadcrumb navigation. Images, video, and audio
are recognised and previewed inline; clicking an item opens a single-item post view for it.
Large directories are paginated (`-s` controls items per page). The gallery respects
`.heheignore`, so omitted files never appear.

Add `-r` to enable on-the-fly image resizing (used for thumbnails). If ffmpeg is on your
`PATH`, video thumbnails are generated too; otherwise image resizing falls back to a pure-Go
implementation and video thumbnails are skipped. The current version is shown in the footer,
linking back to this repository.

Supported types:

- **Images:** `.jpg` `.jpeg` `.png` `.webp` `.svg` (resizable: `.jpg` `.jpeg` `.png`)
- **Video:** `.mov` `.mp4` `.m4v` `.webm`
- **Audio:** `.mp3` `.wav` `.ogg` `.m4a`

## License

This software is licensed under the MIT License. See [LICENSE](LICENSE) for details.
