# negroni-middlewares
Simple middleware for [negroni](https://github.com/urfave/negroni)

- [x] Basic auth
- [x] Body limit
- [x] Request id
- [x] JWT


#### Example

*jwt*:

``` go
func main() {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}
	j := jwt.New(jwt.Config{SigningKey: []byte("asecretsigningstring")})

	mux := http.NewServeMux()
	mux.Handle("/", j.Handler(http.HandlerFunc(h)))

	http.ListenAndServe(":8080", mux)
}
```

*jwt for negroni*

```go
func main() {
	h := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	}
	j := jwt.New(jwt.Config{SigningKey: []byte("asecretsigningstring")})

	mux := http.NewServeMux()
	mux.HandleFunc("/", h)
	
	n := negroni.New()
	n.UseFunc(j.HandlerWithNext)
	n.UseHandler(mux)

	http.ListenAndServe(":8080", n)
}
```