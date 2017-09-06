package handlers

import (
	"log"
	"net/http"
)

// MakeReader implements the OpenFaaS reader handler
func MakeReader() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received request: " + r.URL.RawPath)
	}
}
