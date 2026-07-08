// Package repository provides PostgreSQL data access for ACSGO devices.
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/DWISSNET/acsgo/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DeviceRepository handles device persistence in PostgreSQL.
type DeviceRepository struct {
	db *pgxpool.Pool
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db *pgxpool.Pool) *DeviceRepository {
	return &DeviceRepository{db: db}
}

// Upsert creates or updates a device based on device_id.
func (r *DeviceRepository) Upsert(ctx context.Context, d *models.Device) error {
	q := `
INSERT INTO devices (
    id, device_id, serial_number, product_class, manufacturer, model_name,
    software_version, hardware_version, ip_address, mac_address,
    last_inform, connection_status, isp, tags,
    city, province, latitude, longitude,
    customer_name, customer_phone, customer_email, customer_package,
    created_at, updated_at
) VALUES (
    uuid_generate_v4(), $1, $2, $3, $4, $5,
    $6, $7, $8::inet, $9,
    $10, $11, $12, $13,
    $14, $15, $16, $17,
    $18, $19, $20, $21,
    NOW(), NOW()
)
ON CONFLICT (device_id) DO UPDATE SET
    serial_number    = EXCLUDED.serial_number,
    product_class    = EXCLUDED.product_class,
    manufacturer     = EXCLUDED.manufacturer,
    model_name       = EXCLUDED.model_name,
    software_version = EXCLUDED.software_version,
    hardware_version = EXCLUDED.hardware_version,
    ip_address       = EXCLUDED.ip_address,
    mac_address      = EXCLUDED.mac_address,
    last_inform      = EXCLUDED.last_inform,
    connection_status= EXCLUDED.connection_status,
    isp              = EXCLUDED.isp,
    updated_at       = NOW()
RETURNING id`

	tags, _ := json.Marshal(d.Tags)
	_ = tags

	var id string
	err := r.db.QueryRow(ctx, q,
		d.DeviceID, d.SerialNumber, d.ProductClass, d.Manufacturer, d.ModelName,
		d.SoftwareVersion, d.HardwareVersion, d.IPAddress, d.MACAddress,
		d.LastInform, d.ConnectionStatus, d.ISP, d.Tags,
		d.City, d.Province, d.Latitude, d.Longitude,
		d.CustomerName, d.CustomerPhone, d.CustomerEmail, d.CustomerPackage,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("upsert device: %w", err)
	}
	if d.ID == "" {
		d.ID = id
	}
	return nil
}

// GetByDeviceID retrieves a device by its CWMP device ID.
func (r *DeviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*models.Device, error) {
	q := `
SELECT id, device_id, serial_number, product_class, manufacturer, model_name,
       software_version, hardware_version, ip_address::text, mac_address,
       last_inform, first_seen, connection_status, isp, tags,
       city, province, latitude, longitude,
       customer_name, customer_phone, customer_email, customer_package,
       created_at, updated_at
FROM devices WHERE device_id = $1`

	row := r.db.QueryRow(ctx, q, deviceID)
	return scanDevice(row)
}

// GetByID retrieves a device by its UUID.
func (r *DeviceRepository) GetByID(ctx context.Context, id string) (*models.Device, error) {
	q := `
SELECT id, device_id, serial_number, product_class, manufacturer, model_name,
       software_version, hardware_version, ip_address::text, mac_address,
       last_inform, first_seen, connection_status, isp, tags,
       city, province, latitude, longitude,
       customer_name, customer_phone, customer_email, customer_package,
       created_at, updated_at
FROM devices WHERE id = $1`

	row := r.db.QueryRow(ctx, q, id)
	return scanDevice(row)
}

// DeviceFilter holds search/filter criteria for device listing.
type DeviceFilter struct {
	Manufacturer string
	ModelName    string
	ISP          string
	Status       string
	Search       string
	Tags         []string
	Limit        int
	Offset       int
}

// List returns a paginated list of devices.
func (r *DeviceRepository) List(ctx context.Context, f DeviceFilter) ([]*models.Device, int, error) {
	args := []interface{}{}
	where := []string{}
	n := 1

	if f.Manufacturer != "" {
		where = append(where, fmt.Sprintf("manufacturer = $%d", n))
		args = append(args, f.Manufacturer)
		n++
	}
	if f.ModelName != "" {
		where = append(where, fmt.Sprintf("model_name = $%d", n))
		args = append(args, f.ModelName)
		n++
	}
	if f.ISP != "" {
		where = append(where, fmt.Sprintf("isp = $%d", n))
		args = append(args, f.ISP)
		n++
	}
	if f.Status != "" {
		where = append(where, fmt.Sprintf("connection_status = $%d", n))
		args = append(args, f.Status)
		n++
	}
	if f.Search != "" {
		where = append(where, fmt.Sprintf("(device_id ILIKE $%d OR serial_number ILIKE $%d OR customer_name ILIKE $%d)", n, n, n))
		args = append(args, "%"+f.Search+"%")
		n++
	}
	if len(f.Tags) > 0 {
		where = append(where, fmt.Sprintf("tags @> $%d", n))
		args = append(args, f.Tags)
		n++
	}

	cond := ""
	if len(where) > 0 {
		cond = "WHERE " + strings.Join(where, " AND ")
	}

	countQ := fmt.Sprintf("SELECT COUNT(*) FROM devices %s", cond)
	var total int
	if err := r.db.QueryRow(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count devices: %w", err)
	}

	if f.Limit == 0 {
		f.Limit = 50
	}

	listQ := fmt.Sprintf(`
SELECT id, device_id, serial_number, product_class, manufacturer, model_name,
       software_version, hardware_version, ip_address::text, mac_address,
       last_inform, first_seen, connection_status, isp, tags,
       city, province, latitude, longitude,
       customer_name, customer_phone, customer_email, customer_package,
       created_at, updated_at
FROM devices %s
ORDER BY last_inform DESC NULLS LAST
LIMIT $%d OFFSET $%d`, cond, n, n+1)

	args = append(args, f.Limit, f.Offset)
	rows, err := r.db.Query(ctx, listQ, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list devices: %w", err)
	}
	defer rows.Close()

	var devices []*models.Device
	for rows.Next() {
		d, err := scanDevice(rows)
		if err != nil {
			return nil, 0, err
		}
		devices = append(devices, d)
	}
	return devices, total, nil
}

// SetStatus updates a device's connection status.
func (r *DeviceRepository) SetStatus(ctx context.Context, deviceID string, status models.ConnectionStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE devices SET connection_status = $1, updated_at = NOW() WHERE device_id = $2`,
		status, deviceID,
	)
	return err
}

// SetOfflineSince marks all devices offline that haven't sent an Inform since cutoff.
func (r *DeviceRepository) SetOfflineSince(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := r.db.Exec(ctx,
		`UPDATE devices SET connection_status = 'offline', updated_at = NOW()
         WHERE connection_status = 'online' AND (last_inform IS NULL OR last_inform < $1)`,
		cutoff,
	)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// UpsertParameter stores or updates a TR-069 parameter value.
func (r *DeviceRepository) UpsertParameter(ctx context.Context, deviceUUID, path, value, valueType string) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO device_parameters (id, device_id, path, value, value_type, last_changed)
VALUES (uuid_generate_v4(), $1::uuid, $2, $3, $4, NOW())
ON CONFLICT (device_id, path) DO UPDATE SET
    value        = EXCLUDED.value,
    value_type   = EXCLUDED.value_type,
    last_changed = NOW()`,
		deviceUUID, path, value, valueType,
	)
	return err
}

// GetParameters returns all parameters for a device.
func (r *DeviceRepository) GetParameters(ctx context.Context, deviceUUID string) ([]*models.DeviceParameter, error) {
	rows, err := r.db.Query(ctx, `
SELECT id::text, device_id::text, path, value, value_type, last_changed
FROM device_parameters
WHERE device_id = $1::uuid
ORDER BY path`, deviceUUID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var params []*models.DeviceParameter
	for rows.Next() {
		p := &models.DeviceParameter{}
		if err := rows.Scan(&p.ID, &p.DeviceID, &p.Path, &p.Value, &p.ValueType, &p.LastChanged); err != nil {
			return nil, err
		}
		params = append(params, p)
	}
	return params, nil
}

// AddEvent appends a device event to the log.
func (r *DeviceRepository) AddEvent(ctx context.Context, deviceUUID, code, description string) error {
	_, err := r.db.Exec(ctx, `
INSERT INTO device_events (id, device_id, event_code, description, timestamp)
VALUES (uuid_generate_v4(), $1::uuid, $2, $3, NOW())`,
		deviceUUID, code, description)
	return err
}

// scanDevice reads a device row from a pgx.Row or pgx.Rows.
func scanDevice(row pgx.Row) (*models.Device, error) {
	d := &models.Device{}
	err := row.Scan(
		&d.ID, &d.DeviceID, &d.SerialNumber, &d.ProductClass,
		&d.Manufacturer, &d.ModelName, &d.SoftwareVersion, &d.HardwareVersion,
		&d.IPAddress, &d.MACAddress, &d.LastInform, &d.FirstSeen,
		&d.ConnectionStatus, &d.ISP, &d.Tags,
		&d.City, &d.Province, &d.Latitude, &d.Longitude,
		&d.CustomerName, &d.CustomerPhone, &d.CustomerEmail, &d.CustomerPackage,
		&d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("scan device: %w", err)
	}
	return d, nil
}
