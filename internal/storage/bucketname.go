package storage

import (
	"errors"
	"regexp"
	"strings"
)

// ErrInvalidBucketName is returned by ValidateBucketName for any name that
// violates the S3 bucket-naming rules.
var ErrInvalidBucketName = errors.New("invalid bucket name")

var (
	bucketNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)
	ipv4Pattern       = regexp.MustCompile(`^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9][0-9]|[0-9])$`)
)

// ValidateBucketName enforces the S3 naming rules: 3-63 chars, lowercase
// alphanumerics/dots/dashes, alphanumeric at both ends, no "..", no "." next to
// "-", and not an IPv4 address. FilesystemStore.CreateBucket calls it, so no
// path into the store can create a name the S3 API would reject.
func ValidateBucketName(name string) error {
	if !bucketNamePattern.MatchString(name) {
		return ErrInvalidBucketName
	}
	if strings.Contains(name, "..") {
		return ErrInvalidBucketName
	}
	for i := 1; i < len(name); i++ {
		if (name[i-1] == '.' && name[i] == '-') || (name[i-1] == '-' && name[i] == '.') {
			return ErrInvalidBucketName
		}
	}
	if ipv4Pattern.MatchString(name) {
		return ErrInvalidBucketName
	}
	return nil
}

// ErrInvalidGroupName is returned by ValidateGroupName for a malformed group label.
var ErrInvalidGroupName = errors.New("invalid group name")

// groupNamePattern matches a post-NormalizeGroupName label: 1-40 chars, lowercase
// alphanumerics plus space/underscore/hyphen, alphanumeric start.
var groupNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9 _-]{0,39}$`)

// NormalizeGroupName lowercases and trims; "" means ungrouped.
func NormalizeGroupName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// ValidateGroupName accepts "" (ungrouped) or a normalized label. Run
// NormalizeGroupName first - the pattern only matches lowercase.
func ValidateGroupName(s string) error {
	if s == "" {
		return nil
	}
	if !groupNamePattern.MatchString(s) {
		return ErrInvalidGroupName
	}
	return nil
}
