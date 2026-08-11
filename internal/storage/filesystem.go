package storage

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FilesystemStore struct {
	dataDir      string
	region       string
	index        *Index
	mutationHook MutationHook
}

type FilesystemOptions struct {
	DataDir string
	Region  string
}

func NewFilesystemStore(opts FilesystemOptions) (*FilesystemStore, error) {
	absDataDir, err := filepath.Abs(opts.DataDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve data directory: %w", err)
	}

	if err := os.MkdirAll(absDataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	tmpDir := filepath.Join(absDataDir, ".tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create tmp directory: %w", err)
	}

	index, err := NewIndex(IndexOptions{DataDir: absDataDir})
	if err != nil {
		return nil, fmt.Errorf("failed to create index: %w", err)
	}

	region := opts.Region
	if region == "" {
		region = "us-east-1"
	}

	store := &FilesystemStore{
		dataDir: absDataDir,
		region:  region,
		index:   index,
	}

	if err := store.syncBucketPublicFlags(); err != nil {
		return nil, fmt.Errorf("failed to sync bucket public flags: %w", err)
	}

	return store, nil
}

// syncBucketPublicFlags reconciles the index from the .bucket.json sidecars at
// startup: the sidecar is authoritative for public, group and pinned; the index
// is a cache that can drift.
func (s *FilesystemStore) syncBucketPublicFlags() error {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() || e.Name()[0] == '.' {
			continue
		}
		bucketDir := filepath.Join(s.dataDir, e.Name())
		info, err := ReadBucketMeta(bucketDir)
		if err != nil {
			continue
		}
		if err := s.index.UpdateBucketPublic(e.Name(), info.Public); err != nil {
			slog.Warn("bucket flag sync", "bucket", e.Name(), "err", err)
		}
		if err := s.index.UpdateBucketGroup(e.Name(), info.Group); err != nil {
			slog.Warn("bucket flag sync", "bucket", e.Name(), "err", err)
		}
		if err := s.index.UpdateBucketPinned(e.Name(), info.Pinned); err != nil {
			slog.Warn("bucket flag sync", "bucket", e.Name(), "err", err)
		}
	}
	return nil
}

func (s *FilesystemStore) Close() error {
	return s.index.Close()
}

func (s *FilesystemStore) SetMutationHook(hook MutationHook) {
	s.mutationHook = hook
}

func (s *FilesystemStore) CreateBucket(name string) (*BucketInfo, error) {
	if err := ValidateBucketName(name); err != nil {
		return nil, err
	}

	bucketDir := filepath.Join(s.dataDir, name)

	// Stat the meta file, not the dir: a dir without .bucket.json is not a
	// bucket, and re-creating over one self-heals it.
	if _, err := os.Stat(BucketMetaPath(bucketDir)); err == nil {
		return nil, ErrBucketExists
	}

	if err := os.MkdirAll(bucketDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create bucket directory: %w", err)
	}

	info := &BucketInfo{
		Name:      name,
		CreatedAt: time.Now().UTC(),
		Region:    s.region,
		Public:    false,
	}

	if err := WriteBucketMeta(bucketDir, info); err != nil {
		os.RemoveAll(bucketDir)
		return nil, err
	}

	if err := s.index.InsertBucket(info); err != nil {
		os.RemoveAll(bucketDir)
		return nil, err
	}

	return info, nil
}

func (s *FilesystemStore) UpdateBucketPublic(name string, public bool) error {
	bucketDir := filepath.Join(s.dataDir, name)
	info, err := ReadBucketMeta(bucketDir)
	if err != nil {
		return err
	}
	info.Name = name
	info.Public = public
	if err := WriteBucketMeta(bucketDir, info); err != nil {
		return err
	}
	return s.index.UpdateBucketPublic(name, public)
}

func (s *FilesystemStore) UpdateBucketGroup(name, group string) error {
	bucketDir := filepath.Join(s.dataDir, name)
	info, err := ReadBucketMeta(bucketDir)
	if err != nil {
		return err
	}
	info.Name = name
	info.Group = group
	if err := WriteBucketMeta(bucketDir, info); err != nil {
		return err
	}
	return s.index.UpdateBucketGroup(name, group)
}

func (s *FilesystemStore) UpdateBucketPinned(name string, pinned bool) error {
	bucketDir := filepath.Join(s.dataDir, name)
	info, err := ReadBucketMeta(bucketDir)
	if err != nil {
		return err
	}
	info.Name = name
	info.Pinned = pinned
	if err := WriteBucketMeta(bucketDir, info); err != nil {
		return err
	}
	return s.index.UpdateBucketPinned(name, pinned)
}

func (s *FilesystemStore) DeleteBucket(name string) error {
	hasObjects, err := s.index.BucketHasObjects(name)
	if err != nil {
		return err
	}
	if hasObjects {
		return ErrBucketNotEmpty
	}

	bucketDir := filepath.Join(s.dataDir, name)
	if err := os.RemoveAll(bucketDir); err != nil {
		return fmt.Errorf("failed to remove bucket directory: %w", err)
	}

	if err := s.index.DeleteBucket(name); err != nil {
		return err
	}

	return nil
}

func (s *FilesystemStore) HeadBucket(name string) (*BucketInfo, error) {
	bucketDir := filepath.Join(s.dataDir, name)

	info, err := ReadBucketMeta(bucketDir)
	if err != nil {
		return nil, err
	}

	info.Name = name
	return info, nil
}

func (s *FilesystemStore) ListBuckets() ([]BucketInfo, error) {
	return s.index.ListBuckets()
}

func (s *FilesystemStore) PutObject(bucket string, key string, reader io.Reader, meta ObjectMeta) (*ObjectInfo, error) {
	// Defense at the choke point: the dashboard upload path has no S3-layer
	// key validation, and filepath.Join would silently strip a trailing slash,
	// turning a "folder marker" into a plain file that blocks the prefix.
	if key == "" || strings.HasSuffix(key, "/") {
		return nil, ErrInvalidKey
	}

	if _, err := s.HeadBucket(bucket); err != nil {
		return nil, err
	}

	tmpFile, err := os.CreateTemp(filepath.Join(s.dataDir, ".tmp"), "upload-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	defer func() {
		tmpFile.Close()
		os.Remove(tmpPath)
	}()

	hash := md5.New()
	teeReader := io.TeeReader(reader, hash)

	size, err := io.Copy(tmpFile, teeReader)
	if err != nil {
		return nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync temp file: %w", err)
	}
	tmpFile.Close()

	etagBytes := hash.Sum(nil)
	etag := "\"" + hex.EncodeToString(etagBytes) + "\""

	objectPath, err := ObjectPath(s.dataDir, bucket, key)
	if err != nil {
		return nil, err
	}
	parentDir := filepath.Dir(objectPath)

	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	if err := os.Rename(tmpPath, objectPath); err != nil {
		return nil, fmt.Errorf("failed to move object to final location: %w", err)
	}

	now := time.Now().UTC()
	info := &ObjectInfo{
		Bucket:       bucket,
		Key:          key,
		ETag:         etag,
		Size:         size,
		ContentType:  meta.ContentType,
		LastModified: now,
		UserMetadata: meta.UserMetadata,
	}

	if err := WriteSidecar(objectPath, info); err != nil {
		os.Remove(objectPath)
		return nil, err
	}

	if err := s.index.InsertObject(info); err != nil {
		os.Remove(objectPath)
		DeleteSidecar(objectPath)
		return nil, err
	}

	if s.mutationHook != nil {
		s.mutationHook(bucket, key)
	}

	return info, nil
}

func (s *FilesystemStore) GetObject(bucket string, key string) (io.ReadCloser, *ObjectInfo, error) {
	objectPath, err := ObjectPath(s.dataDir, bucket, key)
	if err != nil {
		return nil, nil, err
	}

	file, err := os.Open(objectPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, ErrObjectNotFound
		}
		return nil, nil, fmt.Errorf("failed to open object: %w", err)
	}

	info, err := ReadSidecar(objectPath)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	info.Bucket = bucket
	info.Key = key

	return file, info, nil
}

func (s *FilesystemStore) DeleteObject(bucket string, key string) error {
	objectPath, err := ObjectPath(s.dataDir, bucket, key)
	if err != nil {
		return err
	}

	if err := os.Remove(objectPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove object: %w", err)
	}

	if err := DeleteSidecar(objectPath); err != nil {
		return err
	}

	if err := s.index.DeleteObject(bucket, key); err != nil {
		return err
	}

	s.cleanupEmptyParentDirs(objectPath, bucket)

	if s.mutationHook != nil {
		s.mutationHook(bucket, key)
	}

	return nil
}

func (s *FilesystemStore) HeadObject(bucket string, key string) (*ObjectInfo, error) {
	objectPath, err := ObjectPath(s.dataDir, bucket, key)
	if err != nil {
		return nil, err
	}

	info, err := ReadSidecar(objectPath)
	if err != nil {
		return nil, err
	}

	info.Bucket = bucket
	info.Key = key

	return info, nil
}

func (s *FilesystemStore) ListObjects(bucket string, opts ListOpts) (*ListResult, error) {
	if _, err := s.HeadBucket(bucket); err != nil {
		return nil, err
	}

	result, err := s.index.ListObjects(bucket, opts)
	if err != nil {
		return nil, err
	}

	if opts.Delimiter != "" {
		return s.applyDelimiter(result, opts), nil
	}

	return result, nil
}

func (s *FilesystemStore) applyDelimiter(result *ListResult, opts ListOpts) *ListResult {
	var contents []ObjectInfo
	commonPrefixes := make(map[string]struct{})

	for _, obj := range result.Contents {
		key := obj.Key

		if opts.Prefix != "" && !strings.HasPrefix(key, opts.Prefix) {
			continue
		}

		suffix := key
		if opts.Prefix != "" {
			suffix = strings.TrimPrefix(key, opts.Prefix)
		}

		idx := strings.Index(suffix, opts.Delimiter)
		if idx >= 0 {
			prefix := key[:len(opts.Prefix)+idx+len(opts.Delimiter)]
			commonPrefixes[prefix] = struct{}{}
		} else {
			contents = append(contents, obj)
		}
	}

	prefixes := make([]string, 0, len(commonPrefixes))
	for p := range commonPrefixes {
		prefixes = append(prefixes, p)
	}
	sort.Strings(prefixes)

	return &ListResult{
		Contents:              contents,
		CommonPrefixes:        prefixes,
		IsTruncated:           result.IsTruncated,
		NextContinuationToken: result.NextContinuationToken,
		KeyCount:              len(contents) + len(prefixes),
	}
}

func (s *FilesystemStore) DeleteObjects(bucket string, keys []string) ([]DeleteResult, error) {
	results := make([]DeleteResult, len(keys))

	for i, key := range keys {
		err := s.DeleteObject(bucket, key)
		if err != nil && err != ErrObjectNotFound {
			results[i] = DeleteResult{
				Key:     key,
				Success: false,
				Error:   err.Error(),
			}
		} else {
			results[i] = DeleteResult{
				Key:     key,
				Success: true,
			}
		}
	}

	return results, nil
}

// uploadMeta is the upload.json sidecar for an in-progress multipart upload,
// written by CreateMultipartUpload and read back by CleanupAbandonedUploads.
type uploadMeta struct {
	Bucket      string `json:"bucket"`
	Key         string `json:"key"`
	ContentType string `json:"content_type,omitempty"`
	Initiated   string `json:"initiated"`
}

func (s *FilesystemStore) CreateMultipartUpload(bucket string, key string, meta ObjectMeta) (string, error) {
	uploadID := uuid.New().String()

	uploadDir := filepath.Join(s.dataDir, ".uploads", uploadID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	now := time.Now().UTC()
	metaData := uploadMeta{
		Bucket:      bucket,
		Key:         key,
		ContentType: meta.ContentType,
		Initiated:   now.Format(time.RFC3339),
	}

	metaPath := filepath.Join(uploadDir, "upload.json")
	data, err := json.MarshalIndent(metaData, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal upload metadata: %w", err)
	}

	if err := os.WriteFile(metaPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write upload metadata: %w", err)
	}

	if err := s.index.InsertMultipartUpload(uploadID, bucket, key, meta.ContentType); err != nil {
		os.RemoveAll(uploadDir)
		return "", err
	}

	return uploadID, nil
}

// validUploadID rejects anything that is not one of our own UUIDs before the
// id is ever joined into a filesystem path.
func validUploadID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrUploadNotFound
	}
	return nil
}

func (s *FilesystemStore) UploadPart(bucket string, key string, uploadID string, partNumber int, reader io.Reader) (*PartInfo, error) {
	if err := validUploadID(uploadID); err != nil {
		return nil, err
	}

	if _, err := s.index.GetMultipartUpload(uploadID); err != nil {
		return nil, err
	}

	uploadDir := filepath.Join(s.dataDir, ".uploads", uploadID)
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	partPath := filepath.Join(uploadDir, fmt.Sprintf("%06d.part", partNumber))

	tmpFile, err := os.CreateTemp(uploadDir, "part-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	hash := md5.New()
	teeReader := io.TeeReader(reader, hash)

	size, err := io.Copy(tmpFile, teeReader)
	if err != nil {
		return nil, fmt.Errorf("failed to write part data: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync part file: %w", err)
	}
	tmpFile.Close()

	if err := os.Rename(tmpFile.Name(), partPath); err != nil {
		return nil, fmt.Errorf("failed to rename part file: %w", err)
	}

	etagBytes := hash.Sum(nil)
	etag := hex.EncodeToString(etagBytes)
	etagQuoted := "\"" + etag + "\""

	info := &PartInfo{
		PartNumber: partNumber,
		ETag:       etagQuoted,
		Size:       size,
	}

	if err := s.index.InsertPart(uploadID, partNumber, etag, size); err != nil {
		os.Remove(partPath)
		return nil, err
	}

	return info, nil
}

func (s *FilesystemStore) ListParts(bucket string, key string, uploadID string) ([]PartInfo, error) {
	if err := validUploadID(uploadID); err != nil {
		return nil, err
	}
	return s.index.ListParts(uploadID)
}

func (s *FilesystemStore) CompleteMultipartUpload(bucket string, key string, uploadID string, parts []PartInfo) (*ObjectInfo, error) {
	if err := validUploadID(uploadID); err != nil {
		return nil, err
	}

	uploadInfo, err := s.index.GetMultipartUpload(uploadID)
	if err != nil {
		return nil, err
	}

	uploadDir := filepath.Join(s.dataDir, ".uploads", uploadID)

	sort.Slice(parts, func(i, j int) bool {
		return parts[i].PartNumber < parts[j].PartNumber
	})

	// Client-supplied part ETags must match what UploadPart recorded, per S3
	// CompleteMultipartUpload semantics.
	for i, p := range parts {
		stored, err := s.index.GetPart(uploadID, p.PartNumber)
		if err != nil {
			return nil, ErrInvalidPart
		}
		if strings.Trim(p.ETag, `"`) != strings.Trim(stored.ETag, `"`) {
			return nil, ErrInvalidPart
		}
		parts[i].Size = stored.Size
	}

	hash := md5.New()
	var totalSize int64
	for _, p := range parts {
		hash.Write([]byte(strings.Trim(p.ETag, "\"")))
		totalSize += p.Size
	}

	compositeETag := hex.EncodeToString(hash.Sum(nil))
	compositeETagQuoted := fmt.Sprintf("\"%s-%d\"", compositeETag, len(parts))

	objectPath, err := ObjectPath(s.dataDir, bucket, key)
	if err != nil {
		return nil, err
	}
	parentDir := filepath.Dir(objectPath)

	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Join(s.dataDir, ".tmp"), "complete-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	for _, p := range parts {
		partPath := filepath.Join(uploadDir, fmt.Sprintf("%06d.part", p.PartNumber))
		partFile, err := os.Open(partPath)
		if err != nil {
			// listed in the index but missing on disk - surface as InvalidPart
			return nil, ErrInvalidPart
		}
		if _, err := io.Copy(tmpFile, partFile); err != nil {
			partFile.Close()
			return nil, fmt.Errorf("failed to concatenate part %d: %w", p.PartNumber, err)
		}
		partFile.Close()
	}

	if err := tmpFile.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync final file: %w", err)
	}
	tmpFile.Close()

	if err := os.Rename(tmpFile.Name(), objectPath); err != nil {
		return nil, fmt.Errorf("failed to move object to final location: %w", err)
	}

	now := time.Now().UTC()
	info := &ObjectInfo{
		Bucket:       bucket,
		Key:          key,
		ETag:         compositeETagQuoted,
		Size:         totalSize,
		ContentType:  uploadInfo.ContentType,
		LastModified: now,
	}

	if err := WriteSidecar(objectPath, info); err != nil {
		os.Remove(objectPath)
		return nil, err
	}

	if err := s.index.InsertObject(info); err != nil {
		os.Remove(objectPath)
		DeleteSidecar(objectPath)
		return nil, err
	}

	if err := s.index.DeleteMultipartUpload(uploadID); err != nil {
		slog.Error("Failed to delete multipart upload from index", "uploadId", uploadID, "err", err)
	}

	if err := s.index.DeleteParts(uploadID); err != nil {
		slog.Error("Failed to delete parts from index", "uploadId", uploadID, "err", err)
	}

	if err := os.RemoveAll(uploadDir); err != nil {
		slog.Error("Failed to remove upload directory", "path", uploadDir, "err", err)
	}

	return info, nil
}

func (s *FilesystemStore) AbortMultipartUpload(bucket string, key string, uploadID string) error {
	if err := validUploadID(uploadID); err != nil {
		return err
	}

	uploadDir := filepath.Join(s.dataDir, ".uploads", uploadID)

	if err := os.RemoveAll(uploadDir); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove upload directory: %w", err)
	}

	if err := s.index.DeleteMultipartUpload(uploadID); err != nil {
		return err
	}

	if err := s.index.DeleteParts(uploadID); err != nil {
		return err
	}

	return nil
}

func (s *FilesystemStore) ListMultipartUploads(bucket string) ([]UploadInfo, error) {
	return s.index.ListMultipartUploads(bucket)
}

func (s *FilesystemStore) ServerStats() (buckets int, objects int64, totalSize int64, err error) {
	return s.index.ServerStats()
}

func (s *FilesystemStore) CountMultipartUploads() (int, error) {
	return s.index.CountMultipartUploads()
}

func (s *FilesystemStore) ListBucketsWithStats() ([]BucketStatsRow, error) {
	return s.index.ListBucketsWithStats()
}

func (s *FilesystemStore) CleanupAbandonedUploads(maxAgeHours int) error {
	uploadsDir := filepath.Join(s.dataDir, ".uploads")

	entries, err := os.ReadDir(uploadsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read uploads directory: %w", err)
	}

	cutoff := time.Now().UTC().Add(-time.Duration(maxAgeHours) * time.Hour)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		uploadID := entry.Name()
		uploadDir := filepath.Join(uploadsDir, uploadID)

		metaPath := filepath.Join(uploadDir, "upload.json")
		data, err := os.ReadFile(metaPath)
		if err != nil {
			if os.IsNotExist(err) {
				slog.Warn("Upload directory missing metadata, removing", "uploadId", uploadID)
				os.RemoveAll(uploadDir)
			}
			continue
		}

		var metaData uploadMeta
		if err := json.Unmarshal(data, &metaData); err != nil {
			slog.Warn("Failed to parse upload metadata, removing", "uploadId", uploadID, "err", err)
			os.RemoveAll(uploadDir)
			s.index.DeleteMultipartUpload(uploadID)
			s.index.DeleteParts(uploadID)
			continue
		}

		initiated, err := time.Parse(time.RFC3339, metaData.Initiated)
		if err != nil {
			slog.Warn("Failed to parse initiated timestamp, removing", "uploadId", uploadID, "err", err)
			os.RemoveAll(uploadDir)
			s.index.DeleteMultipartUpload(uploadID)
			s.index.DeleteParts(uploadID)
			continue
		}

		if initiated.Before(cutoff) {
			slog.Info("Cleaning up abandoned upload", "uploadId", uploadID, "bucket", metaData.Bucket, "key", metaData.Key, "initiated", initiated)
			os.RemoveAll(uploadDir)
			s.index.DeleteMultipartUpload(uploadID)
			s.index.DeleteParts(uploadID)
		}
	}

	return nil
}

func (s *FilesystemStore) cleanupEmptyParentDirs(objectPath, bucket string) {
	bucketDir := filepath.Join(s.dataDir, bucket)
	dir := filepath.Dir(objectPath)

	for dir != bucketDir && dir != s.dataDir && strings.HasPrefix(dir, bucketDir) {
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) > 0 {
			break
		}
		os.Remove(dir)
		dir = filepath.Dir(dir)
	}
}

