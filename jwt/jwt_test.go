package jwt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

var (
	signingKey = []byte("asecretsignstring")
	data       = jwt.MapClaims{
		"id":   "id",
		"name": "name",
	}
	token           = jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	signedString, _ = token.SignedString(signingKey)
)

func TestDefaultJWT_HandleJWT(t *testing.T) {
	j := New(Config{SigningKey: signingKey})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set(HeaderAuthorization, fmt.Sprintf("%s %s", Bearer, signedString))

	assert.NoError(t, j.HandleJWT(w, r))
	value := r.Context().Value("user")
	assert.NotNil(t, value)
	m, ok := value.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, m["id"], "id")
	assert.Equal(t, m["name"], "name")
}

func TestJWT_HandleJWT_FromCookie(t *testing.T) {
	j := New(Config{
		SigningKey: signingKey,
		Extractor:  FromCookie("user")})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	cookie := &http.Cookie{Name: "user", Value: signedString}
	r.AddCookie(cookie)

	assert.NoError(t, j.HandleJWT(w, r))
	value := r.Context().Value("user")
	assert.NotNil(t, value)
	m, ok := value.(jwt.MapClaims)
	assert.True(t, ok)
	assert.Equal(t, m["id"], "id")
	assert.Equal(t, m["name"], "name")
}

type CustomClaims struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (CustomClaims) Valid() error {
	return nil
}

func TestJWT_HandleJWT_CustomClaims(t *testing.T) {
	claims := CustomClaims{
		ID:   "id",
		Name: "name",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(signingKey)

	j := New(Config{
		SigningKey: signingKey,
		Claims:     &CustomClaims{},
	})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	r.Header.Set(HeaderAuthorization, fmt.Sprintf("%s %s", Bearer, tokenString))

	assert.NoError(t, j.HandleJWT(w, r))
	value := r.Context().Value("user")
	assert.NotNil(t, value)
	m, ok := value.(*CustomClaims)
	assert.True(t, ok)
	assert.Equal(t, m.ID, "id")
	assert.Equal(t, m.Name, "name")
}
