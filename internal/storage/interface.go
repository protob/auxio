package storage

import (
	"io"
)

type MutationHook func(bucket, key string)

type ObjectStore interface {
	CreateBucket(name string) (*BucketInfo, error)
	DeleteBucket(name string) error
	HeadBucket(name string) (*BucketInfo, error)
	ListBuckets() ([]BucketInfo, error)
	UpdateBucketPublic(name string, public bool) error
	UpdateBucketGroup(name string, group string) error
	UpdateBucketPinned(name string, pinned bool) error

	PutObject(bucket string, key string, reader io.Reader, meta ObjectMeta) (*ObjectInfo, error)
	GetObject(bucket string, key string) (io.ReadCloser, *ObjectInfo, error)
	DeleteObject(bucket string, key string) error
	HeadObject(bucket string, key string) (*ObjectInfo, error)

	ListObjects(bucket string, opts ListOpts) (*ListResult, error)
	DeleteObjects(bucket string, keys []string) ([]DeleteResult, error)

	CreateMultipartUpload(bucket string, key string, meta ObjectMeta) (string, error)
	UploadPart(bucket string, key string, uploadID string, partNumber int, reader io.Reader) (*PartInfo, error)
	ListParts(bucket string, key string, uploadID string) ([]PartInfo, error)
	CompleteMultipartUpload(bucket string, key string, uploadID string, parts []PartInfo) (*ObjectInfo, error)
	AbortMultipartUpload(bucket string, key string, uploadID string) error
	ListMultipartUploads(bucket string) ([]UploadInfo, error)

	SetMutationHook(hook MutationHook)
}
