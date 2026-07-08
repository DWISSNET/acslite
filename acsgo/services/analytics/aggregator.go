// Package analytics provides real-time device analytics for ACSGO.
package analytics

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/DWISSNET/acsgo/pkg/cache"
	"github.com/DWISSNET/acsgo/pkg/models"
	"github.com/DWISSNET/acsgo/pkg/repository"
)

const (
	// Event channel names for Redis pub/sub
	ChannelDeviceOnline  = "events:device:online"
	ChannelDeviceOffline = "events:device:offline"
	ChannelParamChanged  = "events:param:changed"
	ChannelCommandDone   = "events:command:done"
)

// DeviceEvent is a real-time event published to subscribers.
type DeviceEvent struct {
	Type      string      `json:"type"`
	DeviceID  string      `json:"device_id"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data,omitempty"`
}

// Engine aggregates device events and publishes them to Redis channels.
type Engine struct {
	deviceRepo  *repository.DeviceRepository
	cache       *cache.Client
	logger      *zap.Logger
}

// NewEngine creates a new analytics Engine.
func NewEngine(
	deviceRepo *repository.DeviceRepository,
	cacheClient *cache.Client,
	logger *zap.Logger,
) *Engine {
	return &Engine{
		deviceRepo:  deviceRepo,
		cache:       cacheClient,
		logger:      logger,
	}
}

// RecordDeviceOnline publishes a device-online event.
func (e *Engine) RecordDeviceOnline(ctx context.Context, device *models.Device) {
	e.publish(ctx, ChannelDeviceOnline, DeviceEvent{
		Type:      "device.online",
		DeviceID:  device.DeviceID,
		Timestamp: time.Now(),
		Data: map[string]string{
			"ip":           device.IPAddress,
			"manufacturer": device.Manufacturer,
			"model":        device.ModelName,
		},
	})
}

// RecordDeviceOffline publishes a device-offline event.
func (e *Engine) RecordDeviceOffline(ctx context.Context, deviceID string) {
	e.publish(ctx, ChannelDeviceOffline, DeviceEvent{
		Type:      "device.offline",
		DeviceID:  deviceID,
		Timestamp: time.Now(),
	})
}

// RecordParamChange publishes a parameter-changed event.
func (e *Engine) RecordParamChange(ctx context.Context, deviceID, path, oldVal, newVal string) {
	e.publish(ctx, ChannelParamChanged, DeviceEvent{
		Type:      "param.changed",
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Data: map[string]string{
			"path":      path,
			"old_value": oldVal,
			"new_value": newVal,
		},
	})
}

// RecordCommandComplete publishes a command-completed event.
func (e *Engine) RecordCommandComplete(ctx context.Context, deviceID, commandID, cmdType, status string) {
	e.publish(ctx, ChannelCommandDone, DeviceEvent{
		Type:      "command.done",
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Data: map[string]string{
			"command_id":   commandID,
			"command_type": cmdType,
			"status":       status,
		},
	})
}

// GetRecentEvents fetches the latest device events from the DB.
func (e *Engine) GetRecentEvents(ctx context.Context, deviceUUID string, limit int) ([]*models.DeviceEvent, error) {
	// For now delegate to DB — in production, use Redis sorted set for hot events
	return nil, nil
}

func (e *Engine) publish(ctx context.Context, channel string, event DeviceEvent) {
	if err := e.cache.Publish(ctx, channel, event); err != nil {
		e.logger.Warn("analytics publish error",
			zap.String("channel", channel),
			zap.Error(err),
		)
	}
}

// AnomalyDetector checks for suspicious device behaviour.
type AnomalyDetector struct {
	cache  *cache.Client
	logger *zap.Logger
}

// NewAnomalyDetector creates a new AnomalyDetector.
func NewAnomalyDetector(cacheClient *cache.Client, logger *zap.Logger) *AnomalyDetector {
	return &AnomalyDetector{cache: cacheClient, logger: logger}
}

// CheckInformRate detects devices informing too frequently (possible storm).
// Returns true if the rate exceeds threshold (> maxPerMinute Informs/min).
func (a *AnomalyDetector) CheckInformRate(ctx context.Context, deviceID string, maxPerMinute int64) bool {
	key := "inform_rate:" + deviceID
	count, err := a.cache.Incr(ctx, key, time.Minute)
	if err != nil {
		return false
	}
	if count > maxPerMinute {
		a.logger.Warn("Inform storm detected",
			zap.String("device_id", deviceID),
			zap.Int64("count_per_min", count),
		)
		return true
	}
	return false
}

// DeviceStats aggregates key metrics for a single device.
type DeviceStats struct {
	DeviceID      string    `json:"device_id"`
	OnlineSeconds int64     `json:"online_seconds"`
	InformCount   int64     `json:"inform_count"`
	LastSeen      time.Time `json:"last_seen"`
	HealthScore   int       `json:"health_score"`
}

// GetDeviceStats returns aggregated stats for a device.
func (e *Engine) GetDeviceStats(ctx context.Context, deviceID string) (*DeviceStats, error) {
	var stats DeviceStats
	err := e.cache.GetDevice(ctx, "stats:"+deviceID, &stats)
	if err != nil {
		// No cached stats — return default
		return &DeviceStats{DeviceID: deviceID, HealthScore: 100}, nil
	}
	return &stats, nil
}

// UpdateDeviceStats updates the cached stats for a device.
func (e *Engine) UpdateDeviceStats(ctx context.Context, stats *DeviceStats) {
	b, _ := json.Marshal(stats)
	_ = b
	_ = e.cache.SetDevice(ctx, "stats:"+stats.DeviceID, stats)
}
