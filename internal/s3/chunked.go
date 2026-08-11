package s3

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type ChunkedReader struct {
	reader    *bufio.Reader
	remaining int64
	done      bool
}

func NewChunkedReader(r io.Reader) *ChunkedReader {
	return &ChunkedReader{
		reader: bufio.NewReader(r),
	}
}

func (cr *ChunkedReader) Read(p []byte) (n int, err error) {
	if cr.done {
		return 0, io.EOF
	}

	if cr.remaining == 0 {
		if err := cr.readChunkHeader(); err != nil {
			return 0, err
		}
		if cr.done {
			return 0, io.EOF
		}
	}

	toRead := int64(len(p))
	if toRead > cr.remaining {
		toRead = cr.remaining
	}

	n, err = cr.reader.Read(p[:toRead])
	cr.remaining -= int64(n)

	return n, err
}

func (cr *ChunkedReader) readChunkHeader() error {
	line, err := cr.reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read chunk header: %w", err)
	}

	line = strings.TrimSuffix(line, "\r\n")
	line = strings.TrimSuffix(line, "\n")

	if line == "" {
		line, err = cr.reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read chunk header: %w", err)
		}
		line = strings.TrimSuffix(line, "\r\n")
		line = strings.TrimSuffix(line, "\n")
	}

	parts := strings.SplitN(line, ";", 2)
	if len(parts) == 0 {
		return fmt.Errorf("malformed chunk header: %s", line)
	}

	sizeStr := strings.TrimSpace(parts[0])
	if sizeStr == "" {
		return fmt.Errorf("empty chunk size")
	}

	size, err := strconv.ParseInt(sizeStr, 16, 64)
	if err != nil {
		return fmt.Errorf("invalid chunk size %q: %w", sizeStr, err)
	}

	if size == 0 {
		cr.done = true
		cr.readTrailer()
		return nil
	}

	cr.remaining = size
	return nil
}

func (cr *ChunkedReader) readTrailer() {
	for {
		line, err := cr.reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSuffix(line, "\r\n")
		line = strings.TrimSuffix(line, "\n")
		if line == "" {
			return
		}
	}
}

func IsChunked(r *http.Request) bool {
	contentEncoding := r.Header.Get("Content-Encoding")
	if contentEncoding == "aws-chunked" {
		return true
	}

	contentSha256 := r.Header.Get("x-amz-content-sha256")
	if strings.HasPrefix(contentSha256, "STREAMING") {
		return true
	}

	return false
}

func DecodeChunked(r *http.Request) io.Reader {
	if !IsChunked(r) {
		return r.Body
	}
	return NewChunkedReader(r.Body)
}
