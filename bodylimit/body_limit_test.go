package bodylimit

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testBody = []byte("testmessage") // 11 byte
)

func TestBodyLimit_ServeHTTP(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		d, err := ioutil.ReadAll(r.Body)
		if err != nil {
			if err == ErrTooLarge {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
			}
			w.Write([]byte(err.Error()))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(d)
		}
	}

	// No limit
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", bytes.NewReader(testBody))
	b := NewBodyLimit(11 * B)
	b.ServeHTTP(w, r, next)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d.", http.StatusOK, w.Code)
	}
	if !bytes.Equal(w.Body.Bytes(), testBody) {
		t.Fatalf(
			"Invalid response. Expected [%s], got [%s]",
			string(testBody), w.Body.String(),
		)
	}

	// Limited
	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/", bytes.NewReader(testBody))
	b = NewBodyLimit(10 * B)
	b.ServeHTTP(w, r, next)
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf(
			"Expected status code %d, got %d.",
			http.StatusRequestEntityTooLarge, w.Code,
		)
	}
	if !bytes.Equal(w.Body.Bytes(), []byte(ErrTooLarge.Error())) {
		t.Fatalf(
			"Invalid response. Expected [%s], got [%s]",
			ErrTooLarge.Error(), w.Body.String(),
		)
	}
}
