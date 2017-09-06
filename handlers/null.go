package handlers

import "net/http"

func MakeNull() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
}
