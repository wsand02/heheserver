package ignore

import (
	"log"
	"path/filepath"
	"slices"

	ignore "github.com/wsand02/go-gitignore"
	"github.com/wsand02/heheserver/internal/cache"
)

var ignoreCache *cache.IgnoreCache

// getIgnoreForPath returns a slice of ignore rules for the specified path,
// recursively traverses from root to the path appending all rules along the way.
func GetIgnoreForPath(root, path string) []*ignore.GitIgnore {
	if ignoreCache == nil {
		iC, err := cache.NewIgnoreCache()
		if err != nil {
			log.Fatal(err)
		}
		ignoreCache = iC
	}
	root = filepath.Clean(root)
	dir := filepath.Clean(path)

	var rules []*ignore.GitIgnore

	for {

		ig, ok := ignoreCache.Get(dir)

		if !ok {
			var err error
			ig, err = ignore.CompileIgnoreFile(filepath.Join(dir, ".heheignore"))
			if err != nil {
				ig = nil
			}
			ignoreCache.Set(dir, ig, 1)
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
