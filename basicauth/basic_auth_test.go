package basicauth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	username, password = "username", "password"
	responseStr        = "Continue"
)

func TestBasicAuth_ServeHTTP(t *testing.T) {
	next := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(responseStr))
	}
	h := NewBasicAuth(DefaultValidator(username, password))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	r.SetBasicAuth(username, password)
	h.HandlerWithNext(w, r, next)
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code %d, got %d.", http.StatusOK, w.Code)
	}
	if w.Body.String() != responseStr {
		t.Fatalf("Invalid responseStr. Expected [%s], got [%s]", responseStr, w.Body.String())
	}

	w = httptest.NewRecorder()
	r = httptest.NewRequest("POST", "/", nil)
	r.SetBasicAuth("invalid_username", "invalid_password")
	h.HandlerWithNext(w, r, next)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status code %d, got %d.", http.StatusUnauthorized, w.Code)
	}
	if w.Header().Get(HeaderWWWAuthenticate) == "" {
		t.Fatalf("Have no resposne header: [%s]", HeaderWWWAuthenticate)
	}

}
