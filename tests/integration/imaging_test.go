package integration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/protob/auxio/internal/imaging"
	"github.com/protob/auxio/internal/s3"
	"github.com/protob/auxio/internal/storage"
	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
)

// setupImagingServer wires Processor + Cache + middleware + mutation hook the
// same way main.go does, onto the s3.NewRouter the rest of these tests use.
func setupImagingServer(t *testing.T) *testServer {
	t.Helper()

	dataDir := t.TempDir()

	store, err := storage.NewFilesystemStore(storage.FilesystemOptions{DataDir: dataDir})
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}

	imaging.Startup()
	t.Cleanup(imaging.Shutdown)

	processor := imaging.NewProcessor(4096, 4096)
	cache := imaging.NewCache(filepath.Join(dataDir, ".imgcache"))
	mw := imaging.NewMiddleware(store, processor, cache)
	store.SetMutationHook(func(bucket, key string) {
		_ = cache.InvalidateForObject(bucket, key)
	})

	auth := s3.NewAuthEngine(testAccessKey, testSecretKey, testRegion)
	router := s3.NewRouter(store, testRegion, auth, mw.Handler)

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	endpoint := fmt.Sprintf("http://127.0.0.1:%d", port)

	server := &http.Server{Handler: router}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("server error: %v", err)
		}
	}()

	client := newTestClient(endpoint)

	return &testServer{
		Endpoint: endpoint,
		Store:    store,
		Client:   client,
		Cleanup: func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			server.Shutdown(ctx)
			store.Close()
		},
	}
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 4), G: uint8(y * 4), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// TestImagingTransformRoundTrip exercises store → imaging → cache → mutation
// hook through the real HTTP stack: transform on GET, cache HIT on repeat,
// MISS again after the object is re-uploaded.
func TestImagingTransformRoundTrip(t *testing.T) {
	ts := setupImagingServer(t)
	defer ts.Cleanup()

	ctx := context.Background()
	bucket := uniqueBucketName()
	key := "pic.png"

	if _, err := ts.Client.CreateBucket(ctx, &awss3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil {
		t.Fatalf("CreateBucket failed: %v", err)
	}
	if err := ts.Store.UpdateBucketPublic(bucket, true); err != nil {
		t.Fatalf("UpdateBucketPublic failed: %v", err)
	}
	srcPNG := testPNG(t)
	if _, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
		Body: bytes.NewReader(srcPNG), ContentType: aws.String("image/png"),
	}); err != nil {
		t.Fatalf("PutObject failed: %v", err)
	}

	url := ts.Endpoint + "/" + bucket + "/" + key + "?w=16&fmt=webp"

	fetch := func() (*http.Response, []byte) {
		resp, err := http.Get(url)
		if err != nil {
			t.Fatalf("GET %s failed: %v", url, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return resp, body
	}

	resp, body := fetch()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("transform GET: status %d, want 200 (body: %s)", resp.StatusCode, body)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "image/webp" {
		t.Fatalf("Content-Type = %q, want image/webp", ct)
	}
	if len(body) < 12 || string(body[:4]) != "RIFF" || string(body[8:12]) != "WEBP" {
		t.Fatalf("body is not webp (first bytes: %x)", body[:min(12, len(body))])
	}
	if c := resp.Header.Get("X-Auxio-Cache"); c != "MISS" {
		t.Fatalf("first GET X-Auxio-Cache = %q, want MISS", c)
	}

	resp, _ = fetch()
	if c := resp.Header.Get("X-Auxio-Cache"); c != "HIT" {
		t.Fatalf("second GET X-Auxio-Cache = %q, want HIT", c)
	}

	if _, err := ts.Client.PutObject(ctx, &awss3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key),
		Body: bytes.NewReader(srcPNG), ContentType: aws.String("image/png"),
	}); err != nil {
		t.Fatalf("re-PutObject failed: %v", err)
	}
	resp, _ = fetch()
	if c := resp.Header.Get("X-Auxio-Cache"); c != "MISS" {
		t.Fatalf("GET after re-upload X-Auxio-Cache = %q, want MISS", c)
	}
}
