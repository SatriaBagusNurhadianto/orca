package crypto

import (
	"bytes"
	"testing"
)

func TestDeriveKeyValidatesParams(t *testing.T) {
	p := DefaultKDFParams()
	if _, err := DeriveKey("pw", p); err != nil {
		t.Fatal(err)
	}
	bad := p
	bad.Memory = 1024
	if _, err := DeriveKey("pw", bad); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	p := DefaultKDFParams()
	key, _ := DeriveKey("pw", p)
	pt := []byte("secret data")
	aad := []byte("header")
	blob, err := Encrypt(key, pt, aad)
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decrypt(key, blob, aad)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, pt) {
		t.Fatalf("mismatch")
	}
}

func TestDecryptWrongKey(t *testing.T) {
	p := DefaultKDFParams()
	key, _ := DeriveKey("pw", p)
	blob, _ := Encrypt(key, []byte("x"), []byte("aad"))
	key2, _ := DeriveKey("other", p)
	if _, err := Decrypt(key2, blob, []byte("aad")); err == nil {
		t.Fatal("expected error")
	}
}

func TestDecryptTamperedAAD(t *testing.T) {
	p := DefaultKDFParams()
	key, _ := DeriveKey("pw", p)
	blob, _ := Encrypt(key, []byte("x"), []byte("aad"))
	if _, err := Decrypt(key, blob, []byte("tampered")); err == nil {
		t.Fatal("expected error")
	}
}

func TestNonceUniqueness(t *testing.T) {
	p := DefaultKDFParams()
	key, _ := DeriveKey("pw", p)
	b1, _ := Encrypt(key, []byte("x"), []byte("aad"))
	b2, _ := Encrypt(key, []byte("x"), []byte("aad"))
	if bytes.Equal(b1[:NonceSize], b2[:NonceSize]) {
		t.Fatal("nonce reused")
	}
}
