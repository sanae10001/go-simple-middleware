package requestid

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

const (
	HeaderRequestId = "X-Request-Id"
	defaultLength   = 16
)

func NewRequestId(opts ...Option) *RequestId {
	r := &RequestId{
		header:    HeaderRequestId,
		length:    defaultLength,
		idGenFunc: DefaultIdGenFunc,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

type (
	IdGenFunc func(int) (string, error)

	Option func(*RequestId)

	RequestId struct {
		length    int
		idGenFunc IdGenFunc
		header    string
	}
)

func OptIdGenFunc(f IdGenFunc) Option {
	return func(id *RequestId) {
		id.idGenFunc = f
	}
}

func OptIdLength(l int) Option {
	return func(id *RequestId) {
		id.length = l
	}
}

func OptHeader(h string) Option {
	return func(id *RequestId) {
		id.header = h
	}
}

func (ri *RequestId) handle(w http.ResponseWriter, r *http.Request) {
	requestId := r.Header.Get(ri.header)
	if requestId == "" {
		var err error
		requestId, err = ri.idGenFunc(ri.length)
		if err != nil {
			return
		}
	}
	r.Header.Set(ri.header, requestId)
	w.Header().Set(ri.header, requestId)
}

func (ri *RequestId) HandlerWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	ri.handle(w, r)
	if next != nil {
		next(w, r)
	}
}

func (ri *RequestId) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ri.HandlerWithNext(w, r, h.ServeHTTP)
	})
}

func DefaultIdGenFunc(length int) (string, error) {
	id := make([]byte, length)
	_, err := rand.Read(id)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}
