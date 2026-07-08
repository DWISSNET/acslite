// Package scheduler provides distributed background jobs for ACSGO.
//
// Jobs use Redis distributed locks to prevent racing on multi-node deployments.
package scheduler

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/DWISSNET/acsgo/pkg/cache"
	"github.com/DWISSNET/acsgo/pkg/repository"
	"github.com/DWISSNET/acsgo/services/metrics"
)

// Scheduler runs background maintenance tasks.
type Scheduler struct {
	deviceRepo  *repository.DeviceRepository
	commandRepo *repository.CommandRepository
	cache       *cache.Client
	metrics     *metrics.Collector
	logger      *zap.Logger
	stop        chan struct{}
}

// NewScheduler creates a new Scheduler.
func NewScheduler(
	deviceRepo *repository.DeviceRepository,
	commandRepo *repository.CommandRepository,
	cacheClient *cache.Client,
	metricsCollector *metrics.Collector,
	logger *zap.Logger,
) *Scheduler {
	return &Scheduler{
		deviceRepo:  deviceRepo,
		commandRepo: commandRepo,
		cache:       cacheClient,
		metrics:     metricsCollector,
		logger:      logger,
		stop:        make(chan struct{}),
	}
}

// Start launches all background jobs.
func (s *Scheduler) Start() {
	go s.runWithTicker("offline-checker", 2*time.Minute, s.markOfflineDevices)
	go s.runWithTicker("command-expiry", 5*time.Minute, s.expireCommands)
	go s.runWithTicker("metrics-aggregator", 1*time.Minute, s.aggregateMetrics)
}

// Stop signals all background jobs to exit.
func (s *Scheduler) Stop() {
	close(s.stop)
}

func (s *Scheduler) runWithTicker(name string, interval time.Duration, fn func(ctx context.Context)) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stop:
			return
		case <-ticker.C:
			// Acquire distributed lock so only one instance runs the job
			ctx, cancel := context.WithTimeout(context.Background(), interval-5*time.Second)
			acquired, err := s.cache.AcquireLock(ctx, "scheduler:"+name)
			if err != nil || !acquired {
				cancel()
				continue
			}
			s.logger.Debug("Running scheduled job", zap.String("job", name))
			fn(ctx)
			_ = s.cache.ReleaseLock(ctx, "scheduler:"+name)
			cancel()
		}
	}
}

// markOfflineDevices detects devices that have stopped sending Informs.
func (s *Scheduler) markOfflineDevices(ctx context.Context) {
	cutoff := time.Now().Add(-10 * time.Minute) // 10 min without Inform = offline
	n, err := s.deviceRepo.SetOfflineSince(ctx, cutoff)
	if err != nil {
		s.logger.Error("markOfflineDevices error", zap.Error(err))
		return
	}
	if n > 0 {
		s.logger.Info("Marked devices offline", zap.Int64("count", n))
	}
}

// expireCommands marks stuck commands as expired.
func (s *Scheduler) expireCommands(ctx context.Context) {
	n, err := s.commandRepo.ExpireOld(ctx)
	if err != nil {
		s.logger.Error("expireCommands error", zap.Error(err))
		return
	}
	if n > 0 {
		s.logger.Info("Expired stale commands", zap.Int64("count", n))
	}
}

// aggregateMetrics collects device counts for Prometheus.
func (s *Scheduler) aggregateMetrics(ctx context.Context) {
	// Count online devices
	onlineDevices, _, _ := s.deviceRepo.List(ctx, repository.DeviceFilter{Status: "online", Limit: 1})
	offlineDevices, _, _ := s.deviceRepo.List(ctx, repository.DeviceFilter{Status: "offline", Limit: 1})
	// Get total counts
	_, onlineTotal, _ := s.deviceRepo.List(ctx, repository.DeviceFilter{Status: "online"})
	_, offlineTotal, _ := s.deviceRepo.List(ctx, repository.DeviceFilter{Status: "offline"})
	_ = onlineDevices
	_ = offlineDevices
	s.metrics.SetDeviceCounts(onlineTotal, offlineTotal)
}
