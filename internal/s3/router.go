package s3

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/protob/auxio/internal/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func RegisterRoutes(r chi.Router, store storage.ObjectStore, region string) {
	handler := NewHandler(store, region)

	r.Get("/", handler.handleListBuckets)

	r.Route("/{bucket}", func(r chi.Router) {
		r.Put("/", handler.handleCreateBucket)
		r.Head("/", handler.handleHeadBucket)
		r.Delete("/", handler.handleDeleteBucket)
		r.Get("/", handler.dispatchBucketGet)
		r.Post("/", handler.dispatchBucketPost)

		r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPut:
				handler.dispatchObjectPut(w, r)
			case http.MethodGet:
				handler.handleGetObject(w, r)
			case http.MethodHead:
				handler.handleHeadObject(w, r)
			case http.MethodDelete:
				handler.dispatchObjectDelete(w, r)
			case http.MethodPost:
				handler.dispatchObjectPost(w, r)
			default:
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		})
	})
}

// NewRouter is the test-facing mirror of the wiring in main.go: middleware in a
// Group so route params resolve before auth and imaging run. Keep the two in sync.
func NewRouter(store storage.ObjectStore, region string, auth *AuthEngine, imagingMiddleware func(http.Handler) http.Handler) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(LoggingMiddleware)
	r.Use(middleware.StripSlashes)

	r.Group(func(gr chi.Router) {
		if auth != nil {
			gr.Use(AuthMiddleware(auth, store))
		}
		if imagingMiddleware != nil {
			gr.Use(imagingMiddleware)
		}
		RegisterRoutes(gr, store, region)
	})

	return r
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		status := ww.Status()

		lvl := slog.LevelInfo
		if status >= 400 {
			lvl = slog.LevelWarn
		}
		slog.LogAttrs(r.Context(), lvl, "request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", status),
			slog.Duration("dur", duration))
	})
}
