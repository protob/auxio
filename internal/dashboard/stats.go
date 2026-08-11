package dashboard

import (
	"context"
	"io/fs"
	"log/slog"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

type StatsOutput struct {
	Body struct {
		Buckets           int    `json:"buckets"`
		Objects           int64  `json:"objects"`
		TotalSize         int64  `json:"total_size"`
		TotalSizeHuman    string `json:"total_size_human"`
		UploadsInProgress int    `json:"uploads_in_progress"`
		CacheSize         int64  `json:"cache_size"`
		CacheSizeHuman    string `json:"cache_size_human"`
	}
}

func (h *Handler) GetStats(ctx context.Context, input *struct{}) (*StatsOutput, error) {
	buckets, objects, totalSize, err := h.store.ServerStats()
	if err != nil {
		return nil, err
	}

	var cacheSize int64
	cacheDir := filepath.Join(h.dataDir, ".imgcache")
	_ = filepath.WalkDir(cacheDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			cacheSize += info.Size()
		}
		return nil
	})

	uploads, err := h.store.CountMultipartUploads()
	if err != nil {
		slog.Warn("count multipart uploads", "err", err)
		uploads = 0
	}

	return &StatsOutput{
		Body: struct {
			Buckets           int    `json:"buckets"`
			Objects           int64  `json:"objects"`
			TotalSize         int64  `json:"total_size"`
			TotalSizeHuman    string `json:"total_size_human"`
			UploadsInProgress int    `json:"uploads_in_progress"`
			CacheSize         int64  `json:"cache_size"`
			CacheSizeHuman    string `json:"cache_size_human"`
		}{
			Buckets:           buckets,
			Objects:           objects,
			TotalSize:         totalSize,
			TotalSizeHuman:    humanize.Bytes(uint64(totalSize)),
			UploadsInProgress: uploads,
			CacheSize:         cacheSize,
			CacheSizeHuman:    humanize.Bytes(uint64(cacheSize)),
		},
	}, nil
}
