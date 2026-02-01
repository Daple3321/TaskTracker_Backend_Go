package middleware

import (
	"log"
	"net/http"
)

func Logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[%s : %s] [%s]", r.Method, r.RequestURI, r.RemoteAddr)

		next(w, r)
	}
}
