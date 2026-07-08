-- ACSGO PostgreSQL Schema
-- Version: 1.0.0
-- Description: Complete schema for ACSGO TR-069 ACS Server

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- For fast text search

-- ============================================================
-- USERS & RBAC
-- ============================================================
CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username      VARCHAR(64)  NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(32)  NOT NULL DEFAULT 'operator', -- admin, operator, viewer
    active        BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email    ON users(email);

-- ============================================================
-- DEVICES
-- ============================================================
CREATE TABLE IF NOT EXISTS devices (
    id                 UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id          VARCHAR(255) NOT NULL UNIQUE,
    serial_number      VARCHAR(255) NOT NULL,
    product_class      VARCHAR(128),
    manufacturer       VARCHAR(128),
    model_name         VARCHAR(128),
    software_version   VARCHAR(64),
    hardware_version   VARCHAR(64),
    ip_address         INET,
    mac_address        VARCHAR(17),
    last_inform        TIMESTAMPTZ,
    first_seen         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    connection_status  VARCHAR(16)  NOT NULL DEFAULT 'offline',
    isp                VARCHAR(128),
    tags               TEXT[]       NOT NULL DEFAULT '{}',

    -- Location
    city               VARCHAR(128),
    province           VARCHAR(128),
    latitude           DOUBLE PRECISION,
    longitude          DOUBLE PRECISION,

    -- Customer info
    customer_name      VARCHAR(255),
    customer_phone     VARCHAR(64),
    customer_email     VARCHAR(255),
    customer_package   VARCHAR(128),

    -- Health score (0-100, computed)
    health_score       SMALLINT     NOT NULL DEFAULT 100,

    created_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_devices_device_id         ON devices(device_id);
CREATE INDEX IF NOT EXISTS idx_devices_serial_number     ON devices(serial_number);
CREATE INDEX IF NOT EXISTS idx_devices_manufacturer      ON devices(manufacturer);
CREATE INDEX IF NOT EXISTS idx_devices_model_name        ON devices(model_name);
CREATE INDEX IF NOT EXISTS idx_devices_connection_status ON devices(connection_status);
CREATE INDEX IF NOT EXISTS idx_devices_last_inform       ON devices(last_inform DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_devices_isp               ON devices(isp);
CREATE INDEX IF NOT EXISTS idx_devices_mac               ON devices(mac_address);
-- GIN index for tags array
CREATE INDEX IF NOT EXISTS idx_devices_tags              ON devices USING GIN(tags);
-- Trigram index for text search on device_id/serial/customer
CREATE INDEX IF NOT EXISTS idx_devices_trgm_device_id   ON devices USING GIN(device_id gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_devices_trgm_serial      ON devices USING GIN(serial_number gin_trgm_ops);

-- ============================================================
-- DEVICE PARAMETERS (TR-069 parameter tree)
-- ============================================================
CREATE TABLE IF NOT EXISTS device_parameters (
    id           UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id    UUID         NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    path         VARCHAR(512) NOT NULL,
    value        TEXT         NOT NULL DEFAULT '',
    value_type   VARCHAR(32)  NOT NULL DEFAULT 'string',
    last_changed TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (device_id, path)
);

CREATE INDEX IF NOT EXISTS idx_device_params_device_id ON device_parameters(device_id);
CREATE INDEX IF NOT EXISTS idx_device_params_path      ON device_parameters(path);

-- ============================================================
-- DEVICE EVENTS
-- ============================================================
CREATE TABLE IF NOT EXISTS device_events (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id   UUID         NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    event_code  VARCHAR(64)  NOT NULL,
    description TEXT,
    timestamp   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_events_device_id ON device_events(device_id);
CREATE INDEX IF NOT EXISTS idx_device_events_timestamp ON device_events(timestamp DESC);

-- ============================================================
-- RPC COMMANDS QUEUE
-- ============================================================
CREATE TABLE IF NOT EXISTS commands (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id     UUID         NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    command_type  VARCHAR(64)  NOT NULL,
    parameters    JSONB        NOT NULL DEFAULT '{}',
    status        VARCHAR(16)  NOT NULL DEFAULT 'pending',
    retry_count   SMALLINT     NOT NULL DEFAULT 0,
    max_retries   SMALLINT     NOT NULL DEFAULT 3,
    created_by    UUID         REFERENCES users(id),
    response      TEXT,
    error_msg     TEXT,
    scheduled_at  TIMESTAMPTZ,
    delivered_at  TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    expires_at    TIMESTAMPTZ,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_commands_device_id  ON commands(device_id);
CREATE INDEX IF NOT EXISTS idx_commands_status     ON commands(status);
CREATE INDEX IF NOT EXISTS idx_commands_created_at ON commands(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_commands_pending    ON commands(device_id, status) WHERE status IN ('pending','queued');

-- ============================================================
-- PARAMETER TEMPLATES
-- ============================================================
CREATE TABLE IF NOT EXISTS parameter_templates (
    id           UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name         VARCHAR(255) NOT NULL UNIQUE,
    description  TEXT,
    manufacturer VARCHAR(128),
    model_name   VARCHAR(128),
    isp          VARCHAR(128),
    parameters   JSONB        NOT NULL DEFAULT '[]',
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- AUDIT LOG
-- ============================================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    entity_type VARCHAR(64)  NOT NULL,
    entity_id   VARCHAR(255) NOT NULL,
    action      VARCHAR(64)  NOT NULL,
    old_value   TEXT,
    new_value   TEXT,
    user_id     UUID         REFERENCES users(id),
    ip_address  INET,
    timestamp   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_entity     ON audit_logs(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_user       ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp  ON audit_logs(timestamp DESC);

-- ============================================================
-- METRICS (time-series analytics)
-- ============================================================
CREATE TABLE IF NOT EXISTS metric_samples (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_id  UUID         NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    metric     VARCHAR(128) NOT NULL,
    value      DOUBLE PRECISION NOT NULL,
    labels     JSONB        NOT NULL DEFAULT '{}',
    timestamp  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metrics_device_metric ON metric_samples(device_id, metric, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_metrics_timestamp     ON metric_samples(timestamp DESC);

-- Partition by month for large deployments (optional for production):
-- CREATE TABLE metric_samples_2026_01 PARTITION OF metric_samples
--   FOR VALUES FROM ('2026-01-01') TO ('2026-02-01');

-- ============================================================
-- AUTOMATION RULES
-- ============================================================
CREATE TABLE IF NOT EXISTS automation_rules (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(255) NOT NULL,
    description   TEXT,
    event_type    VARCHAR(64)  NOT NULL,
    conditions    JSONB        NOT NULL DEFAULT '{}',
    action_type   VARCHAR(64)  NOT NULL,
    action_params JSONB        NOT NULL DEFAULT '{}',
    enabled       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_automation_event_type ON automation_rules(event_type) WHERE enabled = TRUE;

-- ============================================================
-- DEFAULT ADMIN USER (change password in production!)
-- ============================================================
-- Password hash below is bcrypt of 'admin123' — CHANGE IN PRODUCTION
INSERT INTO users (username, email, password_hash, role)
VALUES ('admin', 'admin@acsgo.local', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj4tbQBqY4Fy', 'admin')
ON CONFLICT (username) DO NOTHING;

-- ============================================================
-- HELPER FUNCTION: update updated_at automatically
-- ============================================================
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER set_devices_updated_at
  BEFORE UPDATE ON devices
  FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE OR REPLACE TRIGGER set_commands_updated_at
  BEFORE UPDATE ON commands
  FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE OR REPLACE TRIGGER set_templates_updated_at
  BEFORE UPDATE ON parameter_templates
  FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE OR REPLACE TRIGGER set_automation_updated_at
  BEFORE UPDATE ON automation_rules
  FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
