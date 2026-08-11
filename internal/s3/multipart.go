package s3

import (
	"encoding/xml"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/protob/auxio/internal/storage"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) handleCreateMultipartUpload(w http.ResponseWriter, r *http.Request) {
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

	// x-amz-meta-* is not persisted for multipart uploads; plain PUT keeps it.
	meta := storage.ObjectMeta{
		ContentType: contentType,
	}

	uploadID, err := h.store.CreateMultipartUpload(bucket, key, meta)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("CreateMultipartUpload", "bucket", bucket, "key", key, "err", err)
		writeInternalError(w, err)
		return
	}

	resp := InitiateMultipartUploadResult{
		XMLNS:    S3XMLNamespace,
		Bucket:   bucket,
		Key:      key,
		UploadID: uploadID,
	}

	if err := writeXMLResponse(w, resp); err != nil {
		slog.Error("Encode CreateMultipartUpload XML", "err", err)
	}
}

func (h *Handler) handleUploadPart(w http.ResponseWriter, r *http.Request) {
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

	q := r.URL.Query()
	uploadID := q.Get("uploadId")
	partNumberStr := q.Get("partNumber")

	if uploadID == "" || partNumberStr == "" {
		writeInvalidArgument(w, "Missing uploadId or partNumber")
		return
	}

	partNumber, err := strconv.Atoi(partNumberStr)
	if err != nil || partNumber < 1 || partNumber > 10000 {
		writeInvalidArgument(w, "Invalid part number")
		return
	}

	body := DecodeChunked(r)

	partInfo, err := h.store.UploadPart(bucket, key, uploadID, partNumber, body)
	if err != nil {
		if err == storage.ErrUploadNotFound {
			writeNoSuchUpload(w, uploadID)
			return
		}
		slog.Error("UploadPart", "bucket", bucket, "key", key, "uploadId", uploadID, "part", partNumber, "err", err)
		writeInternalError(w, err)
		return
	}

	w.Header().Set("ETag", partInfo.ETag)
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) handleCompleteMultipartUpload(w http.ResponseWriter, r *http.Request) {
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

	q := r.URL.Query()
	uploadID := q.Get("uploadId")

	if uploadID == "" {
		writeInvalidArgument(w, "Missing uploadId")
		return
	}

	var req CompleteMultipartUpload
	if err := xml.NewDecoder(r.Body).Decode(&req); err != nil {
		writeMalformedXML(w)
		return
	}

	if len(req.Parts) == 0 {
		writeInvalidArgument(w, "Must specify at least one part")
		return
	}

	parts := make([]storage.PartInfo, len(req.Parts))
	for i, p := range req.Parts {
		parts[i] = storage.PartInfo{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		}
	}

	info, err := h.store.CompleteMultipartUpload(bucket, key, uploadID, parts)
	if err != nil {
		if err == storage.ErrUploadNotFound {
			writeNoSuchUpload(w, uploadID)
			return
		}
		if err == storage.ErrInvalidPart {
			writeInvalidPart(w)
			return
		}
		slog.Error("CompleteMultipartUpload", "bucket", bucket, "key", key, "uploadId", uploadID, "err", err)
		writeInternalError(w, err)
		return
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	if xf := r.Header.Get("X-Forwarded-Proto"); xf != "" {
		scheme = xf
	}
	location := scheme + "://" + r.Host + "/" + bucket + "/" + key

	resp := CompleteMultipartUploadResult{
		XMLNS:    S3XMLNamespace,
		Location: location,
		Bucket:   bucket,
		Key:      key,
		ETag:     info.ETag,
	}

	if err := writeXMLResponse(w, resp); err != nil {
		slog.Error("Encode CompleteMultipartUpload XML", "err", err)
	}
}

func (h *Handler) handleAbortMultipartUpload(w http.ResponseWriter, r *http.Request) {
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

	q := r.URL.Query()
	uploadID := q.Get("uploadId")

	if uploadID == "" {
		writeInvalidArgument(w, "Missing uploadId")
		return
	}

	err := h.store.AbortMultipartUpload(bucket, key, uploadID)
	if err != nil {
		if err == storage.ErrUploadNotFound {
			writeNoSuchUpload(w, uploadID)
			return
		}
		slog.Error("AbortMultipartUpload", "bucket", bucket, "key", key, "uploadId", uploadID, "err", err)
		writeInternalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleListParts(w http.ResponseWriter, r *http.Request) {
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

	q := r.URL.Query()
	uploadID := q.Get("uploadId")

	if uploadID == "" {
		writeInvalidArgument(w, "Missing uploadId")
		return
	}

	parts, err := h.store.ListParts(bucket, key, uploadID)
	if err != nil {
		if err == storage.ErrUploadNotFound {
			writeNoSuchUpload(w, uploadID)
			return
		}
		slog.Error("ListParts", "bucket", bucket, "key", key, "uploadId", uploadID, "err", err)
		writeInternalError(w, err)
		return
	}

	listParts := make([]ListPartsPart, len(parts))
	for i, p := range parts {
		listParts[i] = ListPartsPart{
			PartNumber: p.PartNumber,
			// PartInfo carries no mtime and the XML field is always emitted, so
			// this is a placeholder.
			LastModified: time.Now().UTC().Format(time.RFC3339),
			ETag:         p.ETag,
			Size:         p.Size,
		}
	}

	resp := ListPartsResult{
		XMLNS:    S3XMLNamespace,
		Bucket:   bucket,
		Key:      key,
		UploadID: uploadID,
		MaxParts: 1000,
		Parts:    listParts,
	}

	if err := writeXMLResponse(w, resp); err != nil {
		slog.Error("Encode ListParts XML", "err", err)
	}
}

func (h *Handler) handleListMultipartUploads(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")

	if !isValidBucketName(bucket) {
		writeInvalidBucketName(w, bucket)
		return
	}

	uploads, err := h.store.ListMultipartUploads(bucket)
	if err != nil {
		if err == storage.ErrBucketNotFound {
			writeNoSuchBucket(w, bucket)
			return
		}
		slog.Error("ListMultipartUploads", "bucket", bucket, "err", err)
		writeInternalError(w, err)
		return
	}

	uploadInfos := make([]MultipartUploadInfo, len(uploads))
	for i, u := range uploads {
		uploadInfos[i] = MultipartUploadInfo{
			Key:      u.Key,
			UploadID: u.UploadID,
			Initiator: Owner{
				ID:          "auxio",
				DisplayName: "auxio",
			},
			Owner: Owner{
				ID:          "auxio",
				DisplayName: "auxio",
			},
			StorageClass: "STANDARD",
			Initiated:    u.Initiated.Format(time.RFC3339),
		}
	}

	resp := ListMultipartUploadsResult{
		XMLNS:      S3XMLNamespace,
		Bucket:     bucket,
		MaxUploads: 1000,
		Uploads:    uploadInfos,
	}

	if err := writeXMLResponse(w, resp); err != nil {
		slog.Error("Encode ListMultipartUploads XML", "err", err)
	}
}
