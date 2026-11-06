package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"io"
	"strings"
	"testing"
)

func TestReadFieldRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	fields := []string{"secret", "title with\nnewline", "user", "pass\r\nwith\r", "http://x", "notes,with,commas"}
	for _, f := range fields {
		binary.Write(w, binary.BigEndian, uint32(len(f)))
		w.WriteString(f)
	}
	w.Flush()

	r := bufio.NewReader(bytes.NewReader(buf.Bytes()))
	for _, want := range fields {
		got, err := readField(r)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
	// trailing read should EOF
	if _, err := readField(r); err != io.EOF {
		t.Fatalf("expected EOF, got %v", err)
	}
}

func TestReadFieldRejectsOversized(t *testing.T) {
	var buf bytes.Buffer
	binary.Write(&buf, binary.BigEndian, uint32(2<<20))
	buf.WriteString(strings.Repeat("a", 2<<20))
	r := bufio.NewReader(bytes.NewReader(buf.Bytes()))
	if _, err := readField(r); err == nil {
		t.Fatal("expected size rejection")
	}
}
