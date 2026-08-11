package integration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/protob/auxio/internal/config"
	"github.com/protob/auxio/internal/s3"
	"github.com/protob/auxio/internal/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	testAccessKey = "testaccesskey"
	testSecretKey = "testsecretkey"
	testRegion    = "us-east-1"
)

type testServer struct {
	Endpoint string
	Store    *storage.FilesystemStore
	Client   *awss3.Client
	Cleanup  func()
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	dataDir := t.TempDir()

	cfg := &config.Config{
		AccessKey:          testAccessKey,
		SecretKey:          testSecretKey,
		DataDir:            dataDir,
		HttpPort:           "0",
		Region:             testRegion,
		UploadCleanupHours: 24,
	}

	store, err := storage.NewFilesystemStore(storage.FilesystemOptions{
		DataDir: cfg.DataDir,
	})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	auth := s3.NewAuthEngine(cfg.AccessKey, cfg.SecretKey, cfg.Region)
	router := s3.NewRouter(store, cfg.Region, auth, nil) // 4th arg: imagingMiddleware

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	endpoint := fmt.Sprintf("http://127.0.0.1:%d", port)

	server := &http.Server{
		Handler: router,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	client := newTestClient(endpoint)

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		store.Close()
	}

	return &testServer{
		Endpoint: endpoint,
		Store:    store,
		Client:   client,
		Cleanup:  cleanup,
	}
}

func newTestClient(endpoint string) *awss3.Client {
	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(testRegion),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			testAccessKey,
			testSecretKey,
			"",
		)),
	)
	if err != nil {
		panic(err) // test helper; nothing useful to do with this error
	}

	return awss3.NewFromConfig(awsCfg, func(o *awss3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})
}

func uniqueBucketName() string {
	return fmt.Sprintf("test-bucket-%d", time.Now().UnixNano())
}

func TestBucketLifecycle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()

	_, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	_, err = ts.Client.HeadBucket(ctx, &awss3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("HeadBucket failed: %v", err)
	}

	listResp, err := ts.Client.ListBuckets(ctx, &awss3.ListBucketsInput{})
	if err != nil {
		t.Fatalf("ListBuckets failed: %v", err)
	}

	found := false
	for _, b := range listResp.Buckets {
		if *b.Name == bucket {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("Bucket %s not found in ListBuckets", bucket)
	}

	_, err = ts.Client.DeleteBucket(ctx, &awss3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("DeleteBucket failed: %v", err)
	}

	_, err = ts.Client.HeadBucket(ctx, &awss3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	if err == nil {
		t.Fatal("HeadBucket should fail after bucket deletion")
	}
}

func TestBucketRecreatePreservesMetadata(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()

	_, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	if err := ts.Store.UpdateBucketPublic(bucket, true); err != nil {
		t.Fatalf("UpdateBucketPublic failed: %v", err)
	}
	if err := ts.Store.UpdateBucketGroup(bucket, "g"); err != nil {
		t.Fatalf("UpdateBucketGroup failed: %v", err)
	}

	// Apps that CreateBucket on every boot: the second create must succeed (200)
	// and leave the metadata untouched.
	_, err = ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("second CreateBucket failed: %v", err)
	}

	info, err := ts.Store.HeadBucket(bucket)
	if err != nil {
		t.Fatalf("HeadBucket failed: %v", err)
	}
	if !info.Public {
		t.Error("public flag wiped by S3 re-create")
	}
	if info.Group != "g" {
		t.Errorf("group = %q, want g", info.Group)
	}
}

func TestObjectLifecycle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "test-object.txt"
	content := []byte("Hello, World!")

	_, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	putResp, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String("text/plain"),
	})
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}
	etag := putResp.ETag

	getResp, err := ts.Client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	defer getResp.Body.Close()

	downloaded, err := io.ReadAll(getResp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	if !bytes.Equal(content, downloaded) {
		t.Fatalf("Content mismatch: expected %q, got %q", content, downloaded)
	}

	if getResp.ETag == nil || *getResp.ETag != *etag {
		t.Fatalf("ETag mismatch: expected %v, got %v", *etag, getResp.ETag)
	}

	headResp, err := ts.Client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("HeadObject failed: %v", err)
	}

	if *headResp.ContentLength != int64(len(content)) {
		t.Fatalf("ContentLength mismatch: expected %d, got %d", len(content), *headResp.ContentLength)
	}

	_, err = ts.Client.DeleteObject(ctx, &awss3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("DeleteObject failed: %v", err)
	}

	_, err = ts.Client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		t.Fatal("GetObject should fail after object deletion")
	}
}

func TestObjectMetadata(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "metadata-test.txt"
	content := []byte("test content")

	_, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	_, err = ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(content),
		ContentType: aws.String("application/json"),
		Metadata: map[string]string{
			"custom": "value123",
		},
	})
	if err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	headResp, err := ts.Client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("HeadObject failed: %v", err)
	}

	if headResp.ContentType == nil || *headResp.ContentType != "application/json" {
		t.Fatalf("ContentType mismatch: expected application/json, got %v", headResp.ContentType)
	}

	if headResp.Metadata["custom"] != "value123" {
		t.Fatalf("Custom metadata mismatch: expected value123, got %v", headResp.Metadata["custom"])
	}
}

func TestListObjectsWithPrefix(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()

	_, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	objects := []struct {
		key  string
		data []byte
	}{
		{"a.txt", []byte("a")},
		{"dir1/b.txt", []byte("b")},
		{"dir1/c.txt", []byte("c")},
		{"dir1/sub/d.txt", []byte("d")},
		{"dir2/e.txt", []byte("e")},
	}

	for _, obj := range objects {
		_, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(obj.key),
			Body:   bytes.NewReader(obj.data),
		})
		if err != nil {
			t.Fatalf("PutObject %s failed: %v", obj.key, err)
		}
	}

	listAll, err := ts.Client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2 failed: %v", err)
	}
	if int(*listAll.KeyCount) != 5 {
		t.Fatalf("Expected 5 objects, got %d", *listAll.KeyCount)
	}

	listPrefix, err := ts.Client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String("dir1/"),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2 with prefix failed: %v", err)
	}
	if int(*listPrefix.KeyCount) != 3 {
		t.Fatalf("Expected 3 objects with prefix dir1/, got %d", *listPrefix.KeyCount)
	}

	listDelimiter, err := ts.Client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String("dir1/"),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2 with delimiter failed: %v", err)
	}

	if len(listDelimiter.Contents) != 2 {
		t.Fatalf("Expected 2 contents with delimiter, got %d", len(listDelimiter.Contents))
	}
	if len(listDelimiter.CommonPrefixes) != 1 {
		t.Fatalf("Expected 1 common prefix, got %d", len(listDelimiter.CommonPrefixes))
	}
	if *listDelimiter.CommonPrefixes[0].Prefix != "dir1/sub/" {
		t.Fatalf("Expected common prefix 'dir1/sub/', got %s", *listDelimiter.CommonPrefixes[0].Prefix)
	}
}

func TestListPrefixWithLikeWildcards(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()

	_, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	// `_` is a SQL LIKE wildcard; prefixes must match literally. `%` is covered at
	// the storage level instead - signing a `%` key is its own SDK problem.
	for _, key := range []string{"a_b/x", "aXb/y"} {
		_, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   bytes.NewReader([]byte("v")),
		})
		if err != nil {
			t.Fatalf("PutObject %s failed: %v", key, err)
		}
	}

	underscore, err := ts.Client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String("a_b/"),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2 failed: %v", err)
	}
	if len(underscore.Contents) != 1 || *underscore.Contents[0].Key != "a_b/x" {
		t.Fatalf("prefix a_b/ matched %d keys, want exactly a_b/x", len(underscore.Contents))
	}
}

func TestTrailingSlashKeyRejected(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()

	_, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	_, err = ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("docs/"),
		Body:   bytes.NewReader([]byte{}),
	})
	if err == nil {
		t.Fatal("PutObject with folder-marker key docs/ should fail")
	}

	_, err = ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("docs/a.txt"),
		Body:   bytes.NewReader([]byte("in folder")),
	})
	if err != nil {
		t.Fatalf("PutObject docs/a.txt failed: %v", err)
	}

	list, err := ts.Client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2 failed: %v", err)
	}
	if len(list.CommonPrefixes) != 1 || *list.CommonPrefixes[0].Prefix != "docs/" {
		t.Fatalf("expected virtual folder docs/ in common prefixes, got %+v", list.CommonPrefixes)
	}
}

func TestBatchDelete(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()

	_, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	for i := 0; i < 5; i++ {
		_, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(fmt.Sprintf("file%d.txt", i)),
			Body:   bytes.NewReader([]byte(fmt.Sprintf("content%d", i))),
		})
		if err != nil {
			t.Fatalf("PutObject failed: %v", err)
		}
	}

	objects := []types.ObjectIdentifier{
		{Key: aws.String("file0.txt")},
		{Key: aws.String("file1.txt")},
		{Key: aws.String("file2.txt")},
	}

	_, err = ts.Client.DeleteObjects(ctx, &awss3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &types.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		t.Fatalf("DeleteObjects failed: %v", err)
	}

	listResp, err := ts.Client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("ListObjectsV2 failed: %v", err)
	}

	if int(*listResp.KeyCount) != 2 {
		t.Fatalf("Expected 2 objects after batch delete, got %d", *listResp.KeyCount)
	}
}

func TestMultipartUpload(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "large-file.bin"

	_, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	size := 15 * 1024 * 1024
	content := make([]byte, size)
	for i := range content {
		content[i] = byte(i % 256)
	}

	_, err = ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(content),
	})
	if err != nil {
		t.Fatalf("PutObject (multipart) failed: %v", err)
	}

	getResp, err := ts.Client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	defer getResp.Body.Close()

	downloaded, err := io.ReadAll(getResp.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	if !bytes.Equal(content, downloaded) {
		t.Fatal("Downloaded content does not match uploaded content")
	}

	if getResp.ETag == nil {
		t.Fatal("ETag is nil")
	}

	etag := *getResp.ETag
	if !bytes.Contains([]byte(etag), []byte("-")) {
		t.Logf("Warning: ETag %s doesn't look like multipart format (expected '-')", etag)
	}
}
