package storage

import (
	"errors"
	"time"
)

var (
	ErrBucketNotFound = errors.New("bucket not found")
	ErrBucketExists   = errors.New("bucket already exists")
	ErrBucketNotEmpty = errors.New("bucket not empty")
	ErrObjectNotFound = errors.New("object not found")
	ErrInvalidKey     = errors.New("invalid object key")
	ErrUploadNotFound = errors.New("multipart upload not found")
	ErrPartNotFound   = errors.New("part not found")
	ErrInvalidPart    = errors.New("invalid part")
)

type BucketInfo struct {
	Name      string
	CreatedAt time.Time
	Region    string
	Public    bool
	Group     string // dashboard-only grouping label; "" = ungrouped
	Pinned    bool   // dashboard-only favorite flag
}

type ObjectMeta struct {
	ContentType  string
	UserMetadata map[string]string
}

type ObjectInfo struct {
	Bucket       string
	Key          string
	ETag         string
	Size         int64
	ContentType  string
	LastModified time.Time
	UserMetadata map[string]string
}

type ListOpts struct {
	Prefix            string
	Delimiter         string
	MaxKeys           int
	ContinuationToken string
	StartAfter        string
}

type ListResult struct {
	Contents              []ObjectInfo
	CommonPrefixes        []string
	IsTruncated           bool
	NextContinuationToken string
	KeyCount              int
}

type DeleteResult struct {
	Key     string
	Success bool
	Error   string
}

type PartInfo struct {
	PartNumber int
	ETag       string
	Size       int64
}

type UploadInfo struct {
	UploadID    string
	Bucket      string
	Key         string
	Initiated   time.Time
	ContentType string
}
