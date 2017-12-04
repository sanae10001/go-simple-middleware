package bodylimit

import (
	"errors"
	"io"
	"net/http"
	"strconv"
)

var (
	ErrTooLarge         = errors.New("request body too large")
	internalErrTooLarge = errors.New("http: request body too large")
)

func NewBodyLimit(limit Byte) *BodyLimit {
	return &BodyLimit{
		limit: int64(limit),
	}
}

type BodyLimit struct {
	limit int64
}

func (bs *BodyLimit) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	contentLength, _ := strconv.ParseInt(
		r.Header.Get("Content-Length"), 10, 64)
	// Check content length
	if contentLength > bs.limit {
		http.Error(w, ErrTooLarge.Error(), http.StatusRequestEntityTooLarge)
		return
	}

	// Check body size
	r.Body = &maxBytesReader{http.MaxBytesReader(w, r.Body, bs.limit)}
	next(w, r)
}

type maxBytesReader struct {
	r io.ReadCloser
}

func (l *maxBytesReader) Read(p []byte) (n int, err error) {
	n, err = l.r.Read(p)
	if err != nil && err.Error() == internalErrTooLarge.Error() {
		return n, ErrTooLarge
	}
	return
}

func (l *maxBytesReader) Close() error {
	return l.r.Close()
}
