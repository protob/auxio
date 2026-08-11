package dashboard

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/protob/auxio/internal/storage"
	"github.com/go-chi/chi/v5"
)

func UploadHandlerChi(store *storage.FilesystemStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := chi.URLParam(r, "bucket")
		if bucket == "" {
			http.Error(w, `{"error":"bucket required"}`, http.StatusBadRequest)
			return
		}

		maxSize := int64(5 << 30)
		r.Body = http.MaxBytesReader(w, r.Body, maxSize)

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, `{"error":"failed to parse form"}`, http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("file")
		if err != nil {
			http.Error(w, `{"error":"no file provided"}`, http.StatusBadRequest)
			return
		}
		defer file.Close()

		key := r.FormValue("key")
		if key == "" {
			key = header.Filename
		}

		contentType := header.Header.Get("Content-Type")
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		_, err = store.PutObject(bucket, key, io.Reader(file), storage.ObjectMeta{
			ContentType: contentType,
		})
		if err != nil {
			if err == storage.ErrBucketNotFound {
				http.Error(w, `{"error":"bucket not found"}`, http.StatusNotFound)
				return
			}
			if err == storage.ErrInvalidKey {
				http.Error(w, `{"error":"invalid key"}`, http.StatusBadRequest)
				return
			}
			http.Error(w, `{"error":"failed to store file"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"key": key})
	}
}
