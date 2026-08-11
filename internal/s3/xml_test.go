package s3

import (
	"encoding/xml"
	"strings"
	"testing"
)

// Storage delete failures are hard to provoke portably, so pin the response
// shape at the marshal level: errors serialize as <Error> entries and a fully
// successful batch emits none.
func TestDeleteResultMarshalsErrors(t *testing.T) {
	withErr := DeleteResult{
		XMLNS:   S3XMLNamespace,
		Deleted: []DeletedObject{{Key: "ok.txt"}},
		Errors:  []DeleteError{{Key: "bad.txt", Code: "InternalError", Message: "disk on fire"}},
	}
	out, err := xml.Marshal(withErr)
	if err != nil {
		t.Fatal(err)
	}
	s := string(out)
	for _, want := range []string{"<Deleted><Key>ok.txt</Key></Deleted>", "<Error><Key>bad.txt</Key>", "<Code>InternalError</Code>"} {
		if !strings.Contains(s, want) {
			t.Errorf("marshal missing %q in %s", want, s)
		}
	}

	clean := DeleteResult{XMLNS: S3XMLNamespace, Deleted: []DeletedObject{{Key: "ok.txt"}}}
	out, err = xml.Marshal(clean)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(out), "<Error>") {
		t.Errorf("successful batch must not emit <Error> entries: %s", out)
	}
}
