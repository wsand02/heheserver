package internal

import (
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type HeheFS struct {
	FileSystem http.FileSystem
	Root       string
}

func Dir(root string) http.FileSystem {
	fs := http.Dir(root)

	return &HeheFS{FileSystem: fs, Root: root}
}

func (hfs HeheFS) Open(name string) (http.File, error) {
	log.Printf("Trying to open %s", name)
	rules := getIgnoreForPath(hfs.Root, filepath.Join(hfs.Root, filepath.Dir(name)))
	// Normalize path by removing leading slash for pattern matching
	normPath := name
	if len(name) > 0 && name[0] == '/' {
		normPath = name[1:]
	}
	if matchesAny(rules, normPath) {
		return nil, fs.ErrNotExist
	}

	f, err := hfs.FileSystem.Open(name)

	if err != nil {
		return nil, err
	}

	return HeheFile{f, name, hfs.Root}, nil
}

type HeheFile struct {
	http.File
	currentPath string
	root        string
}

var errMissingReadDir = errors.New("Missing ReadDir")

func (h HeheFile) Readdir(count int) ([]os.FileInfo, error) {
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
			rules := getIgnoreForPath(h.root, filepath.Join(h.root, h.currentPath))
			// Normalize path by removing leading slash for pattern matching
			pathToCheck := filepath.Join(h.currentPath, info.Name())
			if len(pathToCheck) > 0 && pathToCheck[0] == '/' {
				pathToCheck = pathToCheck[1:]
			}
			if matchesAny(rules, pathToCheck) {
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
