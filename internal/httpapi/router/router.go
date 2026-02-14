// Package router configures HTTP routes for the API.
package router

import (
	"log/slog"
	"net/http"

	"warranty_days/internal/httpapi/handler"
	"warranty_days/internal/httpapi/middleware"
)

func NewMux(
	claimsHandler *handler.ClaimsHandler,
	authHandler *handler.AuthHandler,
	jwtSvc middleware.AccessTokenValidator,
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()

	// public routes
	mux.HandleFunc("/health", claimsHandler.Health)
	mux.Handle("/auth/register", method(http.MethodPost, http.HandlerFunc(authHandler.Register)))
	mux.Handle("/auth/login", method(http.MethodPost, http.HandlerFunc(authHandler.Login)))
	mux.Handle("/auth/refresh", method(http.MethodPost, http.HandlerFunc(authHandler.Refresh)))

	// protected routes
	mux.Handle(
		"/claims",
		middleware.Auth(jwtSvc, method(http.MethodGet, http.HandlerFunc(claimsHandler.GetClaimsByVIN))),
	)
	mux.Handle(
		"/claims/warranty-year",
		middleware.Auth(jwtSvc, method(http.MethodGet, http.HandlerFunc(claimsHandler.GetWarrantyYearClaims))),
	)

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
