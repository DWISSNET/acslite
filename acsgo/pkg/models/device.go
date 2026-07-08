// Package models defines core data structures for ACSGO.
package models

import (
	"time"
)

// ConnectionStatus represents the device connection state.
type ConnectionStatus string

const (
	ConnectionOnline    ConnectionStatus = "online"
	ConnectionOffline   ConnectionStatus = "offline"
	ConnectionSuspended ConnectionStatus = "suspended"
)

// Device represents a TR-069 managed CPE device.
type Device struct {
	ID               string           `json:"id" db:"id"`
	DeviceID         string           `json:"device_id" db:"device_id"`
	SerialNumber     string           `json:"serial_number" db:"serial_number"`
	ProductClass     string           `json:"product_class" db:"product_class"`
	Manufacturer     string           `json:"manufacturer" db:"manufacturer"`
	ModelName        string           `json:"model_name" db:"model_name"`
	SoftwareVersion  string           `json:"software_version" db:"software_version"`
	HardwareVersion  string           `json:"hardware_version" db:"hardware_version"`
	IPAddress        string           `json:"ip_address" db:"ip_address"`
	MACAddress       string           `json:"mac_address" db:"mac_address"`
	LastInform       *time.Time       `json:"last_inform" db:"last_inform"`
	FirstSeen        time.Time        `json:"first_seen" db:"first_seen"`
	ConnectionStatus ConnectionStatus `json:"connection_status" db:"connection_status"`
	ISP              string           `json:"isp" db:"isp"`
	Tags             []string         `json:"tags" db:"tags"`

	// Location
	City        string   `json:"city" db:"city"`
	Province    string   `json:"province" db:"province"`
	Latitude    *float64 `json:"latitude,omitempty" db:"latitude"`
	Longitude   *float64 `json:"longitude,omitempty" db:"longitude"`

	// Customer
	CustomerName    string `json:"customer_name" db:"customer_name"`
	CustomerPhone   string `json:"customer_phone" db:"customer_phone"`
	CustomerEmail   string `json:"customer_email" db:"customer_email"`
	CustomerPackage string `json:"customer_package" db:"customer_package"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
}

// DeviceParameter holds a single TR-069 parameter value for a device.
type DeviceParameter struct {
	ID          string    `json:"id" db:"id"`
	DeviceID    string    `json:"device_id" db:"device_id"`
	Path        string    `json:"path" db:"path"`
	Value       string    `json:"value" db:"value"`
	ValueType   string    `json:"value_type" db:"value_type"`
	LastChanged time.Time `json:"last_changed" db:"last_changed"`
}

// DeviceEvent records a TR-069 event on a device.
type DeviceEvent struct {
	ID          string    `json:"id" db:"id"`
	DeviceID    string    `json:"device_id" db:"device_id"`
	EventCode   string    `json:"event_code" db:"event_code"`
	Description string    `json:"description" db:"description"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
}

// ParameterTemplate defines a reusable parameter configuration.
type ParameterTemplate struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Description  string    `json:"description" db:"description"`
	Manufacturer string    `json:"manufacturer" db:"manufacturer"`
	ModelName    string    `json:"model_name" db:"model_name"`
	ISP          string    `json:"isp" db:"isp"`
	Parameters   []TemplateParameter `json:"parameters" db:"parameters"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// TemplateParameter is a single parameter within a template.
type TemplateParameter struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

// User represents an authenticated user of the ACS dashboard.
type User struct {
	ID           string    `json:"id" db:"id"`
	Username     string    `json:"username" db:"username"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Role         string    `json:"role" db:"role"`
	Active       bool      `json:"active" db:"active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
