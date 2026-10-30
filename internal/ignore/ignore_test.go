package ignore

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/wsand02/heheserver/internal/cache"
)

func TestOutsideRootTerminates(t *testing.T) {
	if os.Getenv("HEHE_TEST_OUTSIDE") == "1" {
		if err := cache.NewIgnoreCache(1); err != nil {
			t.Fatal(err)
		}
		root := t.TempDir()
		rules := GetIgnoreForPath(root, filepath.Dir(root))
		if !Matches(rules, "secret.txt") {
			t.Fatal("outside-root rules must deny access")
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestOutsideRootTerminates$")
	cmd.Env = append(os.Environ(), "HEHE_TEST_OUTSIDE=1")
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatal("ignore traversal failed to terminate outside root")
	}
	if err != nil {
		t.Fatalf("outside-root traversal: %v: %s", err, out)
	}
}

func TestGetIgnoreForPathRace(t *testing.T) {
	root := t.TempDir()
	cache.NewIgnoreCache(16)
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
