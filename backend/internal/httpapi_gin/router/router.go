package router

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"warranty_days/internal/httpapi/handler"
	"warranty_days/internal/httpapi/middleware"
	ginmiddleware "warranty_days/internal/httpapi_gin/middleware"
)

func NewEngine(
	claimsHandler *handler.ClaimsHandler,
	authHandler *handler.AuthHandler,
	jwtSvc middleware.AccessTokenValidator,
	logger *slog.Logger,
) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(ginmiddleware.RequestLogging(logger))

	// public routes
	engine.GET("/health", gin.WrapF(claimsHandler.Health))
	engine.POST("/auth/register", gin.WrapF(authHandler.Register))
	engine.POST("/auth/login", gin.WrapF(authHandler.Login))
	engine.POST("/auth/refresh", gin.WrapF(authHandler.Refresh))

	// protected routes
	engine.GET("/claims", gin.WrapH(middleware.Auth(jwtSvc, http.HandlerFunc(claimsHandler.GetClaimsByVIN))))
	engine.GET(
		"/claims/warranty-year",
		gin.WrapH(middleware.Auth(jwtSvc, http.HandlerFunc(claimsHandler.GetWarrantyYearClaims))),
	)

	return engine
}
