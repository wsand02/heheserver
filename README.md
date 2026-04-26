# heheserver
[![Go](https://github.com/wsand02/heheserver/actions/workflows/go.yml/badge.svg)](https://github.com/wsand02/heheserver/actions/workflows/go.yml)

A ~~simple barebones~~ readonly http file server with file omission using heheignore files.
Written in Go.

Made it because the built-in Python web server didn't properly support serving videos,
and I didn't want to use the bloated http-server in the node ecosystem.

Primarily intended for quick file sharing on trusted local networks. It does not implement
authentication or other security features required for safe exposure to the public internet.

## Features
- Lightweight HTTP file server
- `.heheignore` support for hiding files
- Optional embedded image gallery
- Single static binary

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

`-r` or `-resize` Enables the experimental image resizing endpoint, requires ffmpeg on path. (default omitted => false)

`-s` or `-split` Max items per page for gallery pagination. (default 64)

`-igncache` Size of ignore cache in megabytes, approximate. (default 16)

`-rescache` Size of resize cache in megabytes, approximate. (default 1000)

`-vidtcache` Size of video thumbnail cache in megabytes, approximate. (default 1000)

## Heheignore
Basically just gitignore, but omits matching files from all directory indexes while also making the files appear as if they don't exist when you try to access them.

Files are only read once so if you change a heheignore file that has already been read you will need to restart the server.

Supports subdirectory heheignore files.

## Gallery view
W.I.P.

## License

This software is licensed under the MIT License. See [LICENSE](LICENSE) for details.
