package dashboard

import (
	"context"
	"errors"
	"time"

	"github.com/protob/auxio/internal/storage"
	"github.com/danielgtaylor/huma/v2"
)

type ObjectSummary struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
	LastModified string `json:"last_modified"`
	ETag         string `json:"etag"`
}

type ListObjectsInput struct {
	Bucket            string `path:"bucket"`
	Prefix            string `query:"prefix"`
	Delimiter         string `query:"delimiter"`
	MaxKeys           int    `query:"max_keys"`
	ContinuationToken string `query:"continuation_token"`
}

type ListObjectsOutput struct {
	Body struct {
		Objects               []ObjectSummary `json:"objects"`
		CommonPrefixes        []string        `json:"common_prefixes"`
		IsTruncated           bool            `json:"is_truncated"`
		NextContinuationToken string          `json:"next_continuation_token"`
	}
}

func (h *Handler) ListObjects(ctx context.Context, input *ListObjectsInput) (*ListObjectsOutput, error) {
	opts := storage.ListOpts{
		Prefix:            input.Prefix,
		Delimiter:         input.Delimiter,
		MaxKeys:           input.MaxKeys,
		ContinuationToken: input.ContinuationToken,
	}

	result, err := h.store.ListObjects(input.Bucket, opts)
	if err != nil {
		if errors.Is(err, storage.ErrBucketNotFound) {
			return nil, huma.Error404NotFound("bucket not found")
		}
		return nil, huma.Error500InternalServerError("failed to list objects")
	}

	objects := make([]ObjectSummary, len(result.Contents))
	for i, obj := range result.Contents {
		objects[i] = ObjectSummary{
			Key:          obj.Key,
			Size:         obj.Size,
			ContentType:  obj.ContentType,
			LastModified: obj.LastModified.Format(time.RFC3339),
			ETag:         obj.ETag,
		}
	}

	output := &ListObjectsOutput{}
	output.Body.Objects = objects
	output.Body.CommonPrefixes = result.CommonPrefixes
	output.Body.IsTruncated = result.IsTruncated
	output.Body.NextContinuationToken = result.NextContinuationToken

	return output, nil
}

type DeleteObjectsInput struct {
	Bucket string `path:"bucket"`
	Body   struct {
		Keys []string `json:"keys"`
	}
}

type DeleteObjectsOutput struct {
	Body struct {
		Deleted []string         `json:"deleted"`
		Errors  []DeleteObjError `json:"errors"`
	}
}

type DeleteObjError struct {
	Key   string `json:"key"`
	Error string `json:"error"`
}

func (h *Handler) DeleteObjects(ctx context.Context, input *DeleteObjectsInput) (*DeleteObjectsOutput, error) {
	results, err := h.store.DeleteObjects(input.Bucket, input.Body.Keys)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to delete objects")
	}

	output := &DeleteObjectsOutput{}
	for _, r := range results {
		if r.Success {
			output.Body.Deleted = append(output.Body.Deleted, r.Key)
		} else {
			output.Body.Errors = append(output.Body.Errors, DeleteObjError{
				Key:   r.Key,
				Error: r.Error,
			})
		}
	}

	return output, nil
}
