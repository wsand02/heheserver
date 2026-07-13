package ignore

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	ignore "github.com/wsand02/go-gitignore"
	"github.com/wsand02/heheserver/internal/cache"
)

// GetIgnoreForPath returns the combined ignore rules for the specified path.
// It walks from the target directory up to root, collecting each level's
// .heheignore lines, then compiles them into a single GitIgnore ordered
// root-first so that negation (!pattern) is honored across files, matching
// gitignore's last-match-wins semantics.
func GetIgnoreForPath(root, path string) *ignore.GitIgnore {
	root = filepath.Clean(root)
	dir := filepath.Clean(path)

	var perDir [][]string

	for {
		lines, ok := cache.GetIgnoreCache().Get(dir)
		if !ok {
			bs, err := os.ReadFile(filepath.Join(dir, ".heheignore"))
			if err != nil {
				lines = nil
			} else {
				lines = strings.Split(string(bs), "\n")
			}
			cache.GetIgnoreCache().Set(dir, lines, 1)
		}

		if lines != nil {
			perDir = append(perDir, lines)
		}

		if dir == root || dir == "." || dir == "/" {
			break
		}

		dir = filepath.Dir(dir)
	}

	// perDir is deepest-first; reverse so the root file's lines come first.
	slices.Reverse(perDir)

	var all []string
	for _, lines := range perDir {
		all = append(all, lines...)
	}

	return ignore.CompileIgnoreLines(all...)
}

// Matches reports whether path is ignored by the combined rules. A nil ruleset
// (no .heheignore anywhere up the tree) matches nothing.
func Matches(rules *ignore.GitIgnore, path string) bool {
	if rules == nil {
		return false
	}
	return rules.MatchesPath(path)
}
