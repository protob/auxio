package imaging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type CacheMeta struct {
	ContentType string `json:"content_type"`
	CreatedAt   string `json:"created_at"`
}

type Cache struct {
	dir string
}

func NewCache(dir string) *Cache {
	return &Cache{dir: dir}
}

func (c *Cache) Get(cacheKey string) (io.ReadCloser, string, bool) {
	filePath := filepath.Join(c.dir, cacheKey+".data")
	metaPath := filepath.Join(c.dir, cacheKey+".meta")

	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, "", false
	}

	var meta CacheMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, "", false
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", false
	}

	return file, meta.ContentType, true
}

func (c *Cache) Put(cacheKey string, data []byte, contentType string) error {
	filePath := filepath.Join(c.dir, cacheKey+".data")
	metaPath := filepath.Join(c.dir, cacheKey+".meta")

	// cacheKey is now a structured path ({bucket}/{keyHash}/{paramHash}), so
	// create the leaf's parent directory rather than just c.dir.
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	tmpFile := filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	if err := os.Rename(tmpFile, filePath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	meta := CacheMeta{
		ContentType: contentType,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	metaData, err := json.Marshal(meta)
	if err != nil {
		os.Remove(filePath)
		return fmt.Errorf("failed to marshal cache meta: %w", err)
	}

	tmpMeta := metaPath + ".tmp"
	if err := os.WriteFile(tmpMeta, metaData, 0644); err != nil {
		os.Remove(filePath)
		return fmt.Errorf("failed to write cache meta: %w", err)
	}

	if err := os.Rename(tmpMeta, metaPath); err != nil {
		os.Remove(tmpMeta)
		os.Remove(filePath)
		return fmt.Errorf("failed to rename cache meta: %w", err)
	}

	return nil
}

func (c *Cache) InvalidateForObject(bucket, key string) error {
	keyHash := KeyHash(key)
	cacheDir := filepath.Join(c.dir, bucket, keyHash)

	if err := os.RemoveAll(cacheDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	return nil
}

func (c *Cache) Cleanup(maxAge time.Duration) error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	cutoff := time.Now().UTC().Add(-maxAge)

	for _, entry := range entries {
		if entry.IsDir() {
			c.cleanupBucketDir(filepath.Join(c.dir, entry.Name()), cutoff)
		} else {
			c.cleanupCacheFile(entry.Name(), cutoff)
		}
	}

	return nil
}

func (c *Cache) cleanupBucketDir(bucketDir string, cutoff time.Time) {
	entries, err := os.ReadDir(bucketDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			keyDir := filepath.Join(bucketDir, entry.Name())
			c.cleanupKeyDir(keyDir, cutoff)
		}
	}
}

func (c *Cache) cleanupKeyDir(keyDir string, cutoff time.Time) {
	entries, err := os.ReadDir(keyDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			c.cleanupCacheFile(filepath.Join(keyDir, entry.Name()), cutoff)
		}
	}
}

func (c *Cache) cleanupCacheFile(path string, cutoff time.Time) {
	if !strings.HasSuffix(path, ".meta") {
		return
	}

	metaData, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var meta CacheMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return
	}

	createdAt, err := time.Parse(time.RFC3339, meta.CreatedAt)
	if err != nil {
		return
	}

	if createdAt.Before(cutoff) {
		cacheKey := strings.TrimSuffix(filepath.Base(path), ".meta")
		dataPath := filepath.Join(filepath.Dir(path), cacheKey+".data")
		os.Remove(path)
		os.Remove(dataPath)
	}
}
