package s3

import "encoding/xml"

const S3XMLNamespace = "http://s3.amazonaws.com/doc/2006-03-01/"

type Owner struct {
	ID          string `xml:"ID"`
	DisplayName string `xml:"DisplayName"`
}

type BucketEntry struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

type ListAllMyBucketsResult struct {
	XMLName xml.Name      `xml:"ListAllMyBucketsResult"`
	XMLNS   string        `xml:"xmlns,attr"`
	Owner   Owner         `xml:"Owner"`
	Buckets []BucketEntry `xml:"Buckets>Bucket"`
}

type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

type ObjectSummary struct {
	Key          string `xml:"Key"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
	StorageClass string `xml:"StorageClass"`
}

type ListBucketResultV1 struct {
	XMLName        xml.Name        `xml:"ListBucketResult"`
	XMLNS          string          `xml:"xmlns,attr"`
	Name           string          `xml:"Name"`
	Prefix         string          `xml:"Prefix,omitempty"`
	Delimiter      string          `xml:"Delimiter,omitempty"`
	MaxKeys        int             `xml:"MaxKeys"`
	IsTruncated    bool            `xml:"IsTruncated"`
	Marker         string          `xml:"Marker,omitempty"`
	NextMarker     string          `xml:"NextMarker,omitempty"`
	Contents       []ObjectSummary `xml:"Contents"`
	CommonPrefixes []CommonPrefix  `xml:"CommonPrefixes,omitempty"`
}

// Same <ListBucketResult> element as V1 - S3 keeps the element name and
// changes the children.
type ListBucketResultV2 struct {
	XMLName               xml.Name        `xml:"ListBucketResult"`
	XMLNS                 string          `xml:"xmlns,attr"`
	Name                  string          `xml:"Name"`
	Prefix                string          `xml:"Prefix,omitempty"`
	Delimiter             string          `xml:"Delimiter,omitempty"`
	KeyCount              int             `xml:"KeyCount"`
	MaxKeys               int             `xml:"MaxKeys"`
	IsTruncated           bool            `xml:"IsTruncated"`
	ContinuationToken     string          `xml:"ContinuationToken,omitempty"`
	NextContinuationToken string          `xml:"NextContinuationToken,omitempty"`
	StartAfter            string          `xml:"StartAfter,omitempty"`
	Contents              []ObjectSummary `xml:"Contents"`
	CommonPrefixes        []CommonPrefix  `xml:"CommonPrefixes,omitempty"`
}

type S3Error struct {
	XMLName   xml.Name `xml:"Error"`
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	Resource  string   `xml:"Resource"`
	RequestID string   `xml:"RequestId"`
}

type LocationConstraint struct {
	XMLName xml.Name `xml:"LocationConstraint"`
	XMLNS   string   `xml:"xmlns,attr"`
	Region  string   `xml:",chardata"`
}

type InitiateMultipartUploadResult struct {
	XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
	XMLNS    string   `xml:"xmlns,attr"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	UploadID string   `xml:"UploadId"`
}

type CompletedPart struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

type CompleteMultipartUpload struct {
	XMLName xml.Name        `xml:"CompleteMultipartUpload"`
	Parts   []CompletedPart `xml:"Part"`
}

type CompleteMultipartUploadResult struct {
	XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
	XMLNS    string   `xml:"xmlns,attr"`
	Location string   `xml:"Location"`
	Bucket   string   `xml:"Bucket"`
	Key      string   `xml:"Key"`
	ETag     string   `xml:"ETag"`
}

type ListPartsPart struct {
	PartNumber   int    `xml:"PartNumber"`
	LastModified string `xml:"LastModified"`
	ETag         string `xml:"ETag"`
	Size         int64  `xml:"Size"`
}

type ListPartsResult struct {
	XMLName              xml.Name        `xml:"ListPartsResult"`
	XMLNS                string          `xml:"xmlns,attr"`
	Bucket               string          `xml:"Bucket"`
	Key                  string          `xml:"Key"`
	UploadID             string          `xml:"UploadId"`
	PartNumberMarker     int             `xml:"PartNumberMarker"`
	NextPartNumberMarker int             `xml:"NextPartNumberMarker"`
	MaxParts             int             `xml:"MaxParts"`
	IsTruncated          bool            `xml:"IsTruncated"`
	Parts                []ListPartsPart `xml:"Part"`
}

type MultipartUploadInfo struct {
	Key          string `xml:"Key"`
	UploadID     string `xml:"UploadId"`
	Initiator    Owner  `xml:"Initiator"`
	Owner        Owner  `xml:"Owner"`
	StorageClass string `xml:"StorageClass"`
	Initiated    string `xml:"Initiated"`
}

type ListMultipartUploadsResult struct {
	XMLName            xml.Name              `xml:"ListMultipartUploadsResult"`
	XMLNS              string                `xml:"xmlns,attr"`
	Bucket             string                `xml:"Bucket"`
	KeyMarker          string                `xml:"KeyMarker,omitempty"`
	UploadIDMarker     string                `xml:"UploadIdMarker,omitempty"`
	NextKeyMarker      string                `xml:"NextKeyMarker,omitempty"`
	NextUploadIDMarker string                `xml:"NextUploadIdMarker,omitempty"`
	MaxUploads         int                   `xml:"MaxUploads"`
	IsTruncated        bool                  `xml:"IsTruncated"`
	Prefix             string                `xml:"Prefix,omitempty"`
	Delimiter          string                `xml:"Delimiter,omitempty"`
	CommonPrefixes     []CommonPrefix        `xml:"CommonPrefixes,omitempty"`
	Uploads            []MultipartUploadInfo `xml:"Upload"`
}

type DeleteObject struct {
	Key string `xml:"Key"`
}

type DeleteObjectsRequest struct {
	XMLName xml.Name       `xml:"Delete"`
	Objects []DeleteObject `xml:"Object"`
}

type DeletedObject struct {
	Key string `xml:"Key"`
}

type DeleteError struct {
	Key     string `xml:"Key"`
	Code    string `xml:"Code"`
	Message string `xml:"Message"`
}

type DeleteResult struct {
	XMLName xml.Name        `xml:"DeleteResult"`
	XMLNS   string          `xml:"xmlns,attr"`
	Deleted []DeletedObject `xml:"Deleted"`
	Errors  []DeleteError   `xml:"Error,omitempty"`
}
