package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	ignore "github.com/wsand02/go-gitignore"
)

const (
	portDesc    string = "The port the server will run on."
	addrDesc    string = "The address the server will run on."
	defaultPort int    = 3400
	defaultDir  string = "./"
	defaultAddr string = "0.0.0.0"
)

var heheIgnore *ignore.GitIgnore

type HeheFS struct {
	FileSystem http.FileSystem
}

func Dir(root string) http.FileSystem {
	fs := http.Dir(root)

	return &HeheFS{FileSystem: fs}
}

func (hfs HeheFS) Open(name string) (http.File, error) {
	f, err := hfs.FileSystem.Open(name)
	if err != nil {
		return nil, err
	}
	if heheIgnore != nil && heheIgnore.MatchesPath(name) {
		return nil, fs.ErrNotExist
	}
	log.Printf("Opening %v", name)
	return HeheFile{f, name}, nil
}

type HeheFile struct {
	http.File
	currentPath string // i shouldn't have to do this...
}

var errMissingReadDir = errors.New("hej hej")

func (h HeheFile) Readdir(count int) ([]os.FileInfo, error) {
	log.Print("Shower anyone?")
	d, ok := h.File.(fs.ReadDirFile)
	if !ok {
		return nil, errMissingReadDir
	}
	var list []fs.FileInfo
	for {
		dirs, err := d.ReadDir(count - len(list))
		for _, dir := range dirs {
			info, err := dir.Info()
			if err != nil {
				// Pretend it doesn't exist, like (*os.File).Readdir does.
				continue
			}

			if heheIgnore != nil && heheIgnore.MatchesPath(filepath.Join(h.currentPath, info.Name())) {
				log.Printf("Ignoring %v", info.Name())
				continue
			}
			list = append(list, info)
		}
		if err != nil {
			return list, err
		}
		if count < 0 || len(list) >= count {
			break
		}
	}
	return list, nil
}

func main() {
	// Define long flags
	port := flag.Int("port", defaultPort, portDesc)
	addr := flag.String("address", defaultAddr, addrDesc)
	// Define short flags
	flag.IntVar(port, "p", defaultPort, portDesc)
	flag.StringVar(addr, "a", defaultAddr, addrDesc)
	flag.Parse()

	dirToServe := flag.Arg(0)
	if len(dirToServe) == 0 {
		dirToServe = defaultDir
	}

	heheIgnore, _ = ignore.CompileIgnoreFile(filepath.Join(dirToServe, ".heheignore"))

	http.Handle("/", http.FileServer(Dir(dirToServe)))

	ip := fmt.Sprintf("%s:%v", *addr, *port)

	log.Printf("Serving %v on %v\n", dirToServe, ip)
	err := http.ListenAndServe(ip, nil)
	if err != nil {
		log.Fatal(err)
	}
}
