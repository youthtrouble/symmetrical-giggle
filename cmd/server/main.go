package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/youthtrouble/symmetrical-giggle/internal/api"
	"github.com/youthtrouble/symmetrical-giggle/internal/config"
	"github.com/youthtrouble/symmetrical-giggle/internal/repository"
	"github.com/youthtrouble/symmetrical-giggle/internal/services"
	"github.com/youthtrouble/symmetrical-giggle/pkg/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	logger := logger.New(cfg.LogLevel)

	repo, err := repository.NewSQLiteRepository(cfg.Database.Path)
	if err != nil {
		logger.Fatal("Failed to initialize repository:", err)
	}
	defer repo.Close()

	rssService := services.NewRSSService(logger)
	pollingManager := services.NewPollingManager(repo, rssService, logger)

	pollingManager.StartAll()
	defer pollingManager.StopAll()

	router := setupRouter(repo, pollingManager, logger)
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		logger.Info("Starting server on port " + cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown:", err)
	}
	logger.Info("Server exited")
}

func setupRouter(repo repository.Repository, pollingManager *services.PollingManager, logger *logger.Logger) *gin.Engine {
	router := gin.Default()

	handlers := api.NewHandlers(repo, pollingManager, logger)

	api.SetupRoutes(router, handlers)

	return router
}
