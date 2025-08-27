package services

import (
	"context"
	"sync"
	"time"

	"github.com/youthtrouble/symmetrical-giggle/internal/models"
	"github.com/youthtrouble/symmetrical-giggle/internal/repository"
	"github.com/youthtrouble/symmetrical-giggle/pkg/logger"
)

type PollingManager struct {
	repo       repository.Repository
	rssService *RSSService
	logger     *logger.Logger
	pollers    map[string]*AppPoller
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

type AppPoller struct {
	appID    string
	interval time.Duration
	ticker   *time.Ticker
	done     chan struct{}
}

func NewPollingManager(repo repository.Repository, rssService *RSSService, logger *logger.Logger) *PollingManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &PollingManager{
		repo:       repo,
		rssService: rssService,
		logger:     logger,
		pollers:    make(map[string]*AppPoller),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (pm *PollingManager) StartAll() error {
	activeApps, err := pm.repo.GetActiveApps()
	if err != nil {
		pm.logger.Error("Failed to get active apps", "error", err)
		return err
	}

	pm.logger.Info("Starting polling for active apps", "count", len(activeApps))

	// If no active apps, just return without error
	if len(activeApps) == 0 {
		pm.logger.Info("No active apps found, skipping polling startup")
		return nil
	}

	for _, appID := range activeApps {
		pm.logger.Info("Processing app for polling", "app_id", appID)

		config, err := pm.repo.GetAppConfig(appID)
		if err != nil {
			pm.logger.Error("Failed to get app config", "app_id", appID, "error", err)
			continue
		}

		if config != nil && config.IsActive && config.PollInterval > 0 {
			pm.logger.Info("Starting polling for app", "app_id", appID, "interval", config.PollInterval)
			pm.StartPolling(appID, config.PollInterval)
		} else if config != nil && config.IsActive && config.PollInterval <= 0 {
			pm.logger.Warn("Skipping app with invalid polling interval", "app_id", appID, "interval", config.PollInterval)
		} else if config == nil {
			pm.logger.Warn("No config found for app", "app_id", appID)
		} else if !config.IsActive {
			pm.logger.Info("App is not active, skipping", "app_id", appID)
		}
	}

	pm.logger.Info("Finished starting polling for all apps")
	return nil
}

func (pm *PollingManager) StartPolling(appID string, interval time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if poller, exists := pm.pollers[appID]; exists {
		pm.stopPoller(poller)
	}

	if interval <= 0 {
		pm.logger.Error("Invalid polling interval", "app_id", appID, "interval", interval)
		return
	}

	poller := &AppPoller{
		appID:    appID,
		interval: interval,
		ticker:   time.NewTicker(interval),
		done:     make(chan struct{}),
	}

	pm.pollers[appID] = poller

	go pm.pollApp(poller)

	pm.logger.Info("Started polling", "app_id", appID, "interval", interval)
}

func (pm *PollingManager) StopPolling(appID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if poller, exists := pm.pollers[appID]; exists {
		pm.stopPoller(poller)
		delete(pm.pollers, appID)
		pm.logger.Info("Stopped polling", "app_id", appID)
	}
}

func (pm *PollingManager) StopAll() {
	pm.cancel()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	for appID, poller := range pm.pollers {
		pm.stopPoller(poller)
		delete(pm.pollers, appID)
	}
}

func (pm *PollingManager) stopPoller(poller *AppPoller) {
	poller.ticker.Stop()
	close(poller.done)
}

func (pm *PollingManager) pollApp(poller *AppPoller) {

	pm.fetchAndStore(poller.appID, poller.interval)

	for {
		select {
		case <-poller.ticker.C:
			pm.fetchAndStore(poller.appID, poller.interval)
		case <-poller.done:
			return
		case <-pm.ctx.Done():
			return
		}
	}
}

func (pm *PollingManager) fetchAndStore(appID string, interval time.Duration) {
	ctx, cancel := context.WithTimeout(pm.ctx, 2*time.Minute)
	defer cancel()

	reviews, err := pm.rssService.FetchWithRetry(ctx, appID, 3)
	if err != nil {
		pm.logger.Error("Failed to fetch reviews", "app_id", appID, "error", err)
		return
	}

	stored := 0
	for _, review := range reviews {
		// Check if review already exists
		exists, err := pm.repo.ReviewExists(review.ID)
		if err != nil {
			pm.logger.Error("Failed to check review existence", "review_id", review.ID, "error", err)
			continue
		}

		if !exists {
			if err := pm.repo.CreateReview(&review); err != nil {
				pm.logger.Error("Failed to store review", "review_id", review.ID, "error", err)
				continue
			}
			stored++
		}
	}

	// Update last poll time
	now := time.Now()
	config := &models.AppConfig{
		AppID:    appID,
		LastPoll: &now,
		PollInterval: interval,
		IsActive: true,
	}

	if err := pm.repo.UpsertAppConfig(config); err != nil {
		pm.logger.Error("Failed to update app config", "app_id", appID, "error", err)
	}

	pm.logger.Info("Polling completed", "app_id", appID, "fetched", len(reviews), "stored", stored)
}

func (pm *PollingManager) GetPollingStatus() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	status := make(map[string]interface{})
	for appID, poller := range pm.pollers {
		status[appID] = map[string]interface{}{
			"interval": poller.interval.String(),
			"active":   true,
		}
	}

	return status
}
