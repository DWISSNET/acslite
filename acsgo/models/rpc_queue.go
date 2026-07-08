package models

import "time"

// RPCQueue holds outbound commands queued for a device
type RPCQueue struct {
	ID           string     `json:"id"`
	DeviceSerial string     `json:"device_serial"`
	CommandType  string     `json:"command_type"` // Reboot/FactoryReset/SetParameterValues/GetParameterValues
	Parameters   JSONMap    `json:"parameters"`
	Status       string     `json:"status"` // pending/sent/completed/failed
	CreatedAt    time.Time  `json:"created_at"`
	DeliveredAt  *time.Time `json:"delivered_at,omitempty"`
	ErrorMessage string     `json:"error_message,omitempty"`
}
