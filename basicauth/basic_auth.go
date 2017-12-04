package basicauth

import (
	"context"
	"net/http"
	"strconv"
)

const (
	basic        = "Basic"
	defaultRealm = "Restricted"
)

const (
	HeaderAuthorization   = "Authorization"
	HeaderWWWAuthenticate = "WWW-Authenticate"
)

type (
	BasicAuth struct {
		// A description of the protected area. Default value is "Restricted".
		Realm string

		// A function to validate BasicAuth credentials(like username:password).
		Validator BasicAuthValidator
	}

	BasicAuthValidator func(string, string, context.Context) bool
)

func DefaultValidator(username, password string) BasicAuthValidator {
	return func(u string, p string, c context.Context) bool {
		if username == u && password == p {
			return true
		}
		return false
	}
}

func NewBasicAuth(fn BasicAuthValidator) *BasicAuth {
	b := new(BasicAuth)
	b.Validator = fn
	return b
}

func (ba *BasicAuth) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// Authorization: Basic base64-username-and-password
	username, password, ok := r.BasicAuth()
	if ok && ba.Validator(username, password, r.Context()) {
		next(w, r)
		return
	}

	realm := defaultRealm
	if ba.Realm != "" {
		realm = strconv.Quote(ba.Realm)
	}
	w.Header().Set(HeaderWWWAuthenticate, basic+" realm="+realm)
	w.WriteHeader(http.StatusUnauthorized)
}
