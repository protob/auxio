package storage

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFilesystemStore_Foundation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "auxio-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	store, err := NewFilesystemStore(FilesystemOptions{DataDir: tmpDir})
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	t.Run("CreateBucket", func(t *testing.T) {
		info, err := store.CreateBucket("test-bucket")
		if err != nil {
			t.Fatalf("failed to create bucket: %v", err)
		}

		if info.Name != "test-bucket" {
			t.Errorf("expected bucket name 'test-bucket', got %q", info.Name)
		}

		bucketMetaPath := filepath.Join(tmpDir, "test-bucket", ".bucket.json")
		if _, err := os.Stat(bucketMetaPath); os.IsNotExist(err) {
			t.Error("bucket metadata file was not created")
		}
	})

	t.Run("PutObject", func(t *testing.T) {
		content := []byte("hello world, this is test content")
		reader := bytes.NewReader(content)

		meta := ObjectMeta{
			ContentType: "text/plain",
			UserMetadata: map[string]string{
				"x-amz-meta-description": "test file",
			},
		}

		info, err := store.PutObject("test-bucket", "photos/cat.jpg", reader, meta)
		if err != nil {
			t.Fatalf("failed to put object: %v", err)
		}

		if info.Bucket != "test-bucket" {
			t.Errorf("expected bucket 'test-bucket', got %q", info.Bucket)
		}
		if info.Key != "photos/cat.jpg" {
			t.Errorf("expected key 'photos/cat.jpg', got %q", info.Key)
		}
		if info.Size != int64(len(content)) {
			t.Errorf("expected size %d, got %d", len(content), info.Size)
		}
		if !strings.HasPrefix(info.ETag, "\"") || !strings.HasSuffix(info.ETag, "\"") {
			t.Errorf("ETag should be quoted, got %q", info.ETag)
		}

		objectPath := filepath.Join(tmpDir, "test-bucket", "photos", "cat.jpg")
		if _, err := os.Stat(objectPath); os.IsNotExist(err) {
			t.Error("object file was not created at expected path")
		}

		sidecarPath := objectPath + ".meta.json"
		if _, err := os.Stat(sidecarPath); os.IsNotExist(err) {
			t.Error("sidecar metadata file was not created")
		}
	})

	t.Run("GetObject", func(t *testing.T) {
		reader, info, err := store.GetObject("test-bucket", "photos/cat.jpg")
		if err != nil {
			t.Fatalf("failed to get object: %v", err)
		}
		defer reader.Close()

		if info.ContentType != "text/plain" {
			t.Errorf("expected content type 'text/plain', got %q", info.ContentType)
		}

		content, err := io.ReadAll(reader)
		if err != nil {
			t.Fatalf("failed to read object content: %v", err)
		}

		expected := "hello world, this is test content"
		if string(content) != expected {
			t.Errorf("expected content %q, got %q", expected, string(content))
		}
	})

	t.Run("HeadObject", func(t *testing.T) {
		info, err := store.HeadObject("test-bucket", "photos/cat.jpg")
		if err != nil {
			t.Fatalf("failed to head object: %v", err)
		}

		if info.Key != "photos/cat.jpg" {
			t.Errorf("expected key 'photos/cat.jpg', got %q", info.Key)
		}
		if info.ContentType != "text/plain" {
			t.Errorf("expected content type 'text/plain', got %q", info.ContentType)
		}
	})

	t.Run("DeleteObject", func(t *testing.T) {
		err := store.DeleteObject("test-bucket", "photos/cat.jpg")
		if err != nil {
			t.Fatalf("failed to delete object: %v", err)
		}

		objectPath := filepath.Join(tmpDir, "test-bucket", "photos", "cat.jpg")
		if _, err := os.Stat(objectPath); !os.IsNotExist(err) {
			t.Error("object file should have been deleted")
		}

		sidecarPath := objectPath + ".meta.json"
		if _, err := os.Stat(sidecarPath); !os.IsNotExist(err) {
			t.Error("sidecar file should have been deleted")
		}

		_, err = store.HeadObject("test-bucket", "photos/cat.jpg")
		if err != ErrObjectNotFound {
			t.Errorf("expected ErrObjectNotFound, got %v", err)
		}
	})

	t.Run("DeleteBucket", func(t *testing.T) {
		err := store.DeleteBucket("test-bucket")
		if err != nil {
			t.Fatalf("failed to delete bucket: %v", err)
		}

		bucketDir := filepath.Join(tmpDir, "test-bucket")
		if _, err := os.Stat(bucketDir); !os.IsNotExist(err) {
			t.Error("bucket directory should have been deleted")
		}
	})

	t.Run("DeleteNonEmptyBucket", func(t *testing.T) {
		_, err := store.CreateBucket("nonempty-bucket")
		if err != nil {
			t.Fatalf("failed to create bucket: %v", err)
		}

		content := []byte("test")
		_, err = store.PutObject("nonempty-bucket", "file.txt", bytes.NewReader(content), ObjectMeta{})
		if err != nil {
			t.Fatalf("failed to put object: %v", err)
		}

		err = store.DeleteBucket("nonempty-bucket")
		if err != ErrBucketNotEmpty {
			t.Errorf("expected ErrBucketNotEmpty, got %v", err)
		}

		_ = store.DeleteObject("nonempty-bucket", "file.txt")
		_ = store.DeleteBucket("nonempty-bucket")
	})
}

func TestCreateBucketIdempotencyGuard(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewFilesystemStore(FilesystemOptions{DataDir: tmp})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	first, err := store.CreateBucket("keeper")
	if err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateBucketPublic("keeper", true); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateBucketGroup("keeper", "g"); err != nil {
		t.Fatal(err)
	}

	if _, err := store.CreateBucket("keeper"); err != ErrBucketExists {
		t.Fatalf("second CreateBucket = %v, want ErrBucketExists", err)
	}

	info, err := store.HeadBucket("keeper")
	if err != nil {
		t.Fatal(err)
	}
	if !info.Public {
		t.Error("public flag wiped by re-create")
	}
	if info.Group != "g" {
		t.Errorf("group = %q, want g", info.Group)
	}
	// sidecar stores RFC3339 (second precision), so compare truncated
	if !info.CreatedAt.Equal(first.CreatedAt.Truncate(time.Second)) {
		t.Errorf("CreatedAt changed: %v -> %v", first.CreatedAt, info.CreatedAt)
	}
}

func TestListObjectsPrefixMatchesLiterally(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewFilesystemStore(FilesystemOptions{DataDir: tmp})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if _, err := store.CreateBucket("wild"); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"a_b/x", "aXb/y", "p%q/z", "pXq/w"} {
		if _, err := store.PutObject("wild", key, strings.NewReader("v"), ObjectMeta{}); err != nil {
			t.Fatal(err)
		}
	}

	for prefix, want := range map[string]string{"a_b/": "a_b/x", "p%q/": "p%q/z"} {
		res, err := store.ListObjects("wild", ListOpts{Prefix: prefix})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Contents) != 1 || res.Contents[0].Key != want {
			t.Errorf("prefix %q matched %d keys, want exactly %q", prefix, len(res.Contents), want)
		}
	}
}

func TestUploadIDTraversalRejected(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewFilesystemStore(FilesystemOptions{DataDir: tmp})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	// sentinel outside .uploads that a "../.." RemoveAll would destroy
	if _, err := store.CreateBucket("canary-bucket"); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(tmp, "canary-bucket", ".bucket.json")

	if err := store.AbortMultipartUpload("b", "k", "../.."); err != ErrUploadNotFound {
		t.Fatalf("Abort with traversal id = %v, want ErrUploadNotFound", err)
	}
	if _, err := os.Stat(sentinel); err != nil {
		t.Fatalf("sentinel destroyed by traversal abort: %v", err)
	}

	if _, err := store.UploadPart("b", "k", "../../x", 1, strings.NewReader("data")); err != ErrUploadNotFound {
		t.Fatalf("UploadPart with traversal id = %v, want ErrUploadNotFound", err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(tmp), "x")); !os.IsNotExist(err) {
		t.Fatal("UploadPart created a directory outside the data dir")
	}

	// unknown-but-valid UUID must not leave an orphan dir under .uploads
	orphanID := "01234567-89ab-cdef-0123-456789abcdef"
	if _, err := store.UploadPart("b", "k", orphanID, 1, strings.NewReader("data")); err != ErrUploadNotFound {
		t.Fatalf("UploadPart with unknown UUID = %v, want ErrUploadNotFound", err)
	}
	if _, err := os.Stat(filepath.Join(tmp, ".uploads", orphanID)); !os.IsNotExist(err) {
		t.Fatal("UploadPart left an orphan dir for an unknown uploadId")
	}

	if _, err := store.CompleteMultipartUpload("b", "k", "../..", []PartInfo{{PartNumber: 1}}); err != ErrUploadNotFound {
		t.Fatalf("Complete with traversal id = %v, want ErrUploadNotFound", err)
	}
	if _, err := store.ListParts("b", "k", "../.."); err != ErrUploadNotFound {
		t.Fatalf("ListParts with traversal id = %v, want ErrUploadNotFound", err)
	}
}

func TestPutObjectRejectsTrailingSlashKey(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewFilesystemStore(FilesystemOptions{DataDir: tmp})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if _, err := store.CreateBucket("folders"); err != nil {
		t.Fatal(err)
	}

	if _, err := store.PutObject("folders", "docs/", strings.NewReader(""), ObjectMeta{}); err != ErrInvalidKey {
		t.Fatalf("PutObject(docs/) = %v, want ErrInvalidKey", err)
	}
	if _, err := store.PutObject("folders", "", strings.NewReader(""), ObjectMeta{}); err != ErrInvalidKey {
		t.Fatalf("PutObject(\"\") = %v, want ErrInvalidKey", err)
	}

	// uploading into the (virtual) folder works
	if _, err := store.PutObject("folders", "docs/a.txt", strings.NewReader("hi"), ObjectMeta{}); err != nil {
		t.Fatalf("PutObject(docs/a.txt) failed: %v", err)
	}
	rc, _, err := store.GetObject("folders", "docs/a.txt")
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	body, _ := io.ReadAll(rc)
	rc.Close()
	if string(body) != "hi" {
		t.Fatalf("round-trip body = %q, want hi", body)
	}
}

func TestCountMultipartUploads(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewFilesystemStore(FilesystemOptions{DataDir: tmp})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if _, err := store.CreateBucket("mpu"); err != nil {
		t.Fatal(err)
	}

	if n, _ := store.CountMultipartUploads(); n != 0 {
		t.Fatalf("initial count = %d, want 0", n)
	}

	uploadID, err := store.CreateMultipartUpload("mpu", "big.bin", ObjectMeta{})
	if err != nil {
		t.Fatal(err)
	}
	if n, _ := store.CountMultipartUploads(); n != 1 {
		t.Fatalf("count after create = %d, want 1", n)
	}

	if err := store.AbortMultipartUpload("mpu", "big.bin", uploadID); err != nil {
		t.Fatal(err)
	}
	if n, _ := store.CountMultipartUploads(); n != 0 {
		t.Fatalf("count after abort = %d, want 0", n)
	}
}

func TestBucketGroupDualWrite(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewFilesystemStore(FilesystemOptions{DataDir: tmp})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if _, err := store.CreateBucket("ecom-products"); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateBucketGroup("ecom-products", "ecom"); err != nil {
		t.Fatal(err)
	}

	// sidecar path: HeadBucket reads .bucket.json
	if info, _ := store.HeadBucket("ecom-products"); info.Group != "ecom" {
		t.Fatalf("sidecar group = %q, want ecom", info.Group)
	}
	// index path: ListBucketsWithStats reads index.db
	rows, _ := store.ListBucketsWithStats()
	if len(rows) != 1 || rows[0].Group != "ecom" {
		t.Fatalf("index group = %+v, want ecom", rows)
	}
	// the literal sidecar file carries it (MarshalIndent -> `"group": "ecom"`)
	data, _ := os.ReadFile(filepath.Join(tmp, "ecom-products", ".bucket.json"))
	if !strings.Contains(string(data), `"group": "ecom"`) {
		t.Fatalf(".bucket.json missing group: %s", data)
	}
}

func TestBucketGroupSurvivesIndexSkew(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewFilesystemStore(FilesystemOptions{DataDir: tmp})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateBucket("saas-avatars"); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateBucketGroup("saas-avatars", "saas"); err != nil {
		t.Fatal(err)
	}
	// simulate an index that lost the group while the sidecar still has it
	if err := store.index.UpdateBucketGroup("saas-avatars", ""); err != nil {
		t.Fatal(err)
	}
	store.Close()

	// reopening reconciles group from the sidecar (syncBucketPublicFlags)
	store2, err := NewFilesystemStore(FilesystemOptions{DataDir: tmp})
	if err != nil {
		t.Fatal(err)
	}
	defer store2.Close()
	rows, _ := store2.ListBucketsWithStats()
	if len(rows) != 1 || rows[0].Group != "saas" {
		t.Fatalf("group not reconciled from sidecar: %+v", rows)
	}
}

func TestBucketPinnedDualWrite(t *testing.T) {
	tmp := t.TempDir()
	store, err := NewFilesystemStore(FilesystemOptions{DataDir: tmp})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	if _, err := store.CreateBucket("ecom-products"); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateBucketPinned("ecom-products", true); err != nil {
		t.Fatal(err)
	}

	// sidecar path
	if info, _ := store.HeadBucket("ecom-products"); !info.Pinned {
		t.Fatalf("sidecar pinned = %v, want true", info.Pinned)
	}
	// index path
	rows, _ := store.ListBucketsWithStats()
	if len(rows) != 1 || !rows[0].Pinned {
		t.Fatalf("index pinned = %+v, want true", rows)
	}
	// literal sidecar file carries it
	data, _ := os.ReadFile(filepath.Join(tmp, "ecom-products", ".bucket.json"))
	if !strings.Contains(string(data), `"pinned": true`) {
		t.Fatalf(".bucket.json missing pinned: %s", data)
	}
}
