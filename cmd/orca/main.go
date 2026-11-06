package main

import (
	"bufio"
	"crypto/rand"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"

	"github.com/SatriaBagusNurhadianto/orca/internal/vault"
)

var isTTY bool
var stdinReader *bufio.Reader

func main() {
	isTTY = term.IsTerminal(int(os.Stdin.Fd()))
	if !isTTY {
		stdinReader = bufio.NewReader(os.Stdin)
	}
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	switch cmd {
	case "init":
		cmdInit()
	case "add":
		cmdAdd()
	case "list":
		cmdList()
	case "get":
		cmdGet()
	case "gen":
		cmdGen()
	case "delete":
		cmdDelete()
	case "export":
		cmdExport()
	case "import":
		cmdImport()
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println("orca - personal password manager")
	fmt.Println("commands: init, add, list, get, gen, delete, export, import")
}

func vaultPath() string {
	if p := os.Getenv("ORCA_VAULT"); p != "" {
		return p
	}
	return "vault.orca"
}

func readField(r *bufio.Reader) (string, error) {
	var n uint32
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return "", err
	}
	if n > 1<<20 {
		return "", fmt.Errorf("field too large")
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func readPassword(prompt string) string {
	fmt.Print(prompt)
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		b, err := term.ReadPassword(fd)
		if err != nil {
			fmt.Println()
			fmt.Println("error reading password:", err)
			os.Exit(1)
		}
		fmt.Println()
		s := string(b)
		for i := range b {
			b[i] = 0
		}
		return strings.TrimSpace(s)
	}
	// Non-interactive (subprocess): length-prefixed framing.
	s, err := readField(stdinReader)
	if err != nil {
		fmt.Println("error reading password:", err)
		os.Exit(1)
	}
	return s
}

func readLineOrField() string {
	if isTTY {
		line, _ := stdinReader.ReadString('\n')
		return strings.TrimSpace(line)
	}
	s, err := readField(stdinReader)
	if err != nil {
		fmt.Println("error reading field:", err)
		os.Exit(1)
	}
	return s
}

func cmdInit() {
	path := vaultPath()
	if _, err := os.Stat(path); err == nil {
		fmt.Println("vault already exists:", path)
		os.Exit(1)
	}
	pw := readPassword("Master password: ")
	if err := vault.Create(path, pw); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("vault created:", path)
}

func cmdAdd() {
	path := vaultPath()
	pw := readPassword("Master password: ")
	v, kdf, err := vault.Load(path, pw)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	var title, username, password, url, notes string
	if isTTY {
		r := bufio.NewReader(os.Stdin)
		tl, _ := r.ReadString('\n')
		title = strings.TrimSpace(tl)
		ul, _ := r.ReadString('\n')
		username = strings.TrimSpace(ul)
		pl, _ := r.ReadString('\n')
		password = strings.TrimSpace(pl)
		urll, _ := r.ReadString('\n')
		url = strings.TrimSpace(urll)
		nl, _ := r.ReadString('\n')
		notes = strings.TrimSpace(nl)
	} else {
		title = readLineOrField()
		username = readLineOrField()
		password = readLineOrField()
		url = readLineOrField()
		notes = readLineOrField()
	}

	id, err := newID()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	e := vault.Entry{
		ID:       id,
		Title:    title,
		Username: username,
		Password: password,
		URL:      url,
		Notes:    notes,
	}
	v.Entries = append(v.Entries, e)
	if err := vault.Save(path, pw, v, kdf); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("added:", e.Title)
}

func cmdList() {
	path := vaultPath()
	pw := readPassword("Master password: ")
	v, _, err := vault.Load(path, pw)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	for _, e := range v.Entries {
		if e.Deleted {
			continue
		}
		fmt.Printf("%s\t%s\t%s\n", e.ID, e.Title, e.Username)
	}
}

func cmdGet() {
	if len(os.Args) < 3 {
		fmt.Println("usage: orca get <id>")
		os.Exit(2)
	}
	id := os.Args[2]
	path := vaultPath()
	pw := readPassword("Master password: ")
	v, _, err := vault.Load(path, pw)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	for _, e := range v.Entries {
		if e.ID == id && !e.Deleted {
			fmt.Printf("Title: %s\nUsername: %s\nPassword: %s\nURL: %s\nNotes: %s\n", e.Title, e.Username, e.Password, e.URL, e.Notes)
			return
		}
	}
	fmt.Println("not found")
	os.Exit(1)
}

func cmdDelete() {
	if len(os.Args) < 3 {
		fmt.Println("usage: orca delete <id>")
		os.Exit(2)
	}
	id := os.Args[2]
	path := vaultPath()
	pw := readPassword("Master password: ")
	v, kdf, err := vault.Load(path, pw)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	found := false
	for i := range v.Entries {
		if v.Entries[i].ID == id && !v.Entries[i].Deleted {
			v.Entries[i].Deleted = true
			found = true
			break
		}
	}
	if !found {
		fmt.Println("not found")
		os.Exit(1)
	}
	if err := vault.Save(path, pw, v, kdf); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("deleted:", id)
}

func cmdExport() {
	plaintext := false
	outPath := ""
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--plaintext" {
			plaintext = true
		} else if os.Args[i] == "-o" && i+1 < len(os.Args) {
			outPath = os.Args[i+1]
			i++
		}
	}
	if !plaintext && outPath == "" {
		fmt.Println("error: export requires -o <path> or --plaintext")
		os.Exit(2)
	}
	path := vaultPath()
	pw := readPassword("Master password: ")
	v, _, err := vault.Load(path, pw)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	var buf strings.Builder
	w := csv.NewWriter(&buf)
	w.Write([]string{"title", "username", "password", "url", "notes"})
	for _, e := range v.Entries {
		if e.Deleted {
			continue
		}
		w.Write([]string{csvEscape(e.Title), csvEscape(e.Username), csvEscape(e.Password), csvEscape(e.URL), csvEscape(e.Notes)})
	}
	w.Flush()
	if err := w.Error(); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	data := []byte(buf.String())
	if plaintext {
		fmt.Fprint(os.Stderr, "WARNING: plaintext export - credentials will be visible\n")
		fmt.Print(string(data))
		return
	}
	if err := validatePath(outPath); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	dir := filepath.Dir(outPath)
	if dir == "" {
		dir = "."
	}
	tmp, err := os.CreateTemp(dir, ".export-*")
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		fmt.Println("error:", err)
		os.Exit(1)
	}
	tmp.Close()
	if err := os.Rename(tmpName, outPath); err != nil {
		os.Remove(tmpName)
		fmt.Println("error:", err)
		os.Exit(1)
	}
	os.Chmod(outPath, 0600)
	fmt.Println("exported to", outPath)
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
	// normalize separators
	norm := strings.ReplaceAll(path, "/", string(os.PathSeparator))
	norm = strings.ReplaceAll(norm, "\\", string(os.PathSeparator))
	for _, part := range strings.Split(norm, string(os.PathSeparator)) {
		if part == ".." {
			return fmt.Errorf("invalid path: traversal component")
		}
	}
	// walk prefixes, reject symlink components
	cur := ""
	for _, part := range strings.Split(norm, string(os.PathSeparator)) {
		if part == "" {
			continue
		}
		if cur == "" {
			if filepath.IsAbs(path) {
				cur = string(os.PathSeparator)
			}
		}
		cur = filepath.Join(cur, part)
		fi, err := os.Lstat(cur)
		if err == nil && fi.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("vault path has symlink component: %s", cur)
		}
	}
	// final check
	fi, err := os.Lstat(path)
	if err == nil && fi.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("vault path is a symlink")
	}
	return nil
}

func csvEscape(s string) string {
	if len(s) > 0 && (s[0] == '=' || s[0] == '+' || s[0] == '-' || s[0] == '@') {
		return "	" + s
	}
	return s
}

func cmdImport() {
	if len(os.Args) < 3 {
		fmt.Println("usage: orca import <file>")
		os.Exit(2)
	}
	if err := validatePath(os.Args[2]); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fi, err := os.Lstat(os.Args[2])
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if fi.Mode()&os.ModeSymlink != 0 {
		fmt.Println("error: import path is a symlink")
		os.Exit(1)
	}
	if fi.Size() > 10*1024*1024 {
		fmt.Println("error: import file too large")
		os.Exit(1)
	}
	f, err := os.Open(os.Args[2])
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, 10*1024*1024+1))
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	if len(data) > 10*1024*1024 {
		fmt.Println("error: import file too large")
		os.Exit(1)
	}
	path := vaultPath()
	pw := readPassword("Master password: ")
	v, kdf, err := vault.Load(path, pw)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	for i, rec := range records {
		if i == 0 {
			continue
		}
		if len(rec) < 3 {
			continue
		}
		for _, f := range rec {
			if len(f) > 4096 {
				fmt.Println("error: field too long")
				os.Exit(1)
			}
		}
		id, err := newID()
		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
		e := vault.Entry{
			ID:       id,
			Title:    rec[0],
			Username: safeField(rec, 1),
			Password: safeField(rec, 2),
			URL:      safeField(rec, 3),
			Notes:    safeField(rec, 4),
		}
		v.Entries = append(v.Entries, e)
	}
	if err := vault.Save(path, pw, v, kdf); err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	fmt.Println("imported")
}

func csvField(s string) string {
	if strings.ContainsAny(s, ",\"\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

func safeField(fields []string, idx int) string {
	if idx < len(fields) {
		return strings.TrimSpace(fields[idx])
	}
	return ""
}

func cmdGen() {
	fmt.Println(genPassword(16, true, true))
}

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16])), nil
}

func genPassword(n int, symbols, upper bool) string {
	const lower = "abcdefghijklmnopqrstuvwxyz"
	const up = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const sym = "!@#$%^&*()-_=+"
	pool := lower
	if upper {
		pool += up
	}
	if symbols {
		pool += sym
	}
	out := make([]byte, n)
	for i := range out {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
		if err != nil {
			return ""
		}
		out[i] = pool[idx.Int64()]
	}
	return string(out)
}
