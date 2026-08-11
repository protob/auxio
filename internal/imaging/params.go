package imaging

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
)

type TransformParams struct {
	Width   int
	Height  int
	Format  string
	Quality int
	Fit     string
}

func ParseFromQuery(query url.Values, maxWidth, maxHeight int) (*TransformParams, error) {
	if !query.Has("w") && !query.Has("h") && !query.Has("fmt") && !query.Has("q") && !query.Has("fit") {
		return nil, nil
	}

	params := &TransformParams{
		Quality: 80,
		Fit:     "contain",
	}

	if w := query.Get("w"); w != "" {
		width, err := strconv.Atoi(w)
		if err != nil {
			return nil, fmt.Errorf("invalid width: %s", w)
		}
		if width <= 0 || width > maxWidth {
			return nil, fmt.Errorf("width must be between 1 and %d", maxWidth)
		}
		params.Width = width
	}

	if h := query.Get("h"); h != "" {
		height, err := strconv.Atoi(h)
		if err != nil {
			return nil, fmt.Errorf("invalid height: %s", h)
		}
		if height <= 0 || height > maxHeight {
			return nil, fmt.Errorf("height must be between 1 and %d", maxHeight)
		}
		params.Height = height
	}

	if format := query.Get("fmt"); format != "" {
		format = strings.ToLower(format)
		if format != "jpeg" && format != "jpg" && format != "png" && format != "webp" && format != "avif" {
			return nil, fmt.Errorf("invalid format: %s (supported: jpeg, png, webp, avif)", format)
		}
		if format == "jpg" {
			format = "jpeg"
		}
		params.Format = format
	}

	if q := query.Get("q"); q != "" {
		quality, err := strconv.Atoi(q)
		if err != nil {
			return nil, fmt.Errorf("invalid quality: %s", q)
		}
		if quality < 1 || quality > 100 {
			return nil, fmt.Errorf("quality must be between 1 and 100")
		}
		params.Quality = quality
	}

	if fit := query.Get("fit"); fit != "" {
		fit = strings.ToLower(fit)
		if fit != "cover" && fit != "contain" {
			return nil, fmt.Errorf("invalid fit mode: %s (supported: cover, contain)", fit)
		}
		params.Fit = fit
	}

	return params, nil
}

// CacheKey returns {bucket}/{sha256(key)}/{sha256(params)} so Put creates
// exactly the tree InvalidateForObject and Cleanup walk. The two hash segments
// are hex; bucket arrives unvalidated from the URL, so it is the only segment
// that can shape the path.

func (p *TransformParams) CacheKey(bucket, key string) string {
	var parts []string

	if p.Width > 0 {
		parts = append(parts, fmt.Sprintf("w=%d", p.Width))
	}
	if p.Height > 0 {
		parts = append(parts, fmt.Sprintf("h=%d", p.Height))
	}
	if p.Format != "" {
		parts = append(parts, fmt.Sprintf("fmt=%s", p.Format))
	}
	parts = append(parts, fmt.Sprintf("q=%d", p.Quality))
	parts = append(parts, fmt.Sprintf("fit=%s", p.Fit))

	paramHash := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(parts, "&"))))
	return filepath.Join(bucket, KeyHash(key), paramHash)
}

func (p *TransformParams) IsNoop() bool {
	return p.Width == 0 && p.Height == 0 && p.Format == ""
}

func (p *TransformParams) String() string {
	var parts []string
	if p.Width > 0 {
		parts = append(parts, fmt.Sprintf("w=%d", p.Width))
	}
	if p.Height > 0 {
		parts = append(parts, fmt.Sprintf("h=%d", p.Height))
	}
	if p.Format != "" {
		parts = append(parts, fmt.Sprintf("fmt=%s", p.Format))
	}
	parts = append(parts, fmt.Sprintf("q=%d", p.Quality))
	parts = append(parts, fmt.Sprintf("fit=%s", p.Fit))
	return strings.Join(parts, ",")
}

func KeyHash(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
