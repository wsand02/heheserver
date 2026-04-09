package fs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenIgnoreFile(t *testing.T) {
	dir := t.TempDir()
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

func TestReaddirIgnoreFile(t *testing.T) {
	dir := t.TempDir()
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
