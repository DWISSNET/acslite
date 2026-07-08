-- ACSGO Database Schema
-- Compatible with both SQLite and PostgreSQL

CREATE TABLE IF NOT EXISTS devices (
    id                TEXT PRIMARY KEY,
    serial_number     TEXT NOT NULL UNIQUE,
    device_id         TEXT NOT NULL UNIQUE,
    manufacturer      TEXT NOT NULL DEFAULT '',
    model_name        TEXT NOT NULL DEFAULT '',
    product_class     TEXT NOT NULL DEFAULT '',
    software_version  TEXT NOT NULL DEFAULT '',
    hardware_version  TEXT NOT NULL DEFAULT '',
    ip_address        TEXT NOT NULL DEFAULT '',
    mac_address       TEXT NOT NULL DEFAULT '',
    connection_status TEXT NOT NULL DEFAULT 'offline',
    isp               TEXT NOT NULL DEFAULT '',
    customer_name     TEXT NOT NULL DEFAULT '',
    customer_phone    TEXT NOT NULL DEFAULT '',
    customer_email    TEXT NOT NULL DEFAULT '',
    customer_package  TEXT NOT NULL DEFAULT '',
    location_city     TEXT NOT NULL DEFAULT '',
    location_province TEXT NOT NULL DEFAULT '',
    parameters        TEXT NOT NULL DEFAULT '{}',
    last_inform       TIMESTAMP,
    first_seen        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_devices_serial      ON devices(serial_number);
CREATE INDEX IF NOT EXISTS idx_devices_manufacturer ON devices(manufacturer);
CREATE INDEX IF NOT EXISTS idx_devices_model        ON devices(model_name);
CREATE INDEX IF NOT EXISTS idx_devices_mac          ON devices(mac_address);
CREATE INDEX IF NOT EXISTS idx_devices_status       ON devices(connection_status);
CREATE INDEX IF NOT EXISTS idx_devices_isp          ON devices(isp);
CREATE INDEX IF NOT EXISTS idx_devices_last_inform  ON devices(last_inform);

CREATE TABLE IF NOT EXISTS device_events (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    device_serial TEXT NOT NULL,
    event_type    TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_events_serial ON device_events(device_serial);
CREATE INDEX IF NOT EXISTS idx_events_type   ON device_events(event_type);

CREATE TABLE IF NOT EXISTS parameters (
    path               TEXT PRIMARY KEY,
    name               TEXT NOT NULL DEFAULT '',
    description        TEXT NOT NULL DEFAULT '',
    type               TEXT NOT NULL DEFAULT 'string',
    category           TEXT NOT NULL DEFAULT '',
    subcategory        TEXT NOT NULL DEFAULT '',
    writable           INTEGER NOT NULL DEFAULT 0,
    default_value      TEXT NOT NULL DEFAULT '',
    supported_vendors  TEXT NOT NULL DEFAULT '["All"]',
    supported_models   TEXT NOT NULL DEFAULT '["All"]',
    is_standard_tr069  INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX IF NOT EXISTS idx_params_category ON parameters(category);

CREATE TABLE IF NOT EXISTS rpc_queue (
    id            TEXT PRIMARY KEY,
    device_serial TEXT NOT NULL,
    command_type  TEXT NOT NULL,
    parameters    TEXT NOT NULL DEFAULT '{}',
    status        TEXT NOT NULL DEFAULT 'pending',
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    delivered_at  TIMESTAMP,
    error_message TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_rpc_serial ON rpc_queue(device_serial);
CREATE INDEX IF NOT EXISTS idx_rpc_status ON rpc_queue(status);

CREATE TABLE IF NOT EXISTS users (
    id            TEXT PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login    TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
