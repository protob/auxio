package storage

import (
	"errors"
	"strings"
	"testing"
)

func TestValidateBucketName(t *testing.T) {
	valid := []string{"abc", "my-bucket", "a.b.c", "bucket123", strings.Repeat("a", 63)}
	for _, name := range valid {
		if err := ValidateBucketName(name); err != nil {
			t.Errorf("ValidateBucketName(%q) = %v, want nil", name, err)
		}
	}

	invalid := []string{
		"",
		"ab",
		strings.Repeat("a", 64),
		"My-Bucket",
		"-bucket",
		"bucket-",
		".bucket",
		"bucket.",
		"my..bucket",
		"my.-bucket",  // dot-dash adjacency
		"my-.bucket",  // dash-dot adjacency
		"192.168.1.1", // IPv4
		"bucket/../escape",
	}
	for _, name := range invalid {
		if err := ValidateBucketName(name); !errors.Is(err, ErrInvalidBucketName) {
			t.Errorf("ValidateBucketName(%q) = %v, want ErrInvalidBucketName", name, err)
		}
	}
}

func TestObjectPathContainment(t *testing.T) {
	const dataDir = "/data/auxio"
	const bucket = "bucket"

	ok := map[string]string{
		"ok/key.txt":  "/data/auxio/bucket/ok/key.txt",
		"a/b/c.txt":   "/data/auxio/bucket/a/b/c.txt",
		"single":      "/data/auxio/bucket/single",
	}
	for key, want := range ok {
		got, err := ObjectPath(dataDir, bucket, key)
		if err != nil {
			t.Errorf("ObjectPath(%q) unexpected error %v", key, err)
		}
		if got != want {
			t.Errorf("ObjectPath(%q) = %q, want %q", key, got, want)
		}
	}

	escapes := []string{"../../etc/passwd", "a/../../b", "..", "../sibling"}
	for _, key := range escapes {
		if _, err := ObjectPath(dataDir, bucket, key); !errors.Is(err, ErrPathTraversal) {
			t.Errorf("ObjectPath(%q) err = %v, want ErrPathTraversal", key, err)
		}
	}
}

func TestValidateGroupName(t *testing.T) {
	cases := []struct {
		in   string
		norm string
		ok   bool
	}{
		{"", "", true},
		{"   ", "", true},
		{"ECOM ", "ecom", true},
		{"e-com_1 a", "e-com_1 a", true},
		{"-bad", "-bad", false},
		{"bad!", "bad!", false},
		{strings.Repeat("a", 41), strings.Repeat("a", 41), false}, // >40
	}
	for _, c := range cases {
		if n := NormalizeGroupName(c.in); n != c.norm {
			t.Errorf("NormalizeGroupName(%q) = %q, want %q", c.in, n, c.norm)
		}
		if err := ValidateGroupName(NormalizeGroupName(c.in)); (err == nil) != c.ok {
			t.Errorf("ValidateGroupName(norm %q): ok=%v, want %v", c.in, err == nil, c.ok)
		}
	}
}
