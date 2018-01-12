package jwt

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

const (
	HeaderAuthorization = "Authorization"
	Bearer              = "bearer"
)

var (
	ErrorAuthHeaderFormat = errors.New("authorization header format must be Bearer {token}")
	ErrorTokenMissing     = errors.New("authorization token is required, but not found")
	ErrorExtractingToken  = errors.New("error extracting token")
	ErrorParsingToken     = errors.New("error parsing jwt token")
	ErrorInvalidToken     = errors.New("token is invalid")
)

type (
	// A function called when an error is encountered
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

	// A function called before set value into context
	ContextValueFunc func(token *jwt.Token) (interface{}, error)

	// A function used to extract jwt token raw string from request.
	TokenExtractor func(r *http.Request) (string, error)

	PrintLogger interface {
		Println(v ...interface{})
	}

	Config struct {
		// Signing key to validate token.
		// Required if ValidationKeyGetter is nil.
		SigningKey interface{}

		// The function that will return the Key to validate the JWT.
		// It can be either a shared secret or a public key.
		// Required if SigningKey is nil.
		// First use this.
		ValidationKeyGetter jwt.Keyfunc

		// Signing method, used to verify tokens are signed with the specific signing algorithm.
		// Optional. Default: jwt.SigningMethodHS256.
		SigningMethod jwt.SigningMethod

		// The key name in the context where the user information
		// from the JWT will be stored.
		// Optional. Default: "user"
		ContextKey string

		// Claims are extendable claims data defining token content.
		// Optional. Default: jwt.MapClaims
		Claims jwt.Claims

		// Debug flag turns on debugging output
		// Optional. Default: false
		Debug bool

		// When set, all requests with the OPTIONS method will use authentication
		// Optional. Default: false
		EnableAuthOnOptions bool

		// A boolean indicating if the credentials are optional or not
		// Optional. Default: false
		CredentialsOptional bool

		// A function that extracts the token from the request
		// Optional. Default: FromAuthHeader (i.e., from Authorization header as bearer token)
		Extractor TokenExtractor

		// The function that will be called when there's an error validating the token
		// Optional. Default: OnError
		ErrorHandler ErrorHandler

		// The function that will be called before set value into context.
		// Used to customize the value will be stored into context.
		// Optional. Default: ContextValueSetClaims
		ContextValueFunc ContextValueFunc
	}

	JWT struct {
		cfg    Config
		logger PrintLogger
	}
)

func New(config Config) *JWT {
	if config.SigningMethod == nil {
		config.SigningMethod = jwt.SigningMethodHS256
	}

	if config.ValidationKeyGetter == nil {
		if config.SigningKey == nil {
			panic("JWT: SigningKey or ValidationKeyGetter is required")
		}
		config.ValidationKeyGetter = func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != config.SigningMethod.Alg() {
				return nil, fmt.Errorf("unexpected jwt signing method=%s", t.Method.Alg())
			}
			return config.SigningKey, nil
		}
	}

	if config.ContextKey == "" {
		config.ContextKey = "user"
	}

	if config.Claims == nil {
		config.Claims = jwt.MapClaims{}
	}

	if config.Extractor == nil {
		config.Extractor = FromHeader
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = OnError
	}

	if config.ContextValueFunc == nil {
		config.ContextValueFunc = ContextValueSetClaims
	}

	return &JWT{cfg: config}
}

func (j *JWT) log(s string) {
	if j.cfg.Debug {
		if j.logger == nil {
			j.logger = log.New(os.Stderr, "", log.LstdFlags)
		}
		j.logger.Println(s)
	}
}

func (j *JWT) SetLogger(l PrintLogger) {
	j.logger = l
}

func (j *JWT) HandlerWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	err := j.HandleJWT(w, r)
	if err != nil {
		j.cfg.ErrorHandler(w, r, err)
		return
	}
	if next != nil {
		next(w, r)
	}
}

func (j *JWT) Handler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := j.HandleJWT(w, r)
		if err != nil {
			j.cfg.ErrorHandler(w, r, err)
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (j *JWT) HandleJWT(w http.ResponseWriter, r *http.Request) error {
	if !j.cfg.EnableAuthOnOptions {
		if r.Method == "OPTIONS" {
			return nil
		}
	}

	// Use the specified rawToken extractor to extract a rawToken from the request
	rawToken, err := j.cfg.Extractor(r)

	if err != nil {
		j.log(fmt.Sprintf("error extracting token: %v", err))
		return ErrorExtractingToken
	}
	j.log(fmt.Sprintf("token extracted: %s", rawToken))

	// Check required if rawToken is empty
	if rawToken == "" {
		if j.cfg.CredentialsOptional {
			j.log("no credentials found (CredentialsOptional=true)")
			return nil
		}

		j.log(ErrorTokenMissing.Error())
		return ErrorTokenMissing
	}

	token, err := jwt.ParseWithClaims(rawToken, j.cfg.Claims, j.cfg.ValidationKeyGetter)
	if err != nil {
		j.log(fmt.Sprintf("error parsing jwt token: %v", err))
		return ErrorParsingToken
	} else if !token.Valid {
		j.log(fmt.Sprintf("token is invalid: %v", token))
		return ErrorInvalidToken
	}

	j.log(fmt.Sprintf("JWT: %v", token))

	value, err := j.cfg.ContextValueFunc(token)
	if err != nil {
		return err
	}
	newRequest := r.WithContext(context.WithValue(r.Context(), j.cfg.ContextKey, value))
	*r = *newRequest
	return nil
}

// FromHeader is a "TokenExtractor" that takes a give request and extracts
// the JWT token from the Authorization header.
func FromHeader(r *http.Request) (string, error) {
	auth := r.Header.Get(HeaderAuthorization)
	if auth == "" {
		return "", nil
	}

	authHeaderParts := strings.Split(auth, " ")
	if len(authHeaderParts) != 2 || strings.ToLower(authHeaderParts[0]) != Bearer {
		return "", ErrorAuthHeaderFormat
	}

	return authHeaderParts[1], nil
}

// FromQuery returns a function that extracts the token from the specified
// query string parameter.
func FromQuery(key string) TokenExtractor {
	return func(r *http.Request) (string, error) {
		return r.URL.Query().Get(key), nil
	}
}

// FromCookie returns a function that extracts the token from the first cookie
// matched a special key.
func FromCookie(key string) TokenExtractor {
	return func(r *http.Request) (string, error) {
		cookie, err := r.Cookie(key)
		if err != nil {
			if err == http.ErrNoCookie {
				// Ignore if no cookie
				return "", nil
			} else {
				return "", err
			}
		}
		return cookie.Value, nil
	}
}

// FromFirst returns a function that runs multiple token extractors and takes the
// first token it finds.
func FromFirst(extractors ...TokenExtractor) TokenExtractor {
	return func(r *http.Request) (string, error) {
		for _, ex := range extractors {
			token, err := ex(r)
			if err != nil {
				return "", err
			}
			if token != "" {
				return token, nil
			}
		}
		return "", nil
	}
}

func ContextValueSetClaims(token *jwt.Token) (interface{}, error) {
	return token.Claims, nil
}

func ContextValueSetToken(token *jwt.Token) (interface{}, error) {
	return token, nil
}

func OnError(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusUnauthorized)
}
