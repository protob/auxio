package s3

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const AWSv4Prefix = "AWS4-HMAC-SHA256 "

type AuthEngine struct {
	AccessKey string
	SecretKey string
	Region    string
}

func NewAuthEngine(accessKey, secretKey, region string) *AuthEngine {
	return &AuthEngine{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Region:    region,
	}
}

func awsURLEncode(s string, encodeSlash bool) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
			continue
		}
		if c == '/' && !encodeSlash {
			b.WriteByte(c)
			continue
		}
		b.WriteString("%")
		b.WriteString(strings.ToUpper(hex.EncodeToString([]byte{c})))
	}
	return b.String()
}

func canonicalQueryString(u *url.URL) string {
	if u.RawQuery == "" {
		return ""
	}

	values := u.Query()
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		// SigV4 excludes X-Amz-Signature from the canonical query string. Header-signed
		// requests never carry it - VerifyRequest routes those to verifyPresigned.
		if k == "X-Amz-Signature" {
			continue
		}
		vs := values[k]
		sort.Strings(vs)
		for _, v := range vs {
			encodedKey := awsURLEncode(k, true)
			encodedVal := awsURLEncode(v, true)
			parts = append(parts, encodedKey+"="+encodedVal)
		}
	}

	return strings.Join(parts, "&")
}

func canonicalHeaderValue(v string) string {
	v = strings.TrimSpace(v)
	if v == "" {
		return ""
	}
	fields := strings.Fields(v)
	return strings.Join(fields, " ")
}

func buildCanonicalRequest(r *http.Request, signedHeaderNames []string, payloadHash string) string {
	canonicalURI := awsURLEncode(r.URL.EscapedPath(), false)
	canonicalQS := canonicalQueryString(r.URL)

	lowerNames := make([]string, len(signedHeaderNames))
	for i, h := range signedHeaderNames {
		lowerNames[i] = strings.ToLower(strings.TrimSpace(h))
	}

	var hdrBuilder strings.Builder
	for _, name := range lowerNames {
		if name == "" {
			continue
		}
		var value string
		if name == "host" {
			value = r.Host
			if value == "" {
				value = r.URL.Host
			}
		} else {
			value = r.Header.Get(name)
		}
		value = canonicalHeaderValue(value)
		hdrBuilder.WriteString(name)
		hdrBuilder.WriteString(":")
		hdrBuilder.WriteString(value)
		hdrBuilder.WriteString("\n")
	}
	canonicalHeaders := hdrBuilder.String()
	canonicalSignedHeaders := strings.Join(lowerNames, ";")

	var b strings.Builder
	b.WriteString(r.Method)
	b.WriteString("\n")
	b.WriteString(canonicalURI)
	b.WriteString("\n")
	b.WriteString(canonicalQS)
	b.WriteString("\n")
	b.WriteString(canonicalHeaders)
	b.WriteString("\n")
	b.WriteString(canonicalSignedHeaders)
	b.WriteString("\n")
	b.WriteString(payloadHash)

	return b.String()
}

func hmacSHA256(key []byte, data string) []byte {
	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))
	return h.Sum(nil)
}

// computeSignature returns the SigV4 signature over a canonical request; the
// header-signed and presigned paths share it.
func (e *AuthEngine) computeSignature(amzDate, dateStamp, region, service, canonicalReq string) []byte {
	crHash := sha256.Sum256([]byte(canonicalReq))
	crHashHex := hex.EncodeToString(crHash[:])

	credentialScope := strings.Join([]string{dateStamp, region, service, "aws4_request"}, "/")
	var stsBuilder strings.Builder
	stsBuilder.WriteString("AWS4-HMAC-SHA256\n")
	stsBuilder.WriteString(amzDate)
	stsBuilder.WriteString("\n")
	stsBuilder.WriteString(credentialScope)
	stsBuilder.WriteString("\n")
	stsBuilder.WriteString(crHashHex)
	stringToSign := stsBuilder.String()

	kSecret := []byte("AWS4" + e.SecretKey)
	kDate := hmacSHA256(kSecret, dateStamp)
	kRegion := hmacSHA256(kDate, region)
	kService := hmacSHA256(kRegion, service)
	kSigning := hmacSHA256(kService, "aws4_request")
	return hmacSHA256(kSigning, stringToSign)
}

func (e *AuthEngine) VerifyRequest(r *http.Request) error {
	if r.URL.Query().Get("X-Amz-Signature") != "" {
		return e.verifyPresigned(r)
	}

	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, AWSv4Prefix) {
		return errors.New("missing or invalid Authorization header")
	}

	params := strings.TrimSpace(strings.TrimPrefix(auth, AWSv4Prefix))
	parts := strings.Split(params, ",")
	kv := make(map[string]string, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx := strings.IndexByte(p, '=')
		if idx <= 0 {
			continue
		}
		k := p[:idx]
		v := p[idx+1:]
		kv[k] = strings.TrimSpace(v)
	}

	credStr, okCred := kv["Credential"]
	signedHeadersStr, okSigned := kv["SignedHeaders"]
	signatureHex, okSig := kv["Signature"]
	if !okCred || !okSigned || !okSig {
		return errors.New("missing required Authorization parameters")
	}

	credParts := strings.Split(credStr, "/")
	if len(credParts) != 5 {
		return errors.New("invalid Credential format in Authorization header")
	}
	accessKeyID := credParts[0]
	dateStamp := credParts[1]
	region := credParts[2]
	service := credParts[3]
	term := credParts[4]

	if term != "aws4_request" {
		return errors.New("invalid Credential termination string in Authorization header")
	}
	if accessKeyID != e.AccessKey {
		return errors.New("invalid Access Key ID in Authorization header")
	}
	if region == "" || service == "" {
		return errors.New("missing region or service in Credential")
	}

	amzDate := r.Header.Get("X-Amz-Date")
	if amzDate == "" {
		return errors.New("missing X-Amz-Date header")
	}

	// ±15-minute skew window - without it a captured request replays forever.
	t, err := time.Parse("20060102T150405Z", amzDate)
	if err != nil {
		return errors.New("invalid X-Amz-Date format")
	}
	if d := time.Since(t); d < -15*time.Minute || d > 15*time.Minute {
		return errors.New("request timestamp outside acceptable window")
	}

	payloadHash := r.Header.Get("X-Amz-Content-Sha256")
	if payloadHash == "" {
		return errors.New("missing X-Amz-Content-Sha256 header")
	}

	signedHeaderNames := strings.Split(signedHeadersStr, ";")
	canonicalReq := buildCanonicalRequest(r, signedHeaderNames, payloadHash)
	computedSignature := e.computeSignature(amzDate, dateStamp, region, service, canonicalReq)

	decodedSignature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return fmt.Errorf("error decoding signature string: %w", err)
	}

	if !hmac.Equal(computedSignature, decodedSignature) {
		return errors.New("HMAC signature mismatch")
	}

	return nil
}

// verifyPresigned validates a query-string SigV4 signature. The signed payload
// hash is always the literal UNSIGNED-PAYLOAD, never read from the request.
func (e *AuthEngine) verifyPresigned(r *http.Request) error {
	q := r.URL.Query()

	if q.Get("X-Amz-Algorithm") != "AWS4-HMAC-SHA256" {
		return errors.New("invalid or missing X-Amz-Algorithm")
	}

	credParts := strings.Split(q.Get("X-Amz-Credential"), "/")
	if len(credParts) != 5 {
		return errors.New("invalid X-Amz-Credential format")
	}
	accessKeyID := credParts[0]
	dateStamp := credParts[1]
	region := credParts[2]
	service := credParts[3]
	term := credParts[4]

	if term != "aws4_request" {
		return errors.New("invalid X-Amz-Credential termination string")
	}
	if accessKeyID != e.AccessKey {
		return errors.New("invalid Access Key ID in X-Amz-Credential")
	}
	if region == "" || service == "" {
		return errors.New("missing region or service in X-Amz-Credential")
	}

	amzDate := q.Get("X-Amz-Date")
	if amzDate == "" {
		return errors.New("missing X-Amz-Date")
	}
	t, err := time.Parse("20060102T150405Z", amzDate)
	if err != nil {
		return errors.New("invalid X-Amz-Date format")
	}

	expires, err := strconv.Atoi(q.Get("X-Amz-Expires"))
	if err != nil || expires <= 0 {
		return errors.New("invalid X-Amz-Expires")
	}
	const maxExpiresSeconds = 7 * 24 * 60 * 60 // S3 caps presigned URLs at 7 days.
	if expires > maxExpiresSeconds {
		return errors.New("X-Amz-Expires exceeds maximum of 7 days")
	}

	// X-Amz-Expires bounds the replay window here; only the forward half of the
	// ±15-minute skew check still applies.
	now := time.Now()
	if now.After(t.Add(time.Duration(expires) * time.Second)) {
		return errors.New("presigned URL has expired")
	}
	if t.After(now.Add(15 * time.Minute)) {
		return errors.New("presigned URL timestamp is in the future")
	}

	signedHeadersStr := q.Get("X-Amz-SignedHeaders")
	if signedHeadersStr == "" {
		return errors.New("missing X-Amz-SignedHeaders")
	}
	signedHeaderNames := strings.Split(signedHeadersStr, ";")

	decodedSignature, err := hex.DecodeString(q.Get("X-Amz-Signature"))
	if err != nil {
		return fmt.Errorf("error decoding signature string: %w", err)
	}

	canonicalReq := buildCanonicalRequest(r, signedHeaderNames, "UNSIGNED-PAYLOAD")
	computedSignature := e.computeSignature(amzDate, dateStamp, region, service, canonicalReq)

	if !hmac.Equal(computedSignature, decodedSignature) {
		return errors.New("HMAC signature mismatch")
	}

	return nil
}
