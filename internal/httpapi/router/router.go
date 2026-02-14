// Package router configures HTTP routes for the API.
package router

import (
	"net/http"

	"warranty_days/internal/httpapi/handler"
)

func NewMux(claimsHandler *handler.ClaimsHandler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", claimsHandler.Health)
	mux.Handle("/claims", method(http.MethodGet, http.HandlerFunc(claimsHandler.GetClaimsByVIN)))
	mux.Handle("/claims/warranty-year", method(http.MethodGet, http.HandlerFunc(claimsHandler.GetWarrantyYearClaims)))

	return mux
}

func method(allowed string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowed {
			w.Header().Set("Allow", allowed)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}
