package storage

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RebuildStats struct {
	Buckets       int
	Objects       int
	SidecarReads  int
	SidecarCreate int
	Errors        int
}

func RebuildIndex(db *sql.DB, dataDir string) (*RebuildStats, error) {
	stats := &RebuildStats{}

	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM objects"); err != nil {
		return nil, fmt.Errorf("failed to clear objects table: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM multipart_uploads"); err != nil {
		return nil, fmt.Errorf("failed to clear multipart_uploads table: %w", err)
	}

	if _, err := tx.Exec("DELETE FROM multipart_parts"); err != nil {
		return nil, fmt.Errorf("failed to clear multipart_parts table: %w", err)
	}

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read data directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		bucketDir := filepath.Join(dataDir, name)
		bucketMetaPath := filepath.Join(bucketDir, ".bucket.json")

		if _, err := os.Stat(bucketMetaPath); os.IsNotExist(err) {
			continue
		}

		bucketInfo, err := ReadBucketMeta(bucketDir)
		if err != nil {
			slog.Warn("Failed to read bucket metadata, skipping", "bucket", name, "err", err)
			stats.Errors++
			continue
		}

		bucketInfo.Name = name
		if bucketInfo.CreatedAt.IsZero() {
			bucketInfo.CreatedAt = time.Now().UTC()
		}

		if _, err := tx.Exec(
			"INSERT OR REPLACE INTO buckets (name, created_at, region, public, bucket_group, pinned) VALUES (?, ?, ?, ?, ?, ?)",
			bucketInfo.Name, bucketInfo.CreatedAt.UTC(), bucketInfo.Region, bucketInfo.Public, bucketInfo.Group, bucketInfo.Pinned,
		); err != nil {
			slog.Warn("Failed to insert bucket", "bucket", name, "err", err)
			stats.Errors++
			continue
		}
		stats.Buckets++

		err = filepath.Walk(bucketDir, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() {
				return nil
			}

			relPath, err := filepath.Rel(bucketDir, path)
			if err != nil {
				return err
			}

			if strings.HasPrefix(filepath.Base(relPath), ".") {
				return nil
			}
			if strings.HasSuffix(relPath, ".meta.json") {
				return nil
			}

			key := relPath
			sidecarPath := path + ".meta.json"

			var objInfo *ObjectInfo

			if sidecarData, err := os.ReadFile(sidecarPath); err == nil {
				stats.SidecarReads++
				var meta sidecarMeta
				if err := json.Unmarshal(sidecarData, &meta); err == nil {
					lastModified, _ := time.Parse(time.RFC3339, meta.LastModified)
					objInfo = &ObjectInfo{
						Bucket:       name,
						Key:          key,
						ETag:         meta.ETag,
						Size:         meta.Size,
						ContentType:  meta.ContentType,
						LastModified: lastModified,
						UserMetadata: meta.UserMetadata,
					}
				}
			}

			if objInfo == nil {
				stats.SidecarCreate++

				file, err := os.Open(path)
				if err != nil {
					slog.Warn("Failed to open file", "path", path, "err", err)
					stats.Errors++
					return nil
				}
				defer file.Close()

				hash := md5.New()
				size, err := io.Copy(hash, file)
				if err != nil {
					slog.Warn("Failed to hash file", "path", path, "err", err)
					stats.Errors++
					return nil
				}

				etagBytes := hash.Sum(nil)
				etag := "\"" + hex.EncodeToString(etagBytes) + "\""

				contentType := detectContentType(path)

				objInfo = &ObjectInfo{
					Bucket:       name,
					Key:          key,
					ETag:         etag,
					Size:         size,
					ContentType:  contentType,
					LastModified: fi.ModTime().UTC(),
				}

				if err := WriteSidecar(path, objInfo); err != nil {
					slog.Warn("Failed to write sidecar", "path", sidecarPath, "err", err)
				}
			}

			if _, err := tx.Exec(`
				INSERT OR REPLACE INTO objects (bucket, key, etag, size, content_type, last_modified)
				VALUES (?, ?, ?, ?, ?, ?)
			`, objInfo.Bucket, objInfo.Key, objInfo.ETag, objInfo.Size, objInfo.ContentType, objInfo.LastModified.UTC()); err != nil {
				slog.Warn("Failed to insert object", "bucket", name, "key", key, "err", err)
				stats.Errors++
				return nil
			}

			stats.Objects++
			return nil
		})

		if err != nil {
			slog.Warn("Error walking bucket directory", "bucket", name, "err", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return stats, nil
}

func detectContentType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".avif":
		return "image/avif"
	case ".pdf":
		return "application/pdf"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}
