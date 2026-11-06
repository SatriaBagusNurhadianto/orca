package totp

import (
	"testing"
	"time"
)

func TestCodeKnownVector(t *testing.T) {
	// RFC 6238 test vector: secret "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", time 59 -> code 287082
	code, err := Code("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", time.Unix(59, 0), 8, 30)
	if err != nil {
		t.Fatal(err)
	}
	if code != "94287082" {
		t.Fatalf("got %s, want 94287082", code)
	}
}

func TestGenerateSecret(t *testing.T) {
	s, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}
	if len(s) == 0 {
		t.Fatal("empty secret")
	}
	// should produce a code
	c, err := Code(s, time.Now(), 6, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(c) != 6 {
		t.Fatalf("bad code length: %s", c)
	}
}
