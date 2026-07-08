// Package automation provides event-driven workflow automation for ACSGO.
package automation

import (
	"context"
	"encoding/json"
	"strings"

	"go.uber.org/zap"

	"github.com/DWISSNET/acsgo/pkg/cache"
	"github.com/DWISSNET/acsgo/pkg/models"
	"github.com/DWISSNET/acsgo/pkg/repository"
	"github.com/DWISSNET/acsgo/services/analytics"
)

// Engine evaluates automation rules and triggers actions.
type Engine struct {
	commandRepo *repository.CommandRepository
	cache       *cache.Client
	logger      *zap.Logger
	rules       []*models.AutomationRule
}

// NewEngine creates a new automation Engine.
func NewEngine(
	commandRepo *repository.CommandRepository,
	cacheClient *cache.Client,
	logger *zap.Logger,
) *Engine {
	return &Engine{
		commandRepo: commandRepo,
		cache:       cacheClient,
		logger:      logger,
	}
}

// LoadRules refreshes the rule set from the database/cache.
func (e *Engine) LoadRules(rules []*models.AutomationRule) {
	e.rules = rules
	e.logger.Info("Automation rules loaded", zap.Int("count", len(rules)))
}

// HandleEvent processes a device event against all enabled rules.
func (e *Engine) HandleEvent(ctx context.Context, event analytics.DeviceEvent) {
	for _, rule := range e.rules {
		if !rule.Enabled {
			continue
		}
		if rule.EventType != event.Type {
			continue
		}
		if !e.matchConditions(rule.Conditions, event) {
			continue
		}
		e.logger.Info("Automation rule triggered",
			zap.String("rule", rule.Name),
			zap.String("event", event.Type),
			zap.String("device_id", event.DeviceID),
		)
		e.executeAction(ctx, rule, event)
	}
}

// Subscribe listens to Redis events and invokes HandleEvent.
func (e *Engine) Subscribe(ctx context.Context) error {
	channels := []string{
		analytics.ChannelDeviceOnline,
		analytics.ChannelDeviceOffline,
		analytics.ChannelParamChanged,
		analytics.ChannelCommandDone,
	}
	for _, ch := range channels {
		ch := ch // capture range variable
		if err := e.cache.Subscribe(ctx, ch, func(payload []byte) {
			var ev analytics.DeviceEvent
			if err := json.Unmarshal(payload, &ev); err == nil {
				e.HandleEvent(ctx, ev)
			}
		}); err != nil {
			return err
		}
	}
	return nil
}

// matchConditions checks if event data satisfies rule conditions.
func (e *Engine) matchConditions(conditions map[string]string, event analytics.DeviceEvent) bool {
	if len(conditions) == 0 {
		return true
	}
	data, _ := json.Marshal(event.Data)
	dataStr := string(data)
	for k, v := range conditions {
		if !strings.Contains(dataStr, k+":") && !strings.Contains(dataStr, `"`+k+`"`) {
			return false
		}
		if v != "" && !strings.Contains(dataStr, v) {
			return false
		}
	}
	return true
}

// executeAction carries out the action defined in the rule.
func (e *Engine) executeAction(ctx context.Context, rule *models.AutomationRule, event analytics.DeviceEvent) {
	switch rule.ActionType {
	case "reboot":
		e.sendCommand(ctx, event.DeviceID, models.CommandReboot, nil)
	case "factory_reset":
		e.sendCommand(ctx, event.DeviceID, models.CommandFactoryReset, nil)
	case "set_parameter":
		e.sendCommand(ctx, event.DeviceID, models.CommandSetParameterValues, rule.ActionParams)
	case "get_parameter":
		params := map[string]string{}
		for k := range rule.ActionParams {
			params[k] = ""
		}
		e.sendCommand(ctx, event.DeviceID, models.CommandGetParameterValues, params)
	case "webhook":
		go e.callWebhook(ctx, rule.ActionParams, event)
	default:
		e.logger.Warn("Unknown automation action", zap.String("action", rule.ActionType))
	}
}

func (e *Engine) sendCommand(ctx context.Context, deviceID string, cmdType models.CommandType, params map[string]string) {
	if params == nil {
		params = map[string]string{}
	}
	cmd := &models.Command{
		DeviceID:    deviceID,
		CommandType: cmdType,
		Parameters:  params,
		Status:      models.CommandPending,
		MaxRetries:  3,
		CreatedBy:   "automation",
	}
	if err := e.commandRepo.Create(ctx, cmd); err != nil {
		e.logger.Error("Automation sendCommand failed",
			zap.String("device_id", deviceID),
			zap.String("type", string(cmdType)),
			zap.Error(err),
		)
	}
}

func (e *Engine) callWebhook(ctx context.Context, params map[string]string, event analytics.DeviceEvent) {
	url, ok := params["url"]
	if !ok || url == "" {
		return
	}
	// Webhook call is implemented in the automation HTTP client (see api/http)
	e.logger.Info("Webhook triggered",
		zap.String("url", url),
		zap.String("event_type", event.Type),
	)
}
