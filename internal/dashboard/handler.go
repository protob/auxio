package dashboard

import (
	"context"
	"net/http"

	"github.com/protob/auxio/internal/config"
	"github.com/protob/auxio/internal/storage"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store   *storage.FilesystemStore
	cfg     *config.Config
	dataDir string
	version string
	commit  string
}

func NewHandler(store *storage.FilesystemStore, cfg *config.Config, dataDir, version, commit string) *Handler {
	return &Handler{store: store, cfg: cfg, dataDir: dataDir, version: version, commit: commit}
}

func RegisterRoutes(router chi.Router, store *storage.FilesystemStore, cfg *config.Config, dataDir, version, commit string) huma.API {
	h := NewHandler(store, cfg, dataDir, version, commit)

	api := humachi.New(router, huma.DefaultConfig("Auxio Dashboard API", "0.1.0"))

	huma.Register(api, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/api/login",
		Summary:     "Dashboard login",
		Tags:        []string{"Auth"},
	}, h.Login)

	huma.Register(api, huma.Operation{
		OperationID: "get-health",
		Method:      http.MethodGet,
		Path:        "/api/health",
		Summary:     "Health check",
		Tags:        []string{"System"},
	}, h.GetHealth)

	huma.Register(api, huma.Operation{
		OperationID: "get-stats",
		Method:      http.MethodGet,
		Path:        "/api/stats",
		Summary:     "Get server statistics",
		Tags:        []string{"System"},
	}, h.GetStats)

	huma.Register(api, huma.Operation{
		OperationID: "list-buckets",
		Method:      http.MethodGet,
		Path:        "/api/buckets",
		Summary:     "List all buckets with stats",
		Tags:        []string{"Buckets"},
	}, h.ListBuckets)

	huma.Register(api, huma.Operation{
		OperationID: "create-bucket",
		Method:      http.MethodPost,
		Path:        "/api/buckets",
		Summary:     "Create a new bucket",
		Tags:        []string{"Buckets"},
	}, h.CreateBucket)

	huma.Register(api, huma.Operation{
		OperationID: "delete-bucket",
		Method:      http.MethodDelete,
		Path:        "/api/buckets/{bucket}",
		Summary:     "Delete a bucket",
		Tags:        []string{"Buckets"},
	}, h.DeleteBucket)

	huma.Register(api, huma.Operation{
		OperationID: "update-bucket",
		Method:      http.MethodPatch,
		Path:        "/api/buckets/{bucket}",
		Summary:     "Update bucket settings",
		Tags:        []string{"Buckets"},
	}, h.UpdateBucket)

	huma.Register(api, huma.Operation{
		OperationID: "list-objects",
		Method:      http.MethodGet,
		Path:        "/api/buckets/{bucket}/objects",
		Summary:     "List objects in a bucket",
		Tags:        []string{"Objects"},
	}, h.ListObjects)

	huma.Register(api, huma.Operation{
		OperationID: "delete-objects",
		Method:      http.MethodDelete,
		Path:        "/api/buckets/{bucket}/objects",
		Summary:     "Delete multiple objects",
		Tags:        []string{"Objects"},
	}, h.DeleteObjects)

	// Bucket export (POST /api/export/{bucket}) is a plain chi streaming
	// handler registered in main.go - huma can't stream a tar.gz body.

	return api
}

type HealthOutput struct {
	Body struct {
		Status  string `json:"status"`
		Version string `json:"version"`
		Commit  string `json:"commit"`
	}
}

func (h *Handler) GetHealth(ctx context.Context, input *struct{}) (*HealthOutput, error) {
	return &HealthOutput{
		Body: struct {
			Status  string `json:"status"`
			Version string `json:"version"`
			Commit  string `json:"commit"`
		}{
			Status:  "ok",
			Version: h.version,
			Commit:  h.commit,
		},
	}, nil
}
