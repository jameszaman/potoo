package storage

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalDriver stores files on the local filesystem and serves them via the API.
type LocalDriver struct {
	dir     string // absolute or relative path to the upload directory
	baseURL string // e.g. "http://localhost:8080" — prepended to /uploads/{key}
}

func (d *LocalDriver) Save(r io.Reader, originalName string, size int64) (url string, key string, err error) {
	ext := strings.ToLower(filepath.Ext(originalName))
	key = fmt.Sprintf("%d_%s%s", time.Now().UnixMilli(), randHex(8), ext)
	dst := filepath.Join(d.dir, key)

	f, err := os.Create(dst)
	if err != nil {
		return "", "", fmt.Errorf("storage: create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		_ = os.Remove(dst)
		return "", "", fmt.Errorf("storage: write file: %w", err)
	}

	url = d.baseURL + "/uploads/" + key
	return url, key, nil
}

func (d *LocalDriver) Delete(key string) error {
	path := filepath.Join(d.dir, filepath.Base(key)) // Base() prevents path traversal
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("storage: delete %q: %w", key, err)
	}
	return nil
}

func randHex(n int) string {
	const chars = "abcdef0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}
