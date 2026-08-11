package s3

import (
	"encoding/xml"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/protob/auxio/internal/storage"
	"github.com/go-chi/chi/v5"
)

func getKeyFromPath(r *http.Request, bucket string) string {
	path := r.URL.Path
	prefix := "/" + bucket + "/"
	if strings.HasPrefix(path, prefix) {
		return strings.TrimPrefix(path, prefix)
	}
	return ""
}

type Handler struct {
	store  storage.ObjectStore
	region string
}

func NewHandler(store storage.ObjectStore, region string) *Handler {
	return &Handler{
		store:  store,
		region: region,
	}
}

func isValidBucketName(name string) bool {
	return storage.ValidateBucketName(name) == nil
}

func isValidObjectKey(key string) bool {
	if len(key) == 0 || len(key) > 1024 {
		return false
	}
	if strings.HasPrefix(key, "/") {
		return false
	}
	// Folder-marker keys are rejected: a filesystem store cannot hold both a
	// file "docs" and a directory "docs/". Folders are virtual (S3 semantics).
	if strings.HasSuffix(key, "/") {
		return false
	}
	for _, seg := range strings.Split(key, "/") {
		if seg == ".." {
			return false
		}
	}
	return !strings.ContainsFunc(key, func(c rune) bool {
		return c < 0x20 || c == 0x7f
	})
}

func (h *Handler) handleListBuckets(w http.ResponseWriter, r *http.Request) {
	buckets, err := h.store.ListBuckets()
	if err != nil {
		slog.Error("ListBuckets", "err", err)
		writeInternalError(w, err)
		return
	}

	entries := make([]BucketEntry, len(buckets))
	for i, b := range buckets {
		entries[i] = BucketEntry{
			Name:         b.Name,
			CreationDate: b.CreatedAt.UTC().Format(time.RFC3339),
		}
	}

	resp := ListAllMyBucketsResult{
		XMLNS: S3XMLNamespace,
		Owner: Owner{
			ID:          "auxio",
			DisplayName: "auxio",
		},
		Buckets: entries,
	}

	if err := writeXMLResponse(w, resp); err != nil {
		slog.Error("Encode ListBuckets XML", "err", err)
	}
}

func (h *Handler) handleCreateBucket(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	_, err := h.store.CreateBucket(bucket)
	if err != nil {
		// Re-creating an existing bucket returns 200 without mutating it - the
		// x-amz-acl below is deliberately not applied. AWS same-owner semantics.
		if errors.Is(err, storage.ErrBucketExists) {
			w.Header().Set("Location", "/"+bucket)
			w.WriteHeader(http.StatusOK)
			return
		}
		slog.Error("CreateBucket", "bucket", bucket, "err", err)
		writeInternalError(w, err)
		return
	}

	// x-amz-acl: public-read at create time is the only ACL honored - no
	// per-object ACLs, no bucket policies.
	if strings.EqualFold(r.Header.Get("X-Amz-Acl"), "public-read") {
		if err := h.store.UpdateBucketPublic(bucket, true); err != nil {
			slog.Error("set public-read", "bucket", bucket, "err", err)
			writeInternalError(w, err)
			return
		}
	}

	w.Header().Set("Location", "/"+bucket)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleHeadBucket(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	info, err := h.store.HeadBucket(bucket)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("HeadBucket", "bucket", bucket, "err", err)
		writeInternalError(w, err)
		return
	}

	w.Header().Set("x-amz-bucket-region", info.Region)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleDeleteBucket(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	err := h.store.DeleteBucket(bucket)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		if err == storage.ErrBucketNotEmpty {
			writeBucketNotEmpty(w, bucket)
			return
		}
		slog.Error("DeleteBucket", "bucket", bucket, "err", err)
		writeInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleGetBucketLocation(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	_, err := h.store.HeadBucket(bucket)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("GetBucketLocation", "bucket", bucket, "err", err)
		writeInternalError(w, err)
		return
	}

	resp := LocationConstraint{
		XMLNS:  S3XMLNamespace,
		Region: h.region,
	}

	if err := writeXMLResponse(w, resp); err != nil {
		slog.Error("Encode LocationConstraint XML", "err", err)
	}
}

func (h *Handler) dispatchBucketGet(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	switch {
	case q.Has("location"):
		h.handleGetBucketLocation(w, r)
	case q.Has("uploads"):
		h.handleListMultipartUploads(w, r)
	case q.Get("list-type") == "2":
		h.handleListObjectsV2(w, r)
	default:
		h.handleListObjectsV1(w, r)
	}
}

func (h *Handler) dispatchBucketPost(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if q.Has("delete") {
		h.handleDeleteObjects(w, r)
		return
	}

	h.handleCreateMultipartUpload(w, r)
}

func (h *Handler) handleListObjectsV1(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	q := r.URL.Query()
	opts := storage.ListOpts{
		Prefix:    q.Get("prefix"),
		Delimiter: q.Get("delimiter"),
		MaxKeys:   1000,
	}

	if mk := q.Get("max-keys"); mk != "" {
		if v, err := strconv.Atoi(mk); err == nil && v > 0 {
			if v > 1000 {
				v = 1000 // S3 caps max-keys at 1000
			}
			opts.MaxKeys = v
		}
	}

	if marker := q.Get("marker"); marker != "" {
		opts.StartAfter = marker
	}

	result, err := h.store.ListObjects(bucket, opts)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("ListObjectsV1", "bucket", bucket, "err", err)
		writeInternalError(w, err)
		return
	}

	contents := make([]ObjectSummary, len(result.Contents))
	for i, obj := range result.Contents {
		contents[i] = ObjectSummary{
			Key:          obj.Key,
			LastModified: obj.LastModified.UTC().Format(time.RFC3339),
			ETag:         obj.ETag,
			Size:         obj.Size,
			StorageClass: "STANDARD",
		}
	}

	commonPrefixes := make([]CommonPrefix, len(result.CommonPrefixes))
	for i, p := range result.CommonPrefixes {
		commonPrefixes[i] = CommonPrefix{Prefix: p}
	}

	var nextMarker string
	if result.IsTruncated && len(contents) > 0 {
		nextMarker = contents[len(contents)-1].Key
	}

	resp := ListBucketResultV1{
		XMLNS:          S3XMLNamespace,
		Name:           bucket,
		Prefix:         opts.Prefix,
		Delimiter:      opts.Delimiter,
		MaxKeys:        opts.MaxKeys,
		IsTruncated:    result.IsTruncated,
		Marker:         q.Get("marker"),
		NextMarker:     nextMarker,
		Contents:       contents,
		CommonPrefixes: commonPrefixes,
	}

	if err := writeXMLResponse(w, resp); err != nil {
		slog.Error("Encode ListObjectsV1 XML", "err", err)
	}
}

func (h *Handler) handleListObjectsV2(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	q := r.URL.Query()
	opts := storage.ListOpts{
		Prefix:    q.Get("prefix"),
		Delimiter: q.Get("delimiter"),
		MaxKeys:   1000,
	}

	if mk := q.Get("max-keys"); mk != "" {
		if v, err := strconv.Atoi(mk); err == nil && v > 0 {
			if v > 1000 {
				v = 1000 // S3 caps max-keys at 1000
			}
			opts.MaxKeys = v
		}
	}

	opts.ContinuationToken = q.Get("continuation-token")
	opts.StartAfter = q.Get("start-after")

	result, err := h.store.ListObjects(bucket, opts)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("ListObjectsV2", "bucket", bucket, "err", err)
		writeInternalError(w, err)
		return
	}

	contents := make([]ObjectSummary, len(result.Contents))
	for i, obj := range result.Contents {
		contents[i] = ObjectSummary{
			Key:          obj.Key,
			LastModified: obj.LastModified.UTC().Format(time.RFC3339),
			ETag:         obj.ETag,
			Size:         obj.Size,
			StorageClass: "STANDARD",
		}
	}

	commonPrefixes := make([]CommonPrefix, len(result.CommonPrefixes))
	for i, p := range result.CommonPrefixes {
		commonPrefixes[i] = CommonPrefix{Prefix: p}
	}

	resp := ListBucketResultV2{
		XMLNS:                 S3XMLNamespace,
		Name:                  bucket,
		Prefix:                opts.Prefix,
		Delimiter:             opts.Delimiter,
		KeyCount:              result.KeyCount,
		MaxKeys:               opts.MaxKeys,
		IsTruncated:           result.IsTruncated,
		ContinuationToken:     opts.ContinuationToken,
		NextContinuationToken: result.NextContinuationToken,
		StartAfter:            opts.StartAfter,
		Contents:              contents,
		CommonPrefixes:        commonPrefixes,
	}

	if err := writeXMLResponse(w, resp); err != nil {
		slog.Error("Encode ListObjectsV2 XML", "err", err)
	}
}

func (h *Handler) handleDeleteObjects(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	var req DeleteObjectsRequest
	if err := xml.NewDecoder(r.Body).Decode(&req); err != nil {
		writeMalformedXML(w)
		return
	}

	if len(req.Objects) == 0 {
		writeInvalidArgument(w, "You must specify at least one object to delete.")
		return
	}

	keys := make([]string, len(req.Objects))
	for i, obj := range req.Objects {
		keys[i] = obj.Key
	}

	results, err := h.store.DeleteObjects(bucket, keys)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("DeleteObjects", "bucket", bucket, "err", err)
		writeInternalError(w, err)
		return
	}

	deleted := make([]DeletedObject, 0, len(results))
	var errs []DeleteError
	for _, r := range results {
		if r.Success {
			deleted = append(deleted, DeletedObject{Key: r.Key})
		} else {
			errs = append(errs, DeleteError{Key: r.Key, Code: "InternalError", Message: r.Error})
		}
	}

	resp := DeleteResult{
		XMLNS:   S3XMLNamespace,
		Deleted: deleted,
		Errors:  errs,
	}

	if err := writeXMLResponse(w, resp); err != nil {
		slog.Error("Encode DeleteResult XML", "err", err)
	}
}

func (h *Handler) handlePutObject(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := getKeyFromPath(r, bucket)

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	if !isValidObjectKey(key) {
		writeInvalidObjectName(w, bucket, key)
		return
	}

	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	userMetadata := make(map[string]string)
	for k, v := range r.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-amz-meta-") {
			metaKey := k
			if len(v) > 0 {
				userMetadata[metaKey] = v[0]
			}
		}
	}

	meta := storage.ObjectMeta{
		ContentType:  contentType,
		UserMetadata: userMetadata,
	}

	info, err := h.store.PutObject(bucket, key, DecodeChunked(r), meta)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("PutObject", "bucket", bucket, "key", key, "err", err)
		writeInternalError(w, err)
		return
	}

	w.Header().Set("ETag", info.ETag)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleGetObject(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	if q.Has("uploadId") {
		h.handleListParts(w, r)
		return
	}

	bucket := chi.URLParam(r, "bucket")
	key := getKeyFromPath(r, bucket)

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	if !isValidObjectKey(key) {
		writeInvalidObjectName(w, bucket, key)
		return
	}

	reader, info, err := h.store.GetObject(bucket, key)
	if err != nil {
		if err == storage.ErrObjectNotFound {
			writeNoSuchKey(w, bucket, key)
			return
		}
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("GetObject", "bucket", bucket, "key", key, "err", err)
		writeInternalError(w, err)
		return
	}
	defer reader.Close()

	contentType := info.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("ETag", info.ETag)

	for k, v := range info.UserMetadata {
		w.Header().Set(k, v)
	}

	// ServeContent gives Range, conditional GETs, 304 and 416 for free, and owns
	// Content-Length / Accept-Ranges / status - do not set those above. The copy
	// below is unreachable while FilesystemStore (*os.File) is the only store.
	if rs, ok := reader.(io.ReadSeeker); ok {
		http.ServeContent(w, r, key, info.LastModified, rs)
		return
	}

	w.Header().Set("Last-Modified", info.LastModified.UTC().Format(http.TimeFormat))
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
	w.WriteHeader(http.StatusOK)
	io.Copy(w, reader)
}

func (h *Handler) handleHeadObject(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := getKeyFromPath(r, bucket)

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	if !isValidObjectKey(key) {
		writeInvalidObjectName(w, bucket, key)
		return
	}

	info, err := h.store.HeadObject(bucket, key)
	if err != nil {
		if err == storage.ErrObjectNotFound {
			writeNoSuchKey(w, bucket, key)
			return
		}
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("HeadObject", "bucket", bucket, "key", key, "err", err)
		writeInternalError(w, err)
		return
	}

	contentType := info.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size, 10))
	w.Header().Set("ETag", info.ETag)
	w.Header().Set("Last-Modified", info.LastModified.UTC().Format(http.TimeFormat))
	w.Header().Set("Accept-Ranges", "bytes")

	for k, v := range info.UserMetadata {
		w.Header().Set(k, v)
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleDeleteObject(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	key := getKeyFromPath(r, bucket)

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	if !isValidObjectKey(key) {
		writeInvalidObjectName(w, bucket, key)
		return
	}

	err := h.store.DeleteObject(bucket, key)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("DeleteObject", "bucket", bucket, "key", key, "err", err)
		writeInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) dispatchObjectPut(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	uploadID := q.Get("uploadId")
	partNumber := q.Get("partNumber")

	if uploadID != "" && partNumber != "" {
		h.handleUploadPart(w, r)
		return
	}

	h.handlePutObject(w, r)
}

func (h *Handler) dispatchObjectDelete(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if q.Has("uploadId") {
		h.handleAbortMultipartUpload(w, r)
		return
	}

	h.handleDeleteObject(w, r)
}

func (h *Handler) dispatchObjectPost(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	if q.Has("uploads") {
		h.handleCreateMultipartUpload(w, r)
		return
	}

	if q.Has("uploadId") {
		h.handleCompleteMultipartUpload(w, r)
		return
	}

	writeNotImplemented(w, "ObjectPost")
}
