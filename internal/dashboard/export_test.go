package dashboard

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/protob/auxio/internal/storage"
	"github.com/go-chi/chi/v5"
)

func exportRouter(dataDir string) http.Handler {
	r := chi.NewRouter()
	r.Post("/api/export/{bucket}", ExportBucketHandler(dataDir))
	return r
}

func TestExportBucketRejectsInvalidNames(t *testing.T) {
	dataDir := t.TempDir()
	router := exportRouter(dataDir)

	for _, raw := range []string{"/api/export/..", "/api/export/..%2f..", "/api/export/has_underscore"} {
		req := httptest.NewRequest(http.MethodPost, raw, nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		// want a rejection, not a specific code: chi may 404 on path cleaning
		// before validation gets to return 400
		if rec.Code == http.StatusOK {
			t.Errorf("POST %s = 200, want rejection", raw)
		}
	}
}

func TestExportBucketHappyPath(t *testing.T) {
	dataDir := t.TempDir()
	bucketDir := filepath.Join(dataDir, "pics")
	if err := os.MkdirAll(bucketDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bucketDir, "a.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bucketDir, "a.txt.meta.json"), []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/export/pics", nil)
	rec := httptest.NewRecorder()
	exportRouter(dataDir).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	gzr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatalf("gunzip failed: %v", err)
	}
	tr := tar.NewReader(gzr)
	var names []string
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar read failed: %v", err)
		}
		names = append(names, hdr.Name)
	}

	found := false
	for _, n := range names {
		if n == "a.txt" {
			found = true
		}
		if n == "a.txt.meta.json" {
			t.Error("sidecar .meta.json leaked into export")
		}
	}
	if !found {
		t.Errorf("a.txt missing from export, got %v", names)
	}
}

func TestUploadResponseEncodesKey(t *testing.T) {
	dataDir := t.TempDir()
	store, err := storage.NewFilesystemStore(storage.FilesystemOptions{DataDir: dataDir})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if _, err := store.CreateBucket("ups"); err != nil {
		t.Fatal(err)
	}

	router := chi.NewRouter()
	router.Post("/api/buckets/{bucket}/upload", UploadHandlerChi(store))

	// a key containing a quote must still produce valid JSON
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	if err := mw.WriteField("key", `a"b.txt`); err != nil {
		t.Fatal(err)
	}
	fw, err := mw.CreateFormFile("file", "orig.txt")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write([]byte("content"))
	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/buckets/ups/upload", &body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	var out map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("response is not valid JSON: %v (%s)", err, rec.Body.String())
	}
	if out["key"] != `a"b.txt` {
		t.Errorf("key = %q, want a\"b.txt", out["key"])
	}
}
