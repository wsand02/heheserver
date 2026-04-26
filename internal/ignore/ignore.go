package ignore

import (
	"path/filepath"
	"slices"

	ignore "github.com/wsand02/go-gitignore"
	"github.com/wsand02/heheserver/internal/cache"
)

// getIgnoreForPath returns a slice of ignore rules for the specified path,
// recursively traverses from root to the path appending all rules along the way.
func GetIgnoreForPath(root, path string) []*ignore.GitIgnore {
	root = filepath.Clean(root)
	dir := filepath.Clean(path)

	var rules []*ignore.GitIgnore

	for {

		ig, ok := cache.GetIgnoreCache().Get(dir)

		if !ok {
			var err error
			ig, err = ignore.CompileIgnoreFile(filepath.Join(dir, ".heheignore"))
			if err != nil {
				ig = nil
			}
			cache.GetIgnoreCache().Set(dir, ig, 1)
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
