package ignore

import (
	"os"
	"path/filepath"
	"slices"
	"strings"

	gitignore "github.com/wsand02/go-gitignore"
	"github.com/wsand02/heheserver/internal/cache"
)

type scopedRules struct {
	base               string
	patterns           *gitignore.GitIgnore
	withInheritedMatch *gitignore.GitIgnore
}

// Rules retains each pattern's directory and ordering across ignore files.
type Rules struct {
	scopes []scopedRules
	deny   bool
}

// GetIgnoreForPath reads rules from root through the target directory. Paths
// outside root fail closed without reading ignore files outside the served tree.
func GetIgnoreForPath(root, path string) *Rules {
	root, err := filepath.Abs(root)
	if err != nil {
		return &Rules{deny: true}
	}
	dir, err := filepath.Abs(path)
	if err != nil {
		return &Rules{deny: true}
	}
	rel, err := filepath.Rel(root, dir)
	if err != nil || !filepath.IsLocal(rel) {
		return &Rules{deny: true}
	}
	var perDir []scopedRules

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

		base, _ := filepath.Rel(root, dir)
		base = filepath.ToSlash(base)
		if base == "." {
			base = ""
		}
		if len(lines) > 0 {
			// Seed a match when an ancestor already ignored the path, allowing
			// this file's negations to override it without reparsing patterns.
			inherited := append([]string{"*"}, lines...)
			perDir = append(perDir, scopedRules{
				base:               base,
				patterns:           gitignore.CompileIgnoreLines(lines...),
				withInheritedMatch: gitignore.CompileIgnoreLines(inherited...),
			})
		}
		parent := filepath.Dir(dir)
		if base == "" || parent == dir {
			break
		}

		dir = parent
	}

	// perDir is deepest-first; reverse so the root file's lines come first.
	slices.Reverse(perDir)

	return &Rules{scopes: perDir}
}

// Matches reports whether path is ignored by the combined rules. A nil ruleset
// (no .heheignore anywhere up the tree) matches nothing.
func Matches(rules *Rules, path string) bool {
	if rules == nil {
		return false
	}
	if rules.deny {
		return true
	}
	path = strings.TrimPrefix(filepath.ToSlash(path), "/")
	matched := false
	for _, scope := range rules.scopes {
		rel := path
		if scope.base != "" {
			var ok bool
			rel, ok = strings.CutPrefix(path, scope.base+"/")
			if !ok {
				continue
			}
		}
		if matched {
			matched = scope.withInheritedMatch.MatchesPath(rel)
		} else {
			matched = scope.patterns.MatchesPath(rel)
		}
	}
	return matched
}
