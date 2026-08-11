package s3

import (
	"context"
	"net/http"
	"strings"

	"github.com/protob/auxio/internal/storage"
)

type contextKey string

const userContextKey contextKey = "user"

func AuthMiddleware(auth *AuthEngine, store storage.ObjectStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Route params are populated here under the Group wiring in NewRouter,
			// but not under a raw mux-level Use - parsePath works either way.
			bucket, key := parsePath(r.URL.Path)

			if requiresStrictAuth(r, bucket, key) {
				if err := auth.VerifyRequest(r); err != nil {
					writeAccessDenied(w)
					return
				}
				ctx = context.WithValue(ctx, userContextKey, auth.AccessKey)
				r = r.WithContext(ctx)
			} else if isReadOperation(r) && bucket != "" {
				info, err := store.HeadBucket(bucket)
				if err == nil && !info.Public {
					if err := auth.VerifyRequest(r); err != nil {
						writeAccessDenied(w)
						return
					}
					ctx = context.WithValue(ctx, userContextKey, auth.AccessKey)
					r = r.WithContext(ctx)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func parsePath(p string) (bucket, key string) {
	p = strings.Trim(p, "/")
	if p == "" {
		return "", ""
	}
	bucket, key, _ = strings.Cut(p, "/")
	return bucket, key
}

func requiresStrictAuth(r *http.Request, bucket, key string) bool {
	if r.Method == http.MethodPut || r.Method == http.MethodPost || r.Method == http.MethodDelete {
		return true
	}

	// GET / (list buckets) always requires auth.
	if r.Method == http.MethodGet && bucket == "" {
		return true
	}

	// Listing a bucket's inventory always requires auth, even on a public
	// bucket: AWS's canonical public-bucket policy grants s3:GetObject, not
	// s3:ListBucket. Anonymous GET-by-key (object fetch) stays open below.
	if isListRequest(r, bucket, key) {
		return true
	}

	return false
}

// isListRequest reports whether the request enumerates a bucket's contents
// (ListObjects V1/V2) rather than fetching a single object.
func isListRequest(r *http.Request, bucket, key string) bool {
	if r.Method != http.MethodGet {
		return false
	}
	q := r.URL.Query()
	if q.Has("list-type") || q.Has("prefix") || q.Has("marker") ||
		q.Has("continuation-token") || q.Has("max-keys") {
		return true
	}
	// Bare GET /{bucket} is a V1 list.
	return bucket != "" && key == ""
}

func isReadOperation(r *http.Request) bool {
	return r.Method == http.MethodGet || r.Method == http.MethodHead
}
