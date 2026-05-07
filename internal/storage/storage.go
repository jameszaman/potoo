package storage

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

// Driver is the storage backend.
type Driver interface {
	// Save persists r under a generated key derived from originalName.
	// Returns the public URL to embed in emails.
	Save(r io.Reader, originalName string, size int64) (url string, key string, err error)

	// Delete removes a previously saved file by its key.
	Delete(key string) error
}

// AllowedMIME is the set of image types safe for email clients.
var AllowedMIME = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/gif":  true,
	"image/webp": true,
}

// MaxBytes is the upload size limit. Email images should stay well under 1 MB
// to avoid bloating message size and being blocked by mail servers.
const MaxBytes = 1 * 1024 * 1024 // 1 MB

// ValidateExt returns an error if the extension is not an allowed image type.
func ValidateExt(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	mimeType := mime.TypeByExtension(ext)
	if !AllowedMIME[mimeType] {
		return fmt.Errorf("unsupported file type %q — allowed: jpg, png, gif, webp", ext)
	}
	return nil
}

// New builds a Driver from environment variables.
//
//	STORAGE_DRIVER=local  → LocalDriver (default)
//	STORAGE_DRIVER=s3     → S3Driver (requires STORAGE_S3_BUCKET, STORAGE_S3_REGION)
//
// For local storage, STORAGE_LOCAL_PATH sets the directory (default: ./uploads).
// STORAGE_BASE_URL must be set so that local URLs are absolute (e.g. http://localhost:8080).
func New(baseURL string) (Driver, error) {
	driver := envOr("STORAGE_DRIVER", "local")
	switch driver {
	case "local", "":
		dir := envOr("STORAGE_LOCAL_PATH", "./uploads")
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("storage: create upload dir %q: %w", dir, err)
		}
		return &LocalDriver{dir: dir, baseURL: baseURL}, nil
	case "s3":
		return nil, fmt.Errorf("storage: S3 driver not yet implemented — set STORAGE_DRIVER=local")
	default:
		return nil, fmt.Errorf("storage: unknown driver %q", driver)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
