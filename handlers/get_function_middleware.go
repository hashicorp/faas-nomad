package handlers

import (
	"context"
	"fmt"
	"net/http"
)

// FunctionNameCTXKey is a context key which points to the location of the function
// name set bu the ExtractFunction middleware
var FunctionNameCTXKey = struct{}{}

// MakeExtractFunctionMiddleWare returns a middleware handler which validates
// the presence of the function name in the URI query.
func MakeExtractFunctionMiddleWare(
	getVars func(*http.Request) map[string]string,
	next http.HandlerFunc) http.HandlerFunc {

	return func(rw http.ResponseWriter, r *http.Request) {
		vars := getVars(r)
		functionName := vars["name"]

		if functionName == "" {
			rw.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(rw, fmt.Errorf("No function name"))
			return
		}

		r = r.WithContext(context.WithValue(r.Context(), FunctionNameCTXKey, functionName))
		next(rw, r)
	}
}
