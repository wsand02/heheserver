package ignore

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
)

func TestGetIgnoreForPathRace(t *testing.T) {
	root := t.TempDir()
	cache.NewIgnoreCache()
	sub := filepath.Join(root, "sub")
	os.MkdirAll(sub, 0755)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			GetIgnoreForPath(root, sub)
		}()
	}
	wg.Wait()
}
