package s3

import (
	"encoding/xml"
	"net/http"

	"github.com/google/uuid"
)

func writeXMLResponse(w http.ResponseWriter, v any) error {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(xml.Header))
	if err != nil {
		return err
	}
	return xml.NewEncoder(w).Encode(v)
}

func writeS3Error(w http.ResponseWriter, code string, message string, resource string, httpStatus int) {
	requestID := uuid.New().String()
	err := S3Error{
		Code:      code,
		Message:   message,
		Resource:  resource,
		RequestID: requestID,
	}

	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(httpStatus)
	_, _ = w.Write([]byte(xml.Header))
	_ = xml.NewEncoder(w).Encode(err)
}

func writeNoSuchBucket(w http.ResponseWriter, bucket string) {
	writeS3Error(w, "NoSuchBucket", "The specified bucket does not exist", "/"+bucket, http.StatusNotFound)
}

func writeNoSuchKey(w http.ResponseWriter, bucket, key string) {
	writeS3Error(w, "NoSuchKey", "The specified key does not exist", "/"+bucket+"/"+key, http.StatusNotFound)
}

func writeInternalError(w http.ResponseWriter, err error) {
	writeS3Error(w, "InternalError", "We encountered an internal error. Please try again.", "", http.StatusInternalServerError)
}

func writeMalformedXML(w http.ResponseWriter) {
	writeS3Error(w, "MalformedXML", "The XML you provided was not well-formed or did not validate against our published schema.", "", http.StatusBadRequest)
}

func writeAccessDenied(w http.ResponseWriter) {
	writeS3Error(w, "AccessDenied", "Access Denied", "", http.StatusForbidden)
}

func writeNotImplemented(w http.ResponseWriter, op string) {
	writeS3Error(w, "NotImplemented", op+" is not implemented.", "", http.StatusNotImplemented)
}

func writeInvalidBucketName(w http.ResponseWriter, bucket string) {
	writeS3Error(w, "InvalidBucketName", "The specified bucket is not valid.", "/"+bucket, http.StatusBadRequest)
}

func writeInvalidObjectName(w http.ResponseWriter, bucket, key string) {
	writeS3Error(w, "InvalidObjectName", "The specified key is not valid.", "/"+bucket+"/"+key, http.StatusBadRequest)
}

func writeBucketNotEmpty(w http.ResponseWriter, bucket string) {
	writeS3Error(w, "BucketNotEmpty", "The bucket you tried to delete is not empty.", "/"+bucket, http.StatusConflict)
}

func writeNoSuchUpload(w http.ResponseWriter, uploadID string) {
	writeS3Error(w, "NoSuchUpload", "The specified multipart upload does not exist.", "", http.StatusNotFound)
}

func writeInvalidArgument(w http.ResponseWriter, message string) {
	writeS3Error(w, "InvalidArgument", message, "", http.StatusBadRequest)
}

func writeInvalidPart(w http.ResponseWriter) {
	writeS3Error(w, "InvalidPart", "One or more of the specified parts could not be found. The part may not have been uploaded, or the specified entity tag may not match the part's entity tag.", "", http.StatusBadRequest)
}
