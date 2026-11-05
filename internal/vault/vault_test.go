package vault

import (
	"os"
	"testing"
)

func TestVaultRoundTrip(t *testing.T) {
	path := "testdata_vault.orca"
	defer os.Remove(path)
	if err := Create(path, "secret"); err != nil {
		t.Fatal(err)
	}
	v, kdf, err := Load(path, "secret")
	if err != nil {
		t.Fatal(err)
	}
	v.Entries = append(v.Entries, Entry{ID: "1", Title: "github", Username: "chava", Password: "pw123"})
	if err := Save(path, "secret", v, kdf); err != nil {
		t.Fatal(err)
	}
	v2, _, err := Load(path, "secret")
	if err != nil {
		t.Fatal(err)
	}
	if len(v2.Entries) != 1 || v2.Entries[0].Password != "pw123" {
		t.Fatalf("round trip mismatch: %+v", v2)
	}
}

func TestVaultWrongPassword(t *testing.T) {
	path := "testdata_vault2.orca"
	defer os.Remove(path)
	if err := Create(path, "secret"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := Load(path, "wrong"); err == nil {
		t.Fatal("expected error with wrong password")
	}
}

func TestVaultTampered(t *testing.T) {
	path := "testdata_vault3.orca"
	defer os.Remove(path)
	if err := Create(path, "secret"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	data[len(data)-1] ^= 0xff
	os.WriteFile(path, data, 0600)
	if _, _, err := Load(path, "secret"); err == nil {
		t.Fatal("expected error on tampered vault")
	}
}
