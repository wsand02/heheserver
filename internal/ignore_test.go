package internal

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestGetIgnoreForPathRace(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "sub")
	os.MkdirAll(sub, 0755)

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			getIgnoreForPath(root, sub)
		}()
	}
	wg.Wait()
}
