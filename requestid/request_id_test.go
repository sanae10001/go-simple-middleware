package requestid

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	format = "request id is %s"
)

func TestRequestId_ServeHTTP(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, format, r.Header.Get(HeaderRequestId))
	}
	h := NewRequestId()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set(HeaderRequestId, "random-request-id")
	h.ServeHTTP(w, r, next)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d.", http.StatusOK, w.Code)
	}
	expectedResp := fmt.Sprintf(format, "random-request-id")
	if w.Body.String() != expectedResp {
		t.Fatalf("Invalid responseStr. Expected [%s], got [%s]", expectedResp, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/", nil)
	h.ServeHTTP(w, r, next)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d.", http.StatusOK, w.Code)
	}
	var id string
	_, err := fmt.Fscanf(w.Body, format, &id)
	if err != nil {
		t.Error(err)
	}
	orgId, err := hex.DecodeString(id)
	if err != nil {
		t.Error(err)
	}
	if len(orgId) != 16 {
		t.Fatalf("Expected request id length is %d, got %d.", 16, len(orgId))
	}
}
