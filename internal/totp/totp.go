package totp

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

func GenerateSecret() (string, error) {
	b := make([]byte, 20)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}

func Code(secret string, t time.Time, digits int, period int) (string, error) {
	if digits == 0 {
		digits = 6
	}
	if period == 0 {
		period = 30
	}
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(stringsToUpper(secret))
	if err != nil {
		return "", err
	}
	counter := uint64(t.Unix()) / uint64(period)
	msg := make([]byte, 8)
	binary.BigEndian.PutUint64(msg, counter)
	h := hmac.New(sha1.New, key)
	h.Write(msg)
	sum := h.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	code := (int(sum[offset])&0x7f)<<24 |
		(int(sum[offset+1])&0xff)<<16 |
		(int(sum[offset+2])&0xff)<<8 |
		(int(sum[offset+3]) & 0xff)
	mod := int(math.Pow(10, float64(digits)))
	code %= mod
	return fmt.Sprintf("%0*d", digits, code), nil
}

func stringsToUpper(s string) string {
	// base32 decode is case-insensitive in practice; uppercase to be safe
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'a' && c <= 'z' {
			c -= 32
		}
		out[i] = c
	}
	return string(out)
}
