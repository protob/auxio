package integration

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// TestPublicReadACLAndListAuth: a bucket created with x-amz-acl: public-read
// serves anonymous GET-by-key but still requires auth for listing.
func TestPublicReadACLAndListAuth(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "hero.jpg"
	content := []byte("image-bytes")

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{
		Bucket: aws.String(bucket),
		ACL:    types.BucketCannedACLPublicRead,
	}); err != nil {
		t.Fatalf("CreateBucket(public-read) failed: %v", err)
	}

	if _, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(content),
	}); err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	resp, err := http.Get(ts.Endpoint + "/" + bucket + "/" + key)
	if err != nil {
		t.Fatalf("anonymous GET failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("anonymous GET-by-key: got %d, want 200", resp.StatusCode)
	}
	if !bytes.Equal(body, content) {
		t.Fatalf("anonymous GET body mismatch: got %q", body)
	}

	listResp, err := http.Get(ts.Endpoint + "/" + bucket + "?list-type=2")
	if err != nil {
		t.Fatalf("anonymous list request failed: %v", err)
	}
	listResp.Body.Close()
	if listResp.StatusCode != http.StatusForbidden {
		t.Fatalf("anonymous list: got %d, want 403", listResp.StatusCode)
	}

	if _, err := ts.Client.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}); err != nil {
		t.Fatalf("signed list failed: %v", err)
	}
}

// TestPrivateBucketAnonGetDenied guards against unauthenticated data exfil: the
// auth middleware reads chi route params, which are empty if it is ever mounted
// mux-level instead of inside the Group - a private bucket would then serve
// anonymous GET-by-key.
func TestPrivateBucketAnonGetDenied(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "secret.txt"
	content := []byte("confidential")

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	if _, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key), Body: bytes.NewReader(content),
	}); err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	// Anonymous GET-by-key on a private bucket: denied.
	resp, err := http.Get(ts.Endpoint + "/" + bucket + "/" + key)
	if err != nil {
		t.Fatalf("anonymous GET failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("anonymous GET on private bucket: got %d, want 403", resp.StatusCode)
	}

	// Signed GET: allowed, returns the bytes.
	got, err := ts.Client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	})
	if err != nil {
		t.Fatalf("signed GET failed: %v", err)
	}
	body, _ := io.ReadAll(got.Body)
	got.Body.Close()
	if !bytes.Equal(body, content) {
		t.Fatalf("signed GET body mismatch: got %q", body)
	}
}

// TestExplicitMultipartFlow drives the MPU endpoints one call at a time, the
// way rclone and mc do.
func TestExplicitMultipartFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "explicit.bin"

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	create, err := ts.Client.CreateMultipartUpload(ctx, &awss3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		t.Fatalf("CreateMultipartUpload failed: %v", err)
	}
	uploadID := create.UploadId

	part1 := bytes.Repeat([]byte("a"), 1024)
	part2 := bytes.Repeat([]byte("b"), 512)

	up1, err := ts.Client.UploadPart(ctx, &awss3.UploadPartInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: uploadID,
		PartNumber: aws.Int32(1), Body: bytes.NewReader(part1),
	})
	if err != nil {
		t.Fatalf("UploadPart 1 failed: %v", err)
	}
	up2, err := ts.Client.UploadPart(ctx, &awss3.UploadPartInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: uploadID,
		PartNumber: aws.Int32(2), Body: bytes.NewReader(part2),
	})
	if err != nil {
		t.Fatalf("UploadPart 2 failed: %v", err)
	}

	if _, err := ts.Client.CompleteMultipartUpload(ctx, &awss3.CompleteMultipartUploadInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: uploadID,
		MultipartUpload: &types.CompletedMultipartUpload{Parts: []types.CompletedPart{
			{PartNumber: aws.Int32(1), ETag: up1.ETag},
			{PartNumber: aws.Int32(2), ETag: up2.ETag},
		}},
	}); err != nil {
		t.Fatalf("CompleteMultipartUpload failed: %v", err)
	}

	getResp, err := ts.Client.GetObject(ctx, &awss3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		t.Fatalf("GetObject after complete failed: %v", err)
	}
	got, _ := io.ReadAll(getResp.Body)
	getResp.Body.Close()
	if want := append(append([]byte{}, part1...), part2...); !bytes.Equal(got, want) {
		t.Fatalf("assembled object size %d, want %d", len(got), len(want))
	}
}

// TestCompleteMultipartInvalidPart proves Complete validates the client's
// part list against what was actually uploaded.
func TestCompleteMultipartInvalidPart(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "invalid-part.bin"

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	create, err := ts.Client.CreateMultipartUpload(ctx, &awss3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	})
	if err != nil {
		t.Fatalf("CreateMultipartUpload failed: %v", err)
	}

	up1, err := ts.Client.UploadPart(ctx, &awss3.UploadPartInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: create.UploadId,
		PartNumber: aws.Int32(1), Body: bytes.NewReader(bytes.Repeat([]byte("z"), 128)),
	})
	if err != nil {
		t.Fatalf("UploadPart failed: %v", err)
	}

	// wrong ETag for an uploaded part
	if _, err := ts.Client.CompleteMultipartUpload(ctx, &awss3.CompleteMultipartUploadInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: create.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{Parts: []types.CompletedPart{
			{PartNumber: aws.Int32(1), ETag: aws.String(`"00000000000000000000000000000000"`)},
		}},
	}); err == nil || !strings.Contains(err.Error(), "InvalidPart") {
		t.Fatalf("Complete with wrong ETag = %v, want InvalidPart", err)
	}

	// part number never uploaded
	if _, err := ts.Client.CompleteMultipartUpload(ctx, &awss3.CompleteMultipartUploadInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: create.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{Parts: []types.CompletedPart{
			{PartNumber: aws.Int32(7), ETag: up1.ETag},
		}},
	}); err == nil || !strings.Contains(err.Error(), "InvalidPart") {
		t.Fatalf("Complete with unknown part = %v, want InvalidPart", err)
	}

	// happy path still completes
	if _, err := ts.Client.CompleteMultipartUpload(ctx, &awss3.CompleteMultipartUploadInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: create.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{Parts: []types.CompletedPart{
			{PartNumber: aws.Int32(1), ETag: up1.ETag},
		}},
	}); err != nil {
		t.Fatalf("Complete with correct ETag failed: %v", err)
	}

	head, err := ts.Client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	})
	if err != nil {
		t.Fatalf("HeadObject failed: %v", err)
	}
	if *head.ContentLength != 128 {
		t.Fatalf("ContentLength = %d, want 128", *head.ContentLength)
	}
}

// TestAbortMultipartUpload proves an aborted upload's parts are gone.
func TestAbortMultipartUpload(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "aborted.bin"

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	create, err := ts.Client.CreateMultipartUpload(ctx, &awss3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	})
	if err != nil {
		t.Fatalf("CreateMultipartUpload failed: %v", err)
	}

	if _, err := ts.Client.UploadPart(ctx, &awss3.UploadPartInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: create.UploadId,
		PartNumber: aws.Int32(1), Body: bytes.NewReader(bytes.Repeat([]byte("x"), 256)),
	}); err != nil {
		t.Fatalf("UploadPart failed: %v", err)
	}

	if _, err := ts.Client.AbortMultipartUpload(ctx, &awss3.AbortMultipartUploadInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: create.UploadId,
	}); err != nil {
		t.Fatalf("AbortMultipartUpload failed: %v", err)
	}

	parts, err := ts.Client.ListParts(ctx, &awss3.ListPartsInput{
		Bucket: aws.String(bucket), Key: aws.String(key), UploadId: create.UploadId,
	})
	if err != nil {
		t.Fatalf("ListParts after abort failed: %v", err)
	}
	if len(parts.Parts) != 0 {
		t.Fatalf("ListParts after abort: got %d parts, want 0", len(parts.Parts))
	}

	if _, err := ts.Client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	}); err == nil {
		t.Fatal("HeadObject should fail for an aborted upload's key")
	}
}
