package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "real.orca")
	if err := Create(target, "pw"); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link.orca")
	if err := os.Symlink(target, link); err != nil {
		t.Skip("symlink not supported")
	}
	if _, _, err := Load(link, "pw"); err == nil {
		t.Fatal("expected symlink rejection")
	}
}

func TestSaveAtomicNoPartialOnBadPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "v.orca")
	if err := Create(path, "pw"); err != nil {
		t.Fatal(err)
	}
	// writing to a path whose parent is a file should fail and not corrupt existing
	bad := filepath.Join(path, "nested.orca")
	v, kdf, _ := Load(path, "pw")
	if err := Save(bad, "pw", v, kdf); err == nil {
		t.Fatal("expected error saving into file-as-dir")
	}
	// original still readable
	if _, _, err := Load(path, "pw"); err != nil {
		t.Fatal("original vault corrupted")
	}
}

func TestValidatePathTraversal(t *testing.T) {
	if err := validatePath("a/../etc/x"); err == nil {
		t.Fatal("expected traversal rejection")
	}
	if err := validatePath("foo\x00bar"); err == nil {
		t.Fatal("expected NUL rejection")
	}
}
