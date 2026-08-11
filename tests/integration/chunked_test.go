package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// TestChunkedPutRoundTrip proves the plain-PUT path decodes aws-chunked
// framing instead of storing it as object bytes. The SDK won't produce this
// shape on demand, so the framed body is hand-rolled onto a presigned PUT.
func TestChunkedPutRoundTrip(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "chunked-put.txt"
	payload := "hello auxio chunked"

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	pc := presignClient(ts)
	req, err := pc.PresignPutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	})
	if err != nil {
		t.Fatalf("PresignPutObject failed: %v", err)
	}

	framed := fmt.Sprintf("%x\r\n%s\r\n0\r\n\r\n", len(payload), payload)
	httpReq, err := http.NewRequest(req.Method, req.URL, strings.NewReader(framed))
	if err != nil {
		t.Fatalf("build PUT request failed: %v", err)
	}
	for k, vs := range req.SignedHeader {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}
	httpReq.Header.Set("Content-Encoding", "aws-chunked")
	httpReq.Header.Set("x-amz-content-sha256", "STREAMING-UNSIGNED-PAYLOAD-TRAILER")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("chunked PUT failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("chunked PUT: got status %d, want 200", resp.StatusCode)
	}

	got, err := ts.Client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	})
	if err != nil {
		t.Fatalf("GetObject failed: %v", err)
	}
	stored, _ := io.ReadAll(got.Body)
	got.Body.Close()
	if !bytes.Equal(stored, []byte(payload)) {
		t.Fatalf("stored body = %q, want %q (framing not decoded?)", stored, payload)
	}
}

// TestMultipartMetadataNotPersisted pins a known gap: x-amz-meta-* on a multipart
// upload completes fine but is not carried to the final object.
func TestMultipartMetadataNotPersisted(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "multipart-meta.bin"

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}

	create, err := ts.Client.CreateMultipartUpload(ctx, &awss3.CreateMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		Metadata: map[string]string{"foo": "bar"},
	})
	if err != nil {
		t.Fatalf("CreateMultipartUpload failed: %v", err)
	}

	part, err := ts.Client.UploadPart(ctx, &awss3.UploadPartInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(key),
		UploadId:   create.UploadId,
		PartNumber: aws.Int32(1),
		Body:       bytes.NewReader([]byte("part one bytes")),
	})
	if err != nil {
		t.Fatalf("UploadPart failed: %v", err)
	}

	_, err = ts.Client.CompleteMultipartUpload(ctx, &awss3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadId: create.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: []types.CompletedPart{{ETag: part.ETag, PartNumber: aws.Int32(1)}},
		},
	})
	if err != nil {
		t.Fatalf("CompleteMultipartUpload failed: %v", err)
	}

	head, err := ts.Client.HeadObject(ctx, &awss3.HeadObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	})
	if err != nil {
		t.Fatalf("HeadObject failed: %v", err)
	}
	if v, ok := head.Metadata["foo"]; ok {
		t.Fatalf("multipart metadata unexpectedly persisted: foo=%q", v)
	}
}
