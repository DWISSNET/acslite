package models

import "time"

// CommandType defines the type of RPC command to send to a device.
type CommandType string

const (
	CommandReboot              CommandType = "Reboot"
	CommandFactoryReset        CommandType = "FactoryReset"
	CommandSetParameterValues  CommandType = "SetParameterValues"
	CommandGetParameterValues  CommandType = "GetParameterValues"
	CommandDownload            CommandType = "Download"
	CommandUpload              CommandType = "Upload"
	CommandAddObject           CommandType = "AddObject"
	CommandDeleteObject        CommandType = "DeleteObject"
	CommandGetParameterNames   CommandType = "GetParameterNames"
	CommandScheduleInform      CommandType = "ScheduleInform"
	CommandChangeDUState       CommandType = "ChangeDUState"
)

// CommandStatus tracks the lifecycle of an RPC command.
type CommandStatus string

const (
	CommandPending    CommandStatus = "pending"
	CommandQueued     CommandStatus = "queued"
	CommandDelivered  CommandStatus = "delivered"
	CommandCompleted  CommandStatus = "completed"
	CommandFailed     CommandStatus = "failed"
	CommandExpired    CommandStatus = "expired"
)

// Command represents an RPC command to be delivered to a device.
type Command struct {
	ID          string            `json:"id" db:"id"`
	DeviceID    string            `json:"device_id" db:"device_id"`
	CommandType CommandType       `json:"command_type" db:"command_type"`
	Parameters  map[string]string `json:"parameters" db:"parameters"`
	Status      CommandStatus     `json:"status" db:"status"`
	RetryCount  int               `json:"retry_count" db:"retry_count"`
	MaxRetries  int               `json:"max_retries" db:"max_retries"`
	CreatedBy   string            `json:"created_by" db:"created_by"`
	Response    string            `json:"response" db:"response"`
	ErrorMsg    string            `json:"error_msg" db:"error_msg"`
	ScheduledAt *time.Time        `json:"scheduled_at" db:"scheduled_at"`
	DeliveredAt *time.Time        `json:"delivered_at" db:"delivered_at"`
	CompletedAt *time.Time        `json:"completed_at" db:"completed_at"`
	ExpiresAt   *time.Time        `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}

// AuditLog records every state change for compliance.
type AuditLog struct {
	ID         string    `json:"id" db:"id"`
	EntityType string    `json:"entity_type" db:"entity_type"`
	EntityID   string    `json:"entity_id" db:"entity_id"`
	Action     string    `json:"action" db:"action"`
	OldValue   string    `json:"old_value" db:"old_value"`
	NewValue   string    `json:"new_value" db:"new_value"`
	UserID     string    `json:"user_id" db:"user_id"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	Timestamp  time.Time `json:"timestamp" db:"timestamp"`
}

// MetricSample stores a time-series data point for device analytics.
type MetricSample struct {
	ID        string    `json:"id" db:"id"`
	DeviceID  string    `json:"device_id" db:"device_id"`
	Metric    string    `json:"metric" db:"metric"`
	Value     float64   `json:"value" db:"value"`
	Labels    map[string]string `json:"labels" db:"labels"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
}

// AutomationRule defines an event-driven trigger and action pair.
type AutomationRule struct {
	ID          string            `json:"id" db:"id"`
	Name        string            `json:"name" db:"name"`
	Description string            `json:"description" db:"description"`
	EventType   string            `json:"event_type" db:"event_type"`
	Conditions  map[string]string `json:"conditions" db:"conditions"`
	ActionType  string            `json:"action_type" db:"action_type"`
	ActionParams map[string]string `json:"action_params" db:"action_params"`
	Enabled     bool              `json:"enabled" db:"enabled"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
}
