package crypto

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

const (
	SaltSize   = 16
	NonceSize  = 24
	KeySize    = 32
	MinMemory  = 64 * 1024 // 64 MiB in KiB
	MinTime    = 3
	MinThreads = 4
)

type KDFParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	Salt        []byte
}

func DefaultKDFParams() KDFParams {
	salt := make([]byte, SaltSize)
	if _, err := rand.Read(salt); err != nil {
		panic(err)
	}
	return KDFParams{
		Memory:      64 * 1024,
		Iterations:  3,
		Parallelism: 4,
		Salt:        salt,
	}
}

func ValidateKDFParams(p KDFParams) error {
	if p.Memory < MinMemory {
		return fmt.Errorf("argon2 memory too low: %d < %d", p.Memory, MinMemory)
	}
	if p.Iterations < MinTime {
		return fmt.Errorf("argon2 iterations too low: %d < %d", p.Iterations, MinTime)
	}
	if p.Parallelism < MinThreads {
		return fmt.Errorf("argon2 parallelism too low: %d < %d", p.Parallelism, MinThreads)
	}
	if len(p.Salt) != SaltSize {
		return fmt.Errorf("invalid salt size: %d", len(p.Salt))
	}
	return nil
}

func DeriveKey(password string, p KDFParams) ([]byte, error) {
	if err := ValidateKDFParams(p); err != nil {
		return nil, err
	}
	key := argon2.IDKey([]byte(password), p.Salt, p.Iterations, p.Memory, p.Parallelism, KeySize)
	return key, nil
}

// Encrypt seals plaintext with XChaCha20-Poly1305. aad is authenticated (vault header).
// Returns nonce||ciphertext||tag (nonce prepended).
func Encrypt(key, plaintext, aad []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, NonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	ct := aead.Seal(nonce, nonce, plaintext, aad)
	return ct, nil
}

// Decrypt opens a blob produced by Encrypt.
func Decrypt(key, blob, aad []byte) ([]byte, error) {
	aead, err := chacha20poly1305.NewX(key)
	if err != nil {
		return nil, err
	}
	if len(blob) < NonceSize+aead.Overhead() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := blob[:NonceSize]
	ct := blob[NonceSize:]
	pt, err := aead.Open(nil, nonce, ct, aad)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}
	return pt, nil
}

// EncodeKDFParams serializes params (memory, iterations, parallelism, salt) into 29 bytes.
func EncodeKDFParams(p KDFParams) []byte {
	buf := make([]byte, 4+4+1+SaltSize)
	binary.LittleEndian.PutUint32(buf[0:], p.Memory)
	binary.LittleEndian.PutUint32(buf[4:], p.Iterations)
	buf[8] = p.Parallelism
	copy(buf[9:], p.Salt)
	return buf
}

func DecodeKDFParams(b []byte) (KDFParams, error) {
	if len(b) != 4+4+1+SaltSize {
		return KDFParams{}, fmt.Errorf("invalid kdf params length: %d", len(b))
	}
	p := KDFParams{
		Memory:      binary.LittleEndian.Uint32(b[0:]),
		Iterations:  binary.LittleEndian.Uint32(b[4:]),
		Parallelism: b[8],
		Salt:        append([]byte(nil), b[9:]...),
	}
	if err := ValidateKDFParams(p); err != nil {
		return KDFParams{}, err
	}
	return p, nil
}
