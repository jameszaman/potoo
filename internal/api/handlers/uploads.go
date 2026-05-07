package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/potoo/potoo/internal/storage"
)

type uploadResponse struct {
	URL      string `json:"url"`
	Key      string `json:"key"`
	Filename string `json:"filename"`
}

// UploadImageHTTP handles POST /v1/uploads/image.
// Accepts multipart/form-data with a single "file" field.
// Validates type (jpg/png/gif/webp) and size (≤1 MB) then delegates to the storage driver.
func (h *Handlers) UploadImageHTTP(w http.ResponseWriter, r *http.Request) {
	// Limit the request body before parsing to prevent memory exhaustion.
	// We allow a small overhead beyond MaxBytes for multipart boundary overhead.
	r.Body = http.MaxBytesReader(w, r.Body, storage.MaxBytes+64*1024)

	if err := r.ParseMultipartForm(storage.MaxBytes); err != nil {
		http.Error(w, `{"error":"file too large — maximum size is 1 MB"}`, http.StatusRequestEntityTooLarge)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"missing 'file' field"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	if err := storage.ValidateExt(header.Filename); err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	if header.Size > storage.MaxBytes {
		http.Error(w, `{"error":"file too large — maximum size is 1 MB"}`, http.StatusRequestEntityTooLarge)
		return
	}

	url, key, err := h.storage.Save(file, header.Filename, header.Size)
	if err != nil {
		http.Error(w, `{"error":"upload failed"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(uploadResponse{
		URL:      url,
		Key:      key,
		Filename: header.Filename,
	})
}
