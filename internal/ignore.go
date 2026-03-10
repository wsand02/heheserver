package internal

import (
	"path/filepath"

	ignore "github.com/wsand02/go-gitignore"
)

var ignoreCache = map[string]*ignore.GitIgnore{}

func getIgnoreForPath(root, path string) []*ignore.GitIgnore {
	var rules []*ignore.GitIgnore

	dir := filepath.Clean(path)
	for {
		if ig, ok := ignoreCache[dir]; ok {
			if ig != nil {
				rules = append([]*ignore.GitIgnore{ig}, rules...)
			}
		} else {
			ig, _ := ignore.CompileIgnoreFile(filepath.Join(dir, ".heheignore"))
			ignoreCache[dir] = ig
			if ig != nil {
				rules = append([]*ignore.GitIgnore{ig}, rules...)
			}
		}

		if dir == root || dir == "." || dir == "/" {
			break
		}
		dir = filepath.Dir(dir)
	}
	return rules
}

func matchesAny(rules []*ignore.GitIgnore, path string) bool {
	for _, r := range rules {
		if r.MatchesPath(path) {
			return true
		}
	}
	return false
}
