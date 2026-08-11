package imaging

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/protob/auxio/internal/storage"
	"github.com/go-chi/chi/v5"
)

type Middleware struct {
	store     storage.ObjectStore
	processor *Processor
	cache     *Cache
}

func NewMiddleware(store storage.ObjectStore, processor *Processor, cache *Cache) *Middleware {
	return &Middleware{
		store:     store,
		processor: processor,
		cache:     cache,
	}
}

func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		params, err := ParseFromQuery(r.URL.Query(), m.processor.MaxWidth(), m.processor.MaxHeight())
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if params == nil || params.IsNoop() {
			next.ServeHTTP(w, r)
			return
		}

		bucket := chi.URLParam(r, "bucket")
		key := strings.TrimPrefix(r.URL.Path, "/"+bucket+"/")

		if bucket == "" || key == "" {
			next.ServeHTTP(w, r)
			return
		}

		cacheKey := params.CacheKey(bucket, key)

		if reader, contentType, found := m.cache.Get(cacheKey); found {
			defer reader.Close()
			w.Header().Set("Content-Type", contentType)
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			w.Header().Set("X-Auxio-Cache", "HIT")
			setDisposition(w, key, params.Format)
			io.Copy(w, reader)
			return
		}

		sourceReader, _, err := m.store.GetObject(bucket, key)
		if err != nil {
			if err == storage.ErrObjectNotFound || err == storage.ErrBucketNotFound {
				http.Error(w, "Not Found", http.StatusNotFound)
				return
			}
			slog.Error("Failed to get source object", "bucket", bucket, "key", key, "err", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		defer sourceReader.Close()

		transformed, contentType, err := m.processor.Transform(sourceReader, *params)
		if err != nil {
			slog.Error("Failed to transform image", "bucket", bucket, "key", key, "err", err)
			http.Error(w, "Failed to process image", http.StatusInternalServerError)
			return
		}

		if err := m.cache.Put(cacheKey, transformed, contentType); err != nil {
			slog.Warn("Failed to cache transformed image", "err", err)
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("X-Auxio-Cache", "MISS")
		w.Header().Set("Content-Length", strconv.Itoa(len(transformed)))
		setDisposition(w, key, params.Format)
		io.Copy(w, bytes.NewReader(transformed))
	})
}

// setDisposition sets Content-Disposition with a corrected filename when the
// output format differs from the source file's extension.
func setDisposition(w http.ResponseWriter, key string, outFormat string) {
	if outFormat == "" {
		return
	}
	ext := map[string]string{
		"jpeg": ".jpg",
		"png":  ".png",
		"webp": ".webp",
		"avif": ".avif",
	}
	newExt, ok := ext[outFormat]
	if !ok {
		return
	}
	base := strings.TrimSuffix(path.Base(key), path.Ext(path.Base(key)))
	filename := base + newExt
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, filename))
}
