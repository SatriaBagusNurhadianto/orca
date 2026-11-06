package vault

import (
	"os"
	"path/filepath"
	"testing"
)

func TestKDFParamsEnforcedOnLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "v.orca")
	if err := Create(path, "pw"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	// corrupt KDF memory param to below minimum (offset 6, 4 bytes)
	data[6] = 0x01
	data[7] = 0x00
	data[8] = 0x00
	data[9] = 0x00
	os.WriteFile(path, data, 0600)
	if _, _, err := Load(path, "pw"); err == nil {
		t.Fatal("expected KDF param rejection")
	}
}

func TestValidatePathTraversalComponents(t *testing.T) {
	cases := []string{"a/../etc/x", "foo/../../bar", "..", "a/.."}
	for _, c := range cases {
		if err := validatePath(c); err == nil {
			t.Fatalf("expected rejection for %s", c)
		}
	}
	if err := validatePath("clean/path"); err != nil {
		t.Fatalf("clean path rejected: %v", err)
	}
}

func TestValidatePathNUL(t *testing.T) {
	if err := validatePath("foo\x00bar"); err == nil {
		t.Fatal("expected NUL rejection")
	}
}

func TestSaveAtomicNoTempLeftover(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "v.orca")
	if err := Create(path, "pw"); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if stringsHasPrefix(e.Name(), ".vault-") {
			t.Fatalf("temp file leftover: %s", e.Name())
		}
	}
}

func stringsHasPrefix(s, pre string) bool {
	return len(s) >= len(pre) && s[:len(pre)] == pre
}
