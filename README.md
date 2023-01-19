# heheserver

A simple barebones http file server. Written in Go.

Made it because the built-in Python webserver didn't properly support serving videos.

And I didn't want to use the bloated http-server in the node ecosystem.

## Installation

### Install from source

```
git clone https://github.com/wsand02/heheserver
cd heheserver
go build
```

Then copy the generated executable to a directory of your choosing.

If you want to skip those steps and you aren't in a directory with a go.mod file. You can use:

```
go install github.com/wsand02/heheserver@v0.0.1
```

## Usage

```
heheserver [options] [path]
```

## Available options

`-p` or `--port` The port the server will run on. (default 3400)

`-a` or `--address` The address the server will run on. (default 0.0.0.0)

## License

This software is licensed under the MIT License.