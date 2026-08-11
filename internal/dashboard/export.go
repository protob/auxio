package dashboard

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/protob/auxio/internal/storage"
	"github.com/go-chi/chi/v5"
)

func ExportBucketHandler(dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := chi.URLParam(r, "bucket")
		if storage.ValidateBucketName(bucket) != nil {
			http.Error(w, `{"error":"invalid bucket name"}`, http.StatusBadRequest)
			return
		}

		bucketDir := filepath.Join(dataDir, bucket)
		if _, err := os.Stat(bucketDir); os.IsNotExist(err) {
			http.Error(w, `{"error":"bucket not found"}`, http.StatusNotFound)
			return
		}

		filename := fmt.Sprintf("%s-%s.tar.gz", bucket, time.Now().Format("2006-01-02"))
		w.Header().Set("Content-Type", "application/gzip")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

		if err := exportBucketToWriter(dataDir, bucket, w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func exportBucketToWriter(dataDir, bucket string, w io.Writer) error {
	bucketDir := filepath.Join(dataDir, bucket)

	gzw := gzip.NewWriter(w)
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	return filepath.Walk(bucketDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(bucketDir, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		if strings.HasPrefix(relPath, ".") || strings.HasSuffix(relPath, ".meta.json") {
			if fi.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		header, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tw, file)
		return err
	})
}
