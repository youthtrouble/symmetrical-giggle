package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/youthtrouble/symmetrical-giggle/internal/models"
	"github.com/youthtrouble/symmetrical-giggle/internal/repository"
	"github.com/youthtrouble/symmetrical-giggle/internal/services"
	"github.com/youthtrouble/symmetrical-giggle/pkg/logger"
)

type Handlers struct {
	repo           repository.Repository
	pollingManager *services.PollingManager
	logger         *logger.Logger
}

func NewHandlers(repo repository.Repository, pollingManager *services.PollingManager, logger *logger.Logger) *Handlers {
	return &Handlers{
		repo:           repo,
		pollingManager: pollingManager,
		logger:         logger,
	}
}

func (h *Handlers) GetReviews(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id is required"})
		return
	}

	hours := 48 // default
	if h := c.Query("hours"); h != "" {
		if parsed, err := strconv.Atoi(h); err == nil && parsed > 0 {
			hours = parsed
		}
	}

	limit := 100 // default
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	reviews, err := h.repo.GetReviews(appID, hours, limit)
	if err != nil {
		h.logger.Error("Failed to get reviews", "app_id", appID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reviews": reviews,
		"meta": gin.H{
			"app_id": appID,
			"hours":  hours,
			"count":  len(reviews),
		},
	})
}

func (h *Handlers) ConfigureApp(c *gin.Context) {
	appID := c.Param("appId")
	if appID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "app_id is required"})
		return
	}

	var req struct {
		PollInterval string `json:"poll_interval"`
		IsActive     *bool  `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}


	interval := 5 * time.Minute // default
	if req.PollInterval != "" {
		if parsed, err := time.ParseDuration(req.PollInterval); err == nil {
			interval = parsed
		} else {
			h.logger.Error("Failed to parse poll interval", "poll_interval", req.PollInterval, "error", err)
		}
	}

	h.logger.Info("Parsed poll interval", "input", req.PollInterval, "parsed_nanoseconds", interval.Nanoseconds())

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	config := &models.AppConfig{
		AppID:        appID,
		PollInterval: interval,
		IsActive:     isActive,
	}

	if err := h.repo.UpsertAppConfig(config); err != nil {
		h.logger.Error("Failed to save app config", "app_id", appID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save configuration"})
		return
	}

	if isActive {
		h.pollingManager.StartPolling(appID, interval)
	} else {
		h.pollingManager.StopPolling(appID)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Configuration updated successfully",
		"config":  config,
	})
}

func (h *Handlers) GetPollingStatus(c *gin.Context) {
	status := h.pollingManager.GetPollingStatus()
	c.JSON(http.StatusOK, gin.H{"polling_status": status})
}

func (h *Handlers) ServeIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "iOS App Store Reviews Viewer",
	})
}

func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now().UTC(),
	})
}