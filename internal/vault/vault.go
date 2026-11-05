package vault

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"

	"github.com/SatriaBagusNurhadianto/orca/internal/crypto"
)

const Magic = "ORCA"
const Version uint16 = 1

type Entry struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	URL       string `json:"url"`
	Notes     string `json:"notes"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
	Deleted   bool   `json:"deleted,omitempty"`
}

type Vault struct {
	Entries []Entry `json:"entries"`
}

// headerSize = magic(4) + version(2) + kdf(29)
const headerSize = 4 + 2 + 25

func buildHeader(kdf crypto.KDFParams) []byte {
	buf := make([]byte, headerSize)
	copy(buf[0:], []byte(Magic))
	binary.LittleEndian.PutUint16(buf[4:], Version)
	copy(buf[6:], crypto.EncodeKDFParams(kdf))
	return buf
}

func parseHeader(data []byte) (crypto.KDFParams, []byte, error) {
	if len(data) < headerSize {
		return crypto.KDFParams{}, nil, fmt.Errorf("file too short")
	}
	if string(data[0:4]) != Magic {
		return crypto.KDFParams{}, nil, fmt.Errorf("invalid magic")
	}
	ver := binary.LittleEndian.Uint16(data[4:6])
	if ver != Version {
		return crypto.KDFParams{}, nil, fmt.Errorf("unsupported version: %d", ver)
	}
	kdf, err := crypto.DecodeKDFParams(data[6 : 6+25])
	if err != nil {
		return crypto.KDFParams{}, nil, err
	}
	return kdf, data[headerSize:], nil
}

// Create initializes a new empty vault file at path with the given master password.
func Create(path, password string) error {
	kdf := crypto.DefaultKDFParams()
	v := Vault{Entries: []Entry{}}
	return Save(path, password, v, kdf)
}

// Save writes vault to path, encrypted with password.
func Save(path, password string, v Vault, kdf crypto.KDFParams) error {
	if err := crypto.ValidateKDFParams(kdf); err != nil {
		return err
	}
	pt, err := json.Marshal(v)
	if err != nil {
		return err
	}
	key, err := crypto.DeriveKey(password, kdf)
	if err != nil {
		return err
	}
	defer zero(key)
	header := buildHeader(kdf)
	blob, err := crypto.Encrypt(key, pt, header)
	if err != nil {
		return err
	}
	out := append([]byte(nil), header...)
	out = append(out, blob...)
	return os.WriteFile(path, out, 0600)
}

// Load reads and decrypts a vault from path.
func Load(path, password string) (Vault, crypto.KDFParams, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	kdf, cipher, err := parseHeader(data)
	if err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	key, err := crypto.DeriveKey(password, kdf)
	if err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	defer zero(key)
	header := data[:headerSize]
	pt, err := crypto.Decrypt(key, cipher, header)
	if err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	var v Vault
	if err := json.Unmarshal(pt, &v); err != nil {
		return Vault{}, crypto.KDFParams{}, fmt.Errorf("invalid vault json: %w", err)
	}
	if v.Entries == nil {
		v.Entries = []Entry{}
	}
	return v, kdf, nil
}

func zero(b []byte) {
	for i := range b {
		b[i] = 0
	}
}
