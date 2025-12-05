package main

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

func (f HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f(w, r)
}

func Logging(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("[LOG] %s %s\n", r.Method, r.URL.Path)
		next(w, r)
	}
}

func Auth(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func RateLimit(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("[RateLimit] OK")
		next(w, r)
	}
}

func Use(h HandlerFunc, m ...func(HandlerFunc) HandlerFunc) HandlerFunc {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

func Hello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello with middleware!")
}

func main() {
	base := HandlerFunc(Hello)

	h := Use(base,
		Logging,
		Auth,
		RateLimit,
	)
	fmt.Println("Starting server at :8080")
	http.ListenAndServe(":8080", h)
}
