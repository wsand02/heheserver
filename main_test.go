package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestStartupFailure(t *testing.T) {
	if os.Getenv("HEHE_TEST_STARTUP") == "1" {
		os.Args = []string{"heheserver", "-host", "invalid:host", os.Getenv("HEHE_TEST_ROOT")}
		main()
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestStartupFailure$")
	cmd.Env = append(os.Environ(), "HEHE_TEST_STARTUP=1", "HEHE_TEST_ROOT="+t.TempDir())
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatal("startup did not terminate")
	}
	if err == nil || !strings.Contains(string(out), "too many colons") {
		t.Fatalf("expected nonzero exit and bind error, got %v: %s", err, out)
	}
}
