package vault

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/SatriaBagusNurhadianto/orca/internal/crypto"
)

const Magic = "ORCA"
const Version uint16 = 1

const maxVaultSize = 100 * 1024 * 1024

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

func validatePath(path string) error {
	if path == "" {
		return fmt.Errorf("empty path")
	}
	for i := 0; i < len(path); i++ {
		if path[i] == 0 {
			return fmt.Errorf("invalid path: NUL byte")
		}
	}
	fi, err := os.Lstat(path)
	if err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("vault path is a symlink")
	}
	return nil
}

func Create(path, password string) error {
	if err := validatePath(path); err != nil {
		return err
	}
	kdf := crypto.DefaultKDFParams()
	v := Vault{Entries: []Entry{}}
	return Save(path, password, v, kdf)
}

func Save(path, password string, v Vault, kdf crypto.KDFParams) error {
	if err := validatePath(path); err != nil {
		return err
	}
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
	defer cryptoZero(key)
	header := buildHeader(kdf)
	blob, err := crypto.Encrypt(key, pt, header)
	if err != nil {
		return err
	}
	out := append([]byte(nil), header...)
	out = append(out, blob...)

	dir := filepath.Dir(path)
	if dir == "" {
		dir = "."
	}
	tmp, err := os.CreateTemp(dir, ".vault-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(out); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return err
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		os.Remove(tmpName)
		return err
	}
	if d, err := os.Open(dir); err == nil {
		d.Sync()
		d.Close()
	}
	return os.Chmod(path, 0600)
}

func Load(path, password string) (Vault, crypto.KDFParams, error) {
	if err := validatePath(path); err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	fi, err := os.Lstat(path)
	if err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		return Vault{}, crypto.KDFParams{}, fmt.Errorf("vault path is a symlink")
	}
	if fi.Size() > maxVaultSize {
		return Vault{}, crypto.KDFParams{}, fmt.Errorf("vault too large: %d", fi.Size())
	}
	f, err := os.Open(path)
	if err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, maxVaultSize+1))
	if err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	if int64(len(data)) > maxVaultSize {
		return Vault{}, crypto.KDFParams{}, fmt.Errorf("vault too large")
	}
	kdf, cipher, err := parseHeader(data)
	if err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	key, err := crypto.DeriveKey(password, kdf)
	if err != nil {
		return Vault{}, crypto.KDFParams{}, err
	}
	defer cryptoZero(key)
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

func cryptoZero(b []byte) {
	for i := range b {
		b[i] = 0
	}
	runtime.KeepAlive(b)
}

func zero(b []byte) { cryptoZero(b) }
