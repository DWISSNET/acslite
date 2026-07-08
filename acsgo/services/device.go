package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/DWISSNET/acsgo/models"
	"github.com/DWISSNET/acsgo/pkg"
)

// DeviceService handles device persistence and lookup
type DeviceService struct {
	db *sql.DB
}

func NewDeviceService(db *sql.DB) *DeviceService {
	return &DeviceService{db: db}
}

// Upsert inserts or updates a device record, marking it online.
func (s *DeviceService) Upsert(d *models.Device) error {
	params, err := json.Marshal(d.Parameters)
	if err != nil {
		params = []byte("{}")
	}

	now := time.Now()
	d.UpdatedAt = now
	if d.LastInform == nil {
		d.LastInform = &now
	}
	if d.ID == "" {
		d.ID = pkg.NewUUID()
	}
	if d.FirstSeen.IsZero() {
		d.FirstSeen = now
	}

	_, err = s.db.Exec(`
		INSERT INTO devices (
			id, serial_number, device_id, manufacturer, model_name, product_class,
			software_version, hardware_version, ip_address, mac_address,
			connection_status, isp,
			customer_name, customer_phone, customer_email, customer_package,
			location_city, location_province, parameters,
			last_inform, first_seen, updated_at
		) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(serial_number) DO UPDATE SET
			device_id         = excluded.device_id,
			manufacturer      = excluded.manufacturer,
			model_name        = excluded.model_name,
			product_class     = excluded.product_class,
			software_version  = excluded.software_version,
			hardware_version  = excluded.hardware_version,
			ip_address        = excluded.ip_address,
			mac_address       = excluded.mac_address,
			connection_status = excluded.connection_status,
			parameters        = excluded.parameters,
			last_inform       = excluded.last_inform,
			updated_at        = excluded.updated_at
	`,
		d.ID, d.SerialNumber, d.DeviceID, d.Manufacturer, d.ModelName, d.ProductClass,
		d.SoftwareVersion, d.HardwareVersion, d.IPAddress, d.MACAddress,
		d.ConnectionStatus, d.ISP,
		d.CustomerName, d.CustomerPhone, d.CustomerEmail, d.CustomerPackage,
		d.LocationCity, d.LocationProvince, string(params),
		d.LastInform, d.FirstSeen, d.UpdatedAt,
	)
	return err
}

// GetBySerial returns a device by serial number.
func (s *DeviceService) GetBySerial(serial string) (*models.Device, error) {
	row := s.db.QueryRow(`SELECT id, serial_number, device_id, manufacturer, model_name,
		product_class, software_version, hardware_version, ip_address, mac_address,
		connection_status, isp, customer_name, customer_phone, customer_email, customer_package,
		location_city, location_province, parameters, last_inform, first_seen, updated_at
		FROM devices WHERE serial_number = ?`, serial)
	return scanDevice(row)
}

// List returns paginated devices with optional search filter.
func (s *DeviceService) List(page, limit int, search string) ([]*models.Device, int, error) {
	offset := (page - 1) * limit
	var where string
	var args []interface{}

	if search != "" {
		where = " WHERE serial_number LIKE ? OR mac_address LIKE ? OR model_name LIKE ? OR customer_name LIKE ?"
		like := "%" + search + "%"
		args = append(args, like, like, like, like)
	}

	countRow := s.db.QueryRow("SELECT COUNT(*) FROM devices"+where, args...)
	var total int
	if err := countRow.Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, limit, offset)
	rows, err := s.db.Query(`SELECT id, serial_number, device_id, manufacturer, model_name,
		product_class, software_version, hardware_version, ip_address, mac_address,
		connection_status, isp, customer_name, customer_phone, customer_email, customer_package,
		location_city, location_province, parameters, last_inform, first_seen, updated_at
		FROM devices`+where+` ORDER BY last_inform DESC LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		d, err := scanDeviceRow(rows)
		if err != nil {
			return nil, 0, err
		}
		devices = append(devices, d)
	}
	return devices, total, nil
}

// Delete removes a device by serial number.
func (s *DeviceService) Delete(serial string) error {
	_, err := s.db.Exec("DELETE FROM devices WHERE serial_number = ?", serial)
	return err
}

// MarkOffline sets connection_status=offline for devices not seen since threshold.
func (s *DeviceService) MarkOffline(threshold time.Duration) (int64, error) {
	cutoff := time.Now().Add(-threshold)
	res, err := s.db.Exec(`UPDATE devices SET connection_status = 'offline', updated_at = ?
		WHERE connection_status = 'online' AND (last_inform IS NULL OR last_inform < ?)`,
		time.Now(), cutoff)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// UpdateParameters stores a device's parameters JSON.
func (s *DeviceService) UpdateParameters(serial string, params models.JSONMap) error {
	b, err := json.Marshal(params)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("UPDATE devices SET parameters = ?, updated_at = ? WHERE serial_number = ?",
		string(b), time.Now(), serial)
	return err
}

// Stats returns aggregate statistics.
func (s *DeviceService) Stats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	row := s.db.QueryRow("SELECT COUNT(*) FROM devices")
	var total int
	if err := row.Scan(&total); err != nil {
		return nil, err
	}
	stats["total"] = total

	row = s.db.QueryRow("SELECT COUNT(*) FROM devices WHERE connection_status = 'online'")
	var online int
	if err := row.Scan(&online); err != nil {
		return nil, err
	}
	stats["online"] = online
	stats["offline"] = total - online
	return stats, nil
}

// LogEvent appends a device event.
func (s *DeviceService) LogEvent(serial, eventType, description string) {
	_, err := s.db.Exec(`INSERT INTO device_events (device_serial, event_type, description) VALUES (?,?,?)`,
		serial, eventType, description)
	if err != nil {
		log.Printf("⚠️  Failed to log event for %s: %v", serial, err)
	}
}

// GetEvents retrieves events for a device.
func (s *DeviceService) GetEvents(serial string, limit int) ([]models.DeviceEvent, error) {
	rows, err := s.db.Query(`SELECT id, device_serial, event_type, description, created_at
		FROM device_events WHERE device_serial = ? ORDER BY created_at DESC LIMIT ?`, serial, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.DeviceEvent
	for rows.Next() {
		var e models.DeviceEvent
		if err := rows.Scan(&e.ID, &e.DeviceSerial, &e.EventType, &e.Description, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

// --------------------------------------------------------------------------
// Supported modem database (same as TypeScript version)
// --------------------------------------------------------------------------

type ModemInfo struct {
	Model        string
	Name         string
	Type         string
	ISPs         []string
	Manufacturer string
}

var SupportedModems = map[string][]ModemInfo{
	"huawei": {
		{Model: "HG8245H", Name: "Huawei EchoLife HG8245H", Type: "GPON", ISPs: []string{"Indihome", "First Media"}},
		{Model: "HG8546M", Name: "Huawei EchoLife HG8546M", Type: "GPON", ISPs: []string{"Indihome"}},
		{Model: "EG8141A5", Name: "Huawei EchoLife EG8141A5", Type: "GPON", ISPs: []string{"Indihome"}},
		{Model: "EG8145V5", Name: "Huawei EG8145V5", Type: "GPON", ISPs: []string{"Indihome", "MyRepublic"}},
		{Model: "HG659", Name: "Huawei HG659", Type: "VDSL", ISPs: []string{"First Media", "MNC Play"}},
		{Model: "HS8546V5", Name: "Huawei HS8546V5", Type: "GPON", ISPs: []string{"Indihome", "Biznet"}},
		{Model: "B618s-22d", Name: "Huawei B618", Type: "LTE", ISPs: []string{"XL", "Telkomsel", "Indosat"}},
		{Model: "B525s-65a", Name: "Huawei B525", Type: "LTE", ISPs: []string{"Smartfren", "Tri"}},
	},
	"zte": {
		{Model: "F609", Name: "ZTE ZXHN F609", Type: "GPON", ISPs: []string{"Indihome"}},
		{Model: "F670L", Name: "ZTE ZXHN F670L", Type: "GPON", ISPs: []string{"Indihome", "First Media"}},
		{Model: "F670Y", Name: "ZTE F670Y", Type: "GPON", ISPs: []string{"Indihome"}},
		{Model: "F680", Name: "ZTE ZXHN F680", Type: "GPON", ISPs: []string{"Biznet"}},
		{Model: "ZXHN F601", Name: "ZTE ZXHN F601", Type: "GPON", ISPs: []string{"Indihome"}},
		{Model: "MF286R", Name: "ZTE MF286R", Type: "LTE", ISPs: []string{"XL", "Telkomsel"}},
	},
	"mikrotik": {
		{Model: "RB750Gr3", Name: "MikroTik hEX", Type: "RouterOS", ISPs: []string{"All"}},
		{Model: "RB4011iGS+", Name: "MikroTik RB4011", Type: "RouterOS", ISPs: []string{"All"}},
		{Model: "CCR1009", Name: "MikroTik CCR1009", Type: "RouterOS", ISPs: []string{"Biznet"}},
	},
	"tplink": {
		{Model: "XC220-G3v", Name: "TP-Link XC220-G3v", Type: "GPON", ISPs: []string{"Indihome"}},
		{Model: "AX1800", Name: "TP-Link Archer AX1800", Type: "WiFi6", ISPs: []string{"All"}},
	},
	"dlink": {
		{Model: "DPR-1041", Name: "D-Link DPR-1041", Type: "GPON", ISPs: []string{"First Media"}},
		{Model: "DSL-2750U", Name: "D-Link DSL-2750U", Type: "ADSL/VDSL", ISPs: []string{"All"}},
	},
}

// DetectModel attempts to identify modem model from manufacturer + product class.
func DetectModel(manufacturer, productClass, serial string) ModemInfo {
	vendor := pkg.DetectVendorFromManufacturer(manufacturer)
	modems, ok := SupportedModems[vendor]
	if !ok {
		return ModemInfo{
			Model:        productClass,
			Manufacturer: manufacturer,
			Type:         "Unknown",
			ISPs:         []string{"Unknown"},
		}
	}

	for _, m := range modems {
		if strings.Contains(productClass, m.Model) || strings.Contains(strings.ToUpper(serial), strings.ToUpper(m.Model)) {
			m.Manufacturer = vendor
			return m
		}
	}
	// Return first match for the vendor
	first := modems[0]
	first.Manufacturer = vendor
	return first
}

// GetAllModems returns a flat list of all supported modems.
func GetAllModems() []ModemInfo {
	var all []ModemInfo
	for vendor, modems := range SupportedModems {
		for _, m := range modems {
			m.Manufacturer = vendor
			all = append(all, m)
		}
	}
	return all
}

// --------------------------------------------------------------------------
// scanner helpers
// --------------------------------------------------------------------------

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanDevice(row rowScanner) (*models.Device, error) {
	return scanDeviceRow(row)
}

func scanDeviceRow(row rowScanner) (*models.Device, error) {
	var d models.Device
	var paramsJSON string
	var lastInform sql.NullTime
	var firstSeen, updatedAt time.Time

	err := row.Scan(
		&d.ID, &d.SerialNumber, &d.DeviceID, &d.Manufacturer, &d.ModelName,
		&d.ProductClass, &d.SoftwareVersion, &d.HardwareVersion, &d.IPAddress, &d.MACAddress,
		&d.ConnectionStatus, &d.ISP,
		&d.CustomerName, &d.CustomerPhone, &d.CustomerEmail, &d.CustomerPackage,
		&d.LocationCity, &d.LocationProvince,
		&paramsJSON, &lastInform, &firstSeen, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device not found")
	}
	if err != nil {
		return nil, err
	}
	if lastInform.Valid {
		d.LastInform = &lastInform.Time
	}
	d.FirstSeen = firstSeen
	d.UpdatedAt = updatedAt
	if err := json.Unmarshal([]byte(paramsJSON), &d.Parameters); err != nil {
		d.Parameters = models.JSONMap{}
	}
	return &d, nil
}
