package dashboard

import (
	"context"
	"errors"
	"time"

	"github.com/protob/auxio/internal/storage"
	"github.com/danielgtaylor/huma/v2"
)

type BucketSummary struct {
	Name        string `json:"name"`
	CreatedAt   string `json:"created_at"`
	Region      string `json:"region"`
	Public      bool   `json:"public"`
	Group       string `json:"group"`
	Pinned      bool   `json:"pinned"`
	ObjectCount int64  `json:"object_count"`
	TotalSize   int64  `json:"total_size"`
}

type ListBucketsOutput struct {
	Body []BucketSummary
}

func (h *Handler) ListBuckets(ctx context.Context, input *struct{}) (*ListBucketsOutput, error) {
	stats, err := h.store.ListBucketsWithStats()
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to list buckets")
	}

	result := make([]BucketSummary, len(stats))
	for i, s := range stats {
		result[i] = BucketSummary{
			Name:        s.Name,
			CreatedAt:   s.CreatedAt.Format(time.RFC3339),
			Region:      s.Region,
			Public:      s.Public,
			Group:       s.Group,
			Pinned:      s.Pinned,
			ObjectCount: s.ObjectCount,
			TotalSize:   s.TotalSize,
		}
	}

	return &ListBucketsOutput{Body: result}, nil
}

type CreateBucketInput struct {
	Body struct {
		Name   string `json:"name" validate:"required,min=3,max=63"`
		Public bool   `json:"public"`
		Group  string `json:"group"`
	}
}

type CreateBucketOutput struct {
	Body struct {
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
		Region    string `json:"region"`
		Public    bool   `json:"public"`
		Group     string `json:"group"`
	}
}

func (h *Handler) CreateBucket(ctx context.Context, input *CreateBucketInput) (*CreateBucketOutput, error) {
	info, err := h.store.CreateBucket(input.Body.Name)
	if err != nil {
		if errors.Is(err, storage.ErrBucketExists) {
			return nil, huma.Error409Conflict("bucket already exists")
		}
		if errors.Is(err, storage.ErrBucketNotFound) {
			return nil, huma.Error404NotFound("bucket not found")
		}
		return nil, huma.Error400BadRequest(err.Error())
	}

	if input.Body.Public {
		if err := h.store.UpdateBucketPublic(input.Body.Name, true); err != nil {
			return nil, huma.Error500InternalServerError("failed to update bucket metadata")
		}
		info.Public = true
	}

	group := storage.NormalizeGroupName(input.Body.Group)
	if err := storage.ValidateGroupName(group); err != nil {
		return nil, huma.Error400BadRequest("invalid group name")
	}
	if group != "" {
		if err := h.store.UpdateBucketGroup(input.Body.Name, group); err != nil {
			return nil, huma.Error500InternalServerError("failed to set bucket group")
		}
		info.Group = group
	}

	return &CreateBucketOutput{
		Body: struct {
			Name      string `json:"name"`
			CreatedAt string `json:"created_at"`
			Region    string `json:"region"`
			Public    bool   `json:"public"`
			Group     string `json:"group"`
		}{
			Name:      info.Name,
			CreatedAt: info.CreatedAt.Format(time.RFC3339),
			Region:    info.Region,
			Public:    info.Public,
			Group:     info.Group,
		},
	}, nil
}

type DeleteBucketInput struct {
	Bucket string `path:"bucket"`
}

func (h *Handler) DeleteBucket(ctx context.Context, input *DeleteBucketInput) (*struct{}, error) {
	err := h.store.DeleteBucket(input.Bucket)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotFound) {
			return nil, huma.Error404NotFound("bucket not found")
		}
		if errors.Is(err, storage.ErrBucketNotEmpty) {
			return nil, huma.Error409Conflict("bucket not empty")
		}
		return nil, huma.Error500InternalServerError("failed to delete bucket")
	}
	return nil, nil
}

type UpdateBucketInput struct {
	Bucket string `path:"bucket"`
	Body   struct {
		Public *bool   `json:"public,omitempty"`
		Group  *string `json:"group,omitempty"`
		Pinned *bool   `json:"pinned,omitempty"`
	}
}

type UpdateBucketOutput struct {
	Body struct {
		Name      string `json:"name"`
		CreatedAt string `json:"created_at"`
		Region    string `json:"region"`
		Public    bool   `json:"public"`
		Group     string `json:"group"`
		Pinned    bool   `json:"pinned"`
	}
}

func (h *Handler) UpdateBucket(ctx context.Context, input *UpdateBucketInput) (*UpdateBucketOutput, error) {
	info, err := h.store.HeadBucket(input.Bucket)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotFound) {
			return nil, huma.Error404NotFound("bucket not found")
		}
		return nil, huma.Error500InternalServerError("failed to get bucket")
	}

	if input.Body.Public != nil {
		if err := h.store.UpdateBucketPublic(input.Bucket, *input.Body.Public); err != nil {
			return nil, huma.Error500InternalServerError("failed to update bucket metadata")
		}
		info.Public = *input.Body.Public
	}

	if input.Body.Group != nil {
		group := storage.NormalizeGroupName(*input.Body.Group)
		if err := storage.ValidateGroupName(group); err != nil {
			return nil, huma.Error400BadRequest("invalid group name")
		}
		if err := h.store.UpdateBucketGroup(input.Bucket, group); err != nil {
			return nil, huma.Error500InternalServerError("failed to update bucket group")
		}
		info.Group = group
	}

	if input.Body.Pinned != nil {
		if err := h.store.UpdateBucketPinned(input.Bucket, *input.Body.Pinned); err != nil {
			return nil, huma.Error500InternalServerError("failed to update bucket pin")
		}
		info.Pinned = *input.Body.Pinned
	}

	return &UpdateBucketOutput{
		Body: struct {
			Name      string `json:"name"`
			CreatedAt string `json:"created_at"`
			Region    string `json:"region"`
			Public    bool   `json:"public"`
			Group     string `json:"group"`
			Pinned    bool   `json:"pinned"`
		}{
			Name:      info.Name,
			CreatedAt: info.CreatedAt.Format(time.RFC3339),
			Region:    info.Region,
			Public:    info.Public,
			Group:     info.Group,
			Pinned:    info.Pinned,
		},
	}, nil
}
