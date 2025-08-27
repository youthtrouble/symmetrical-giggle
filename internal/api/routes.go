package api

import (
	"github.com/gin-gonic/gin"
	"github.com/youthtrouble/symmetrical-giggle/internal/api/middleware"
)

func SetupRoutes(router *gin.Engine, handlers *Handlers) {

	router.Use(middleware.CORS())
	router.Use(middleware.RateLimiter())
	router.Use(middleware.RequestLogger())

	router.GET("/health", handlers.HealthCheck)

	router.GET("/", handlers.ServeIndex)

	api := router.Group("/api")
	{
		api.GET("/reviews/:appId", handlers.GetReviews)
		api.POST("/apps/:appId/configure", handlers.ConfigureApp)
		api.GET("/polling/status", handlers.GetPollingStatus)
	}
}
