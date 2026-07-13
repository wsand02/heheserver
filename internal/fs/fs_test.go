package fs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
)

func TestOpenIgnoreFile(t *testing.T) {
	dir := t.TempDir()
	cache.NewIgnoreCache(16)
	os.WriteFile(filepath.Join(dir, ".heheignore"), []byte("secret.txt\n"), 0644)
	os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("shh"), 0644)
	os.WriteFile(filepath.Join(dir, "public.txt"), []byte("hi"), 0644)

	hfs := Dir(dir)

	_, err := hfs.Open("/secret.txt")
	if err == nil {
		t.Fatal("expected error when opneing ignored file secret.txt")
	}

	public, err := hfs.Open("/public.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer public.Close()
}

func TestOpenRecycleBin(t *testing.T) {
	dir := t.TempDir()
	cache.NewIgnoreCache(16)
	os.WriteFile(filepath.Join(dir, ".heheignore"), []byte("\\$RECYCLE.BIN\n"), 0644)
	os.WriteFile(filepath.Join(dir, "$RECYCLE.BIN"), []byte("shh"), 0644)

	hfs := Dir(dir)

	_, err := hfs.Open("/$RECYCLE.BIN")
	if err == nil {
		t.Fatal("expected error when opening ignored file $RECYCLE.BIN")
	}
}

func TestOpenErr(t *testing.T) {
	dir := t.TempDir()
	hfs := Dir(dir)
	_, err := hfs.Open("/test.txt")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSubDirectory(t *testing.T) {
	dir := t.TempDir()
	cache.NewIgnoreCache(16)
	os.WriteFile(filepath.Join(dir, ".heheignore"), []byte("secret.txt\n"), 0644)
	os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("shh"), 0644)
	os.WriteFile(filepath.Join(dir, "public.txt"), []byte("hi"), 0644)
	subdir := filepath.Join(dir, "subdir")
	os.Mkdir(subdir, 0644)
	os.WriteFile(filepath.Join(subdir, ".heheignore"), []byte("no.txt"), 0755)
	os.WriteFile(filepath.Join(dir, "no.txt"), []byte("yes"), 0644)
	os.WriteFile(filepath.Join(subdir, "no.txt"), []byte("yes"), 0644)
	hfs := Dir(dir)

	_, err := hfs.Open("subdir/no.txt")
	if err == nil {
		t.Fatal("expected error")
	}

	file, err := hfs.Open("no.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
}

func TestOpenNegatedAcrossFiles(t *testing.T) {
	dir := t.TempDir()
	cache.NewIgnoreCache(16)
	os.WriteFile(filepath.Join(dir, ".heheignore"), []byte("*\n"), 0644)
	subdir := filepath.Join(dir, "subdir")
	os.Mkdir(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, ".heheignore"), []byte("!keep.txt\n"), 0644)
	os.WriteFile(filepath.Join(subdir, "keep.txt"), []byte("keep"), 0644)
	os.WriteFile(filepath.Join(subdir, "other.txt"), []byte("nope"), 0644)

	hfs := Dir(dir)

	keep, err := hfs.Open("subdir/keep.txt")
	if err != nil {
		t.Fatalf("expected subdir/keep.txt to be re-included by negation, got %v", err)
	}
	keep.Close()

	if _, err := hfs.Open("subdir/other.txt"); err == nil {
		t.Fatal("expected subdir/other.txt to remain ignored")
	}
}

func TestOpenNegatedSameFile(t *testing.T) {
	dir := t.TempDir()
	cache.NewIgnoreCache(16)
	os.WriteFile(filepath.Join(dir, ".heheignore"), []byte("*\n!keep.txt\n"), 0644)
	os.WriteFile(filepath.Join(dir, "keep.txt"), []byte("keep"), 0644)
	os.WriteFile(filepath.Join(dir, "other.txt"), []byte("nope"), 0644)

	hfs := Dir(dir)

	keep, err := hfs.Open("keep.txt")
	if err != nil {
		t.Fatalf("expected keep.txt to be re-included by negation, got %v", err)
	}
	keep.Close()

	if _, err := hfs.Open("other.txt"); err == nil {
		t.Fatal("expected other.txt to remain ignored")
	}
}

func TestReaddirIgnoreFile(t *testing.T) {
	dir := t.TempDir()
	cache.NewIgnoreCache(16)
	os.WriteFile(filepath.Join(dir, ".heheignore"), []byte("secret.txt\n"), 0644)
	os.WriteFile(filepath.Join(dir, "secret.txt"), []byte("shh"), 0644)
	os.WriteFile(filepath.Join(dir, "public.txt"), []byte("hi"), 0644)

	hfs := Dir(dir)
	file, err := hfs.Open(".")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	df, err := file.Readdir(-1)
	for _, entry := range df {
		if entry.Name() == "secret.txt" {
			t.Errorf("expected secret.txt to be ignored")
		}
	}
}
