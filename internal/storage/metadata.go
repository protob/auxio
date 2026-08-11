package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ErrPathTraversal is returned when a bucket/key would resolve outside the data dir.
var ErrPathTraversal = errors.New("path traversal detected")

type sidecarMeta struct {
	ETag         string            `json:"etag"`
	Size         int64             `json:"size"`
	ContentType  string            `json:"content_type"`
	LastModified string            `json:"last_modified"`
	UserMetadata map[string]string `json:"user_metadata,omitempty"`
}

type bucketMeta struct {
	CreatedAt string `json:"created_at"`
	Region    string `json:"region"`
	Public    bool   `json:"public"`
	Group     string `json:"group,omitempty"`
	Pinned    bool   `json:"pinned,omitempty"`
}

// ObjectPath joins dataDir/bucket/key and verifies the result stays inside the
// bucket directory. Every object read and write resolves its path here, so a
// traversal key that slips past the API layer - the dashboard upload does not
// pre-validate keys - cannot escape the bucket. Lexical check only: the data dir
// is assumed symlink-free, which also avoids a TOCTOU window.
func ObjectPath(dataDir, bucket, key string) (string, error) {
	bucketRoot := filepath.Join(filepath.Clean(dataDir), bucket)
	joined := filepath.Join(bucketRoot, key)
	if joined != bucketRoot && !strings.HasPrefix(joined, bucketRoot+string(filepath.Separator)) {
		return "", ErrPathTraversal
	}
	return joined, nil
}

func SidecarPath(objectPath string) string {
	return objectPath + ".meta.json"
}

func BucketMetaPath(bucketDir string) string {
	return filepath.Join(bucketDir, ".bucket.json")
}

func WriteSidecar(objectPath string, info *ObjectInfo) error {
	meta := sidecarMeta{
		ETag:         info.ETag,
		Size:         info.Size,
		ContentType:  info.ContentType,
		LastModified: info.LastModified.UTC().Format(time.RFC3339),
		UserMetadata: info.UserMetadata,
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sidecar metadata: %w", err)
	}

	sidecarPath := SidecarPath(objectPath)
	if err := os.WriteFile(sidecarPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write sidecar file: %w", err)
	}

	return nil
}

func ReadSidecar(objectPath string) (*ObjectInfo, error) {
	sidecarPath := SidecarPath(objectPath)
	data, err := os.ReadFile(sidecarPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrObjectNotFound
		}
		return nil, fmt.Errorf("failed to read sidecar file: %w", err)
	}

	var meta sidecarMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sidecar metadata: %w", err)
	}

	lastModified, err := time.Parse(time.RFC3339, meta.LastModified)
	if err != nil {
		return nil, fmt.Errorf("failed to parse last_modified: %w", err)
	}

	info := &ObjectInfo{
		ETag:         meta.ETag,
		Size:         meta.Size,
		ContentType:  meta.ContentType,
		LastModified: lastModified,
		UserMetadata: meta.UserMetadata,
	}

	return info, nil
}

func WriteBucketMeta(bucketDir string, info *BucketInfo) error {
	meta := bucketMeta{
		CreatedAt: info.CreatedAt.UTC().Format(time.RFC3339),
		Region:    info.Region,
		Public:    info.Public,
		Group:     info.Group,
		Pinned:    info.Pinned,
	}

	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal bucket metadata: %w", err)
	}

	metaPath := BucketMetaPath(bucketDir)
	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write bucket metadata file: %w", err)
	}

	return nil
}

func ReadBucketMeta(bucketDir string) (*BucketInfo, error) {
	metaPath := BucketMetaPath(bucketDir)
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrBucketNotFound
		}
		return nil, fmt.Errorf("failed to read bucket metadata file: %w", err)
	}

	var meta bucketMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bucket metadata: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339, meta.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	info := &BucketInfo{
		CreatedAt: createdAt,
		Region:    meta.Region,
		Public:    meta.Public,
		Group:     meta.Group,
		Pinned:    meta.Pinned,
	}

	return info, nil
}

func DeleteSidecar(objectPath string) error {
	sidecarPath := SidecarPath(objectPath)
	err := os.Remove(sidecarPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
