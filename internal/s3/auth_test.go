package s3

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

const (
	tAccess = "AKIATESTKEY"
	tSecret = "testsecretkey"
	tRegion = "us-east-1"
)

var emptyPayloadHash = hex.EncodeToString(func() []byte { s := sha256.Sum256(nil); return s[:] }())

// signedRequest returns a GET request signed for the given time. The same
// *http.Request goes to VerifyRequest, so Host and headers stay identical
// between signing and verification.
func signedRequest(t *testing.T, at time.Time) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "http://auxio.local/mybucket", nil)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("X-Amz-Content-Sha256", emptyPayloadHash)
	signer := v4.NewSigner()
	creds := aws.Credentials{AccessKeyID: tAccess, SecretAccessKey: tSecret}
	if err := signer.SignHTTP(context.Background(), creds, req, emptyPayloadHash, "s3", tRegion, at); err != nil {
		t.Fatalf("SignHTTP: %v", err)
	}
	return req
}

func TestVerifyRequest(t *testing.T) {
	engine := NewAuthEngine(tAccess, tSecret, tRegion)

	t.Run("valid", func(t *testing.T) {
		if err := engine.VerifyRequest(signedRequest(t, time.Now().UTC())); err != nil {
			t.Fatalf("valid request rejected: %v", err)
		}
	})

	t.Run("expired timestamp", func(t *testing.T) {
		// Signature is valid for the old time, but the ±15-min window rejects it.
		err := engine.VerifyRequest(signedRequest(t, time.Now().UTC().Add(-30*time.Minute)))
		if err == nil || !strings.Contains(err.Error(), "window") {
			t.Fatalf("expired: got %v, want window error", err)
		}
	})

	t.Run("missing authorization", func(t *testing.T) {
		req := signedRequest(t, time.Now().UTC())
		req.Header.Del("Authorization")
		if err := engine.VerifyRequest(req); err == nil {
			t.Fatal("missing Authorization should fail")
		}
	})

	t.Run("missing x-amz-date", func(t *testing.T) {
		req := signedRequest(t, time.Now().UTC())
		req.Header.Del("X-Amz-Date")
		if err := engine.VerifyRequest(req); err == nil {
			t.Fatal("missing X-Amz-Date should fail")
		}
	})

	t.Run("missing content-sha256", func(t *testing.T) {
		req := signedRequest(t, time.Now().UTC())
		req.Header.Del("X-Amz-Content-Sha256")
		if err := engine.VerifyRequest(req); err == nil {
			t.Fatal("missing X-Amz-Content-Sha256 should fail")
		}
	})

	t.Run("malformed prefix", func(t *testing.T) {
		req := signedRequest(t, time.Now().UTC())
		req.Header.Set("Authorization", "Bearer nonsense")
		if err := engine.VerifyRequest(req); err == nil {
			t.Fatal("non-AWS4 Authorization should fail")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		other := NewAuthEngine(tAccess, "wrongsecret", tRegion)
		if err := other.VerifyRequest(signedRequest(t, time.Now().UTC())); err == nil {
			t.Fatal("signature computed with a different secret should fail")
		}
	})

	t.Run("wrong access key", func(t *testing.T) {
		other := NewAuthEngine("AKIAOTHER", tSecret, tRegion)
		if err := other.VerifyRequest(signedRequest(t, time.Now().UTC())); err == nil {
			t.Fatal("mismatched access key should fail")
		}
	})

	t.Run("tampered signature", func(t *testing.T) {
		req := signedRequest(t, time.Now().UTC())
		auth := req.Header.Get("Authorization")
		// Signature= is the last Authorization parameter, so the final byte is
		// signature hex - flipping it leaves the header parseable.
		last := auth[len(auth)-1]
		repl := byte('0')
		if last == '0' {
			repl = '1'
		}
		req.Header.Set("Authorization", auth[:len(auth)-1]+string(repl))
		if err := engine.VerifyRequest(req); err == nil {
			t.Fatal("tampered signature should fail")
		}
	})
}
