package ignore

import (
	"path/filepath"
	"slices"
	"sync"

	ignore "github.com/wsand02/go-gitignore"
)

var (
	ignoreCache = map[string]*ignore.GitIgnore{}
	cacheMu     sync.RWMutex
)

// getIgnoreForPath returns a slice of ignore rules for the specified path,
// recursively traverses from root to the path appending all rules along the way.
func GetIgnoreForPath(root, path string) []*ignore.GitIgnore {
	root = filepath.Clean(root)
	dir := filepath.Clean(path)

	var rules []*ignore.GitIgnore

	for {
		cacheMu.RLock()
		ig, ok := ignoreCache[dir]
		cacheMu.RUnlock()

		if !ok {
			var err error
			ig, err = ignore.CompileIgnoreFile(filepath.Join(dir, ".heheignore"))
			if err != nil {
				ig = nil
			}

			cacheMu.Lock()
			ignoreCache[dir] = ig
			cacheMu.Unlock()
		}

		if ig != nil {
			rules = append(rules, ig)
		}

		if dir == root || dir == "." || dir == "/" {
			break
		}

		dir = filepath.Dir(dir)
	}

	slices.Reverse(rules)
	return rules
}

func MatchesAny(rules []*ignore.GitIgnore, path string) bool {
	for _, r := range rules {
		if r.MatchesPath(path) {
			return true
		}
	}
	return false
}
