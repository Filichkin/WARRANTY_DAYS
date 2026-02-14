// Package router configures HTTP routes for the API.
package router

import (
	"log/slog"
	"net/http"

	"warranty_days/internal/httpapi/handler"
	"warranty_days/internal/httpapi/middleware"
)

func NewMux(claimsHandler *handler.ClaimsHandler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", claimsHandler.Health)
	mux.Handle("/claims", method(http.MethodGet, http.HandlerFunc(claimsHandler.GetClaimsByVIN)))
	mux.Handle("/claims/warranty-year", method(http.MethodGet, http.HandlerFunc(claimsHandler.GetWarrantyYearClaims)))

	return middleware.RequestLogging(logger, mux)
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
