package s3

import "testing"

func TestIsValidObjectKey(t *testing.T) {
	valid := []string{"file.txt", "a/b/c.png", "deep/nested/key", "with space.txt", "ünïcode.txt"}
	for _, k := range valid {
		if !isValidObjectKey(k) {
			t.Errorf("isValidObjectKey(%q) = false, want true", k)
		}
	}

	invalid := []string{
		"",
		"/leading",
		"trailing/", // folders are virtual
		"docs/sub/",
		"../escape",
		"a/../../b",
		"dir/../../etc/passwd",
		"with\x00null",
		"with\nnewline",
	}
	for _, k := range invalid {
		if isValidObjectKey(k) {
			t.Errorf("isValidObjectKey(%q) = true, want false", k)
		}
	}
}
