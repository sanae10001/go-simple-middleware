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

func (h *RequestId) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	requestId := r.Header.Get(h.header)
	if requestId == "" {
		var err error
		requestId, err = h.idGenFunc(h.length)
		if err != nil {
			goto Next
		}
	}
	r.Header.Set(h.header, requestId)
	w.Header().Set(h.header, requestId)

Next:
	next(w, r)

}

func DefaultIdGenFunc(length int) (string, error) {
	id := make([]byte, length)
	_, err := rand.Read(id)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(id), nil
}
