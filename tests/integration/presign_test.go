package integration

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// presignClient builds an S3 presign client sharing the test server's endpoint
// and credentials.
func presignClient(ts *testServer) *awss3.PresignClient {
	base := awss3.New(awss3.Options{
		Region:                     testRegion,
		Credentials:                ts.Client.Options().Credentials,
		BaseEndpoint:               aws.String(ts.Endpoint),
		UsePathStyle:               true,
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
	})
	return awss3.NewPresignClient(base)
}

func TestPresignGetRoundTrip(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "presigned-get.txt"
	content := []byte("presigned payload")

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	if _, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key), Body: bytes.NewReader(content),
	}); err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	pc := presignClient(ts)
	req, err := pc.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	})
	if err != nil {
		t.Fatalf("PresignGetObject failed: %v", err)
	}

	resp, err := http.Get(req.URL)
	if err != nil {
		t.Fatalf("GET presigned URL failed: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("valid presigned GET: got status %d, want 200", resp.StatusCode)
	}
	if !bytes.Equal(body, content) {
		t.Fatalf("valid presigned GET: body = %q, want %q", body, content)
	}

	tampered := flipSignature(req.URL)
	tResp, err := http.Get(tampered)
	if err != nil {
		t.Fatalf("GET tampered URL failed: %v", err)
	}
	tResp.Body.Close()
	if tResp.StatusCode != http.StatusForbidden {
		t.Fatalf("tampered presigned GET: got status %d, want 403", tResp.StatusCode)
	}
}

func TestPresignGetExpired(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "presigned-expired.txt"

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	if _, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key), Body: bytes.NewReader([]byte("x")),
	}); err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	pc := presignClient(ts)
	req, err := pc.PresignGetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	}, func(o *awss3.PresignOptions) { o.Expires = time.Second })
	if err != nil {
		t.Fatalf("PresignGetObject failed: %v", err)
	}

	time.Sleep(1100 * time.Millisecond) // let the 1s window elapse

	resp, err := http.Get(req.URL)
	if err != nil {
		t.Fatalf("GET expired URL failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expired presigned GET: got status %d, want 403", resp.StatusCode)
	}
}

func TestPresignPutRoundTrip(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "presigned-put.txt"
	content := []byte("uploaded via presigned PUT")

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

	httpReq, err := http.NewRequest(req.Method, req.URL, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("build PUT request failed: %v", err)
	}
	for k, vs := range req.SignedHeader {
		for _, v := range vs {
			httpReq.Header.Add(k, v)
		}
	}
	putResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		t.Fatalf("presigned PUT failed: %v", err)
	}
	putResp.Body.Close()
	if putResp.StatusCode < 200 || putResp.StatusCode >= 300 {
		t.Fatalf("presigned PUT: got status %d, want 2xx", putResp.StatusCode)
	}

	got, err := ts.Client.GetObject(ctx, &awss3.GetObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
	})
	if err != nil {
		t.Fatalf("GetObject after presigned PUT failed: %v", err)
	}
	stored, _ := io.ReadAll(got.Body)
	got.Body.Close()
	if !bytes.Equal(stored, content) {
		t.Fatalf("stored body = %q, want %q", stored, content)
	}
}

// flipSignature mutates the last hex character of X-Amz-Signature so the URL is
// structurally valid but the signature no longer verifies.
func flipSignature(rawURL string) string {
	const marker = "X-Amz-Signature="
	i := strings.Index(rawURL, marker)
	if i < 0 {
		return rawURL
	}
	start := i + len(marker)
	end := strings.IndexByte(rawURL[start:], '&')
	if end < 0 {
		end = len(rawURL) - start
	}
	sig := []byte(rawURL[start : start+end])
	last := sig[len(sig)-1]
	if last == '0' {
		sig[len(sig)-1] = '1'
	} else {
		sig[len(sig)-1] = '0'
	}
	return rawURL[:start] + string(sig) + rawURL[start+end:]
}
