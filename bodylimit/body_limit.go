package bodylimit

import (
	"io"
	"net/http"
	"strconv"
)

const (
	internalErrTooLargeString = "http: request body too large"
)

func NewBodyLimit(limit Byte) *BodyLimit {
	return &BodyLimit{
		limit: int64(limit),
	}
}

type BodyLimit struct {
	limit int64
}

/*
The http.ResponseWriter must be net/http.*response.
Reason see: net/http.maxBytesReader.Read
*/
func (bs *BodyLimit) HandlerWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	contentLength, _ := strconv.ParseInt(
		r.Header.Get("Content-Length"), 10, 64)
	// Check content length
	if contentLength > bs.limit {
		http.Error(w, internalErrTooLargeString, http.StatusRequestEntityTooLarge)
		return
	}

	// Check body size
	r.Body = &maxBytesReader{w, http.MaxBytesReader(w, r.Body, bs.limit)}

	next(w, r)
}

func (bs *BodyLimit) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bs.HandlerWithNext(w, r, h.ServeHTTP)

	})
}

type maxBytesReader struct {
	w http.ResponseWriter
	r io.ReadCloser
}

func (l *maxBytesReader) Read(p []byte) (n int, err error) {
	n, err = l.r.Read(p)
	if err != nil && err.Error() == internalErrTooLargeString {
		l.w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}
	return
}

func (l *maxBytesReader) Close() error {
	return l.r.Close()
}
