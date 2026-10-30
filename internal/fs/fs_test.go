package fs

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/wsand02/heheserver/internal/cache"
)

func TestNestedIgnoreScope(t *testing.T) {
	root := t.TempDir()
	if err := cache.NewIgnoreCache(1); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{
		"photos/.heheignore": "/secret.txt\n/private/*.txt\n*.tmp\n!/keep.tmp\n!\\#keep.tmp\n",
		"photos/secret.txt":  "secret", "photos/private/data.txt": "secret",
		"photos/deep/secret.txt": "public", "photos/deep/private/data.txt": "public",
		"photos/hidden.tmp": "secret", "photos/keep.tmp": "public",
		"photos/#keep.tmp": "public",
		"other/secret.txt": "public", "other/private/data.txt": "public",
	} {
		full := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	hfs := Dir(root)
	for name, hidden := range map[string]bool{"photos/secret.txt": true, "photos/private/data.txt": true, "photos/hidden.tmp": true, "photos/keep.tmp": false, "photos/#keep.tmp": false, "photos/deep/secret.txt": false, "photos/deep/private/data.txt": false, "other/secret.txt": false, "other/private/data.txt": false} {
		t.Run(name, func(t *testing.T) {
			f, err := hfs.Open("/" + name)
			if f != nil {
				f.Close()
			}
			if hidden && !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("hidden file open: %v", err)
			}
			if !hidden && err != nil {
				t.Errorf("public file open: %v", err)
			}
			dir, err := hfs.Open("/" + filepath.ToSlash(filepath.Dir(name)))
			if err != nil {
				t.Fatal(err)
			}
			defer dir.Close()
			entries, err := dir.Readdir(-1)
			if err != nil {
				t.Fatal(err)
			}
			found := false
			for _, entry := range entries {
				if entry.Name() == filepath.Base(name) {
					found = true
				}
			}
			if found == hidden {
				t.Errorf("listed=%v, hidden=%v", found, hidden)
			}
		})
	}
}

func TestOpenPathValidation(t *testing.T) {
	root := t.TempDir()
	if err := cache.NewIgnoreCache(1); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".heheignore"), []byte("/secret.txt\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "secret.txt"), []byte("hidden"), 0600); err != nil {
		t.Fatal(err)
	}
	hfs := Dir(root)
	for _, raw := range []string{"/./secret.txt", "//secret.txt"} {
		if f, err := hfs.Open(raw); !errors.Is(err, fs.ErrNotExist) {
			if f != nil {
				f.Close()
			}
			t.Errorf("%q bypassed ignore rule: %v", raw, err)
		}
	}
	for _, raw := range []string{"/../secret.txt", "/sub/../../secret.txt", `/sub\secret.txt`, "/bad\x00.txt"} {
		if f, err := hfs.Open(raw); !errors.Is(err, fs.ErrInvalid) {
			if f != nil {
				f.Close()
			}
			t.Errorf("%q: got %v, want invalid path", raw, err)
		}
	}
}

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
