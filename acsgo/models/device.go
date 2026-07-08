package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// JSONMap is a helper for storing JSON blobs in SQL columns
type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONMap) Scan(src interface{}) error {
	var b []byte
	switch v := src.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	case nil:
		*j = JSONMap{}
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
	return json.Unmarshal(b, j)
}

// JSONStrings handles JSON arrays of strings
type JSONStrings []string

func (j JSONStrings) Value() (driver.Value, error) {
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONStrings) Scan(src interface{}) error {
	var b []byte
	switch v := src.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	case nil:
		*j = JSONStrings{}
		return nil
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
	return json.Unmarshal(b, j)
}

// Device represents a managed TR-069 device
type Device struct {
	ID               string     `json:"id"`
	SerialNumber     string     `json:"serial_number"`
	DeviceID         string     `json:"device_id"`
	Manufacturer     string     `json:"manufacturer"`
	ModelName        string     `json:"model_name"`
	ProductClass     string     `json:"product_class"`
	SoftwareVersion  string     `json:"software_version"`
	HardwareVersion  string     `json:"hardware_version"`
	IPAddress        string     `json:"ip_address"`
	MACAddress       string     `json:"mac_address"`
	ConnectionStatus string     `json:"connection_status"` // online/offline/suspended
	ISP              string     `json:"isp"`
	CustomerName     string     `json:"customer_name"`
	CustomerPhone    string     `json:"customer_phone"`
	CustomerEmail    string     `json:"customer_email"`
	CustomerPackage  string     `json:"customer_package"`
	LocationCity     string     `json:"location_city"`
	LocationProvince string     `json:"location_province"`
	Parameters       JSONMap    `json:"parameters"`
	LastInform       *time.Time `json:"last_inform"`
	FirstSeen        time.Time  `json:"first_seen"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// DeviceEvent represents a log entry for a device
type DeviceEvent struct {
	ID          int64     `json:"id"`
	DeviceSerial string   `json:"device_serial"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
