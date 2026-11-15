-- ============================================================================
-- HRIG · Astronaut profiles and measurements
-- 001_schema.sql — generated from .routing/BD.md
-- Target: PostgreSQL 17+
-- ============================================================================

BEGIN;

CREATE TYPE metric_tier      AS ENUM ('primary', 'secondary');
CREATE TYPE value_type       AS ENUM ('numeric', 'scale', 'text', 'image');
CREATE TYPE environment_mode AS ENUM ('iva', 'eva');
CREATE TYPE device_kind      AS ENUM ('wearable', 'environment_sensor', 'manual', 'lab_assay', 'imaging');
CREATE TYPE media_kind       AS ENUM ('ultrasound', 'ocular');
CREATE TYPE reading_status   AS ENUM ('normal', 'abnormal', 'critical');

-- ----------------------------------------------------------------------------
-- 1. astronaut
-- ----------------------------------------------------------------------------
CREATE TABLE astronaut (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name         text NOT NULL,
    birth_date        date NOT NULL,
    nationality_code  char(2) NOT NULL,
    nationality_label text,
    is_active         boolean NOT NULL DEFAULT true,
    created_at        timestamptz NOT NULL DEFAULT now(),
    updated_at        timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT astronaut_full_name_len    CHECK (length(trim(full_name)) BETWEEN 1 AND 120),
    CONSTRAINT astronaut_birth_not_future CHECK (birth_date <= CURRENT_DATE),
    CONSTRAINT astronaut_nationality_iso  CHECK (nationality_code ~ '^[A-Z]{2}$')
);

-- ----------------------------------------------------------------------------
-- 2. unit
-- ----------------------------------------------------------------------------
CREATE TABLE unit (
    id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code      text NOT NULL UNIQUE,
    symbol    text NOT NULL,
    dimension text NOT NULL,
    decimals  smallint DEFAULT 1,
    is_active boolean NOT NULL DEFAULT true
);

-- ----------------------------------------------------------------------------
-- 3. metric
-- ----------------------------------------------------------------------------
CREATE TABLE metric (
    id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code                text NOT NULL UNIQUE,
    name                text NOT NULL,
    short_name          text,
    category            text NOT NULL,
    tier                metric_tier NOT NULL,
    metric_type         smallint NOT NULL,
    value_type          value_type NOT NULL,
    base_unit_id        uuid REFERENCES unit (id),
    display_unit_symbol text,
    scale_min           smallint,
    scale_max           smallint,
    display_order       smallint DEFAULT 0,
    is_active           boolean NOT NULL DEFAULT true,
    CONSTRAINT metric_type_domain        CHECK (metric_type IN (1, 2)),
    CONSTRAINT metric_scale_needs_range  CHECK (value_type <> 'scale' OR (scale_min IS NOT NULL AND scale_max IS NOT NULL)),
    CONSTRAINT metric_scale_range_order  CHECK (scale_min IS NULL OR scale_max IS NULL OR scale_min < scale_max),
    CONSTRAINT metric_numeric_needs_unit CHECK (value_type NOT IN ('numeric', 'scale') OR base_unit_id IS NOT NULL),
    CONSTRAINT metric_text_no_unit       CHECK (value_type NOT IN ('text', 'image') OR base_unit_id IS NULL)
);

-- Thresholds are stored in the metric's canonical unit. NULL means that side
-- has no configured limit; the server applies critical limits before warnings.
CREATE TABLE metric_threshold (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    metric_id       uuid NOT NULL UNIQUE REFERENCES metric (id) ON DELETE RESTRICT,
    warning_min     numeric(12,4),
    warning_max     numeric(12,4),
    critical_min    numeric(12,4),
    critical_max    numeric(12,4),
    classification_mode text NOT NULL DEFAULT 'numeric',
    source           text NOT NULL DEFAULT 'demo',
    updated_at       timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT metric_threshold_mode CHECK (classification_mode IN ('numeric', 'manual')),
    CONSTRAINT metric_threshold_has_limit CHECK (
        classification_mode = 'manual' OR warning_min IS NOT NULL OR warning_max IS NOT NULL OR
        critical_min IS NOT NULL OR critical_max IS NOT NULL
    ),
    CONSTRAINT metric_threshold_manual_no_limits CHECK (
        classification_mode <> 'manual' OR (warning_min IS NULL AND warning_max IS NULL AND critical_min IS NULL AND critical_max IS NULL)
    ),
    CONSTRAINT metric_threshold_low_order CHECK (
        critical_min IS NULL OR warning_min IS NULL OR critical_min < warning_min
    ),
    CONSTRAINT metric_threshold_high_order CHECK (
        critical_max IS NULL OR warning_max IS NULL OR critical_max > warning_max
    ),
    CONSTRAINT metric_threshold_order CHECK (
        warning_min IS NULL OR warning_max IS NULL OR warning_min <= warning_max
    )
);

-- ----------------------------------------------------------------------------
-- 4. device
-- ----------------------------------------------------------------------------
CREATE TABLE device (
    id        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    code      text NOT NULL UNIQUE,
    name      text NOT NULL,
    kind      device_kind NOT NULL,
    is_active boolean NOT NULL DEFAULT true
);

-- ----------------------------------------------------------------------------
-- 5. measurement
-- ----------------------------------------------------------------------------
CREATE TABLE measurement (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    astronaut_id    uuid REFERENCES astronaut (id),
    device_id       uuid REFERENCES device (id),
    environment     environment_mode NOT NULL,
    recorded_at_utc timestamptz NOT NULL,
    note            text,
    CONSTRAINT measurement_not_orphan CHECK (num_nonnulls(astronaut_id, device_id) >= 1)
);

CREATE INDEX idx_measurement_astronaut ON measurement (astronaut_id, recorded_at_utc DESC);
CREATE INDEX idx_measurement_device    ON measurement (device_id, recorded_at_utc DESC);

-- ----------------------------------------------------------------------------
-- 6. measurement_value
-- ----------------------------------------------------------------------------
CREATE TABLE measurement_value (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    measurement_id  uuid NOT NULL REFERENCES measurement (id) ON DELETE RESTRICT,
    metric_id       uuid NOT NULL REFERENCES metric (id),
    unit_id         uuid REFERENCES unit (id),
    recorded_at_utc timestamptz NOT NULL,
    value_numeric   numeric(12,4),
    value_scale     smallint,
    value_text      text,
    status          reading_status NOT NULL DEFAULT 'normal',
    note            text,
    CONSTRAINT measurement_value_once_per_metric UNIQUE (measurement_id, metric_id),
    CONSTRAINT measurement_value_has_value       CHECK (num_nonnulls(value_numeric, value_scale, value_text) >= 1),
    CONSTRAINT measurement_value_numeric_needs_unit CHECK (value_numeric IS NULL OR unit_id IS NOT NULL)
);

-- Serves both "latest value per metric" (HUD) and 24-hour trends.
CREATE INDEX idx_measurement_value_metric_time ON measurement_value (metric_id, recorded_at_utc DESC);

-- ----------------------------------------------------------------------------
-- 7. media_asset
-- ----------------------------------------------------------------------------
CREATE TABLE media_asset (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    measurement_id  uuid NOT NULL REFERENCES measurement (id),
    kind            media_kind NOT NULL,
    file_path       text NOT NULL,
    mime_type       text NOT NULL,
    findings        text,
    captured_at_utc timestamptz,
    CONSTRAINT media_asset_mime CHECK (mime_type IN ('image/jpeg', 'image/png', 'application/pdf'))
);

-- ----------------------------------------------------------------------------
-- Triggers
-- ----------------------------------------------------------------------------

-- astronaut.updated_at
CREATE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_astronaut_updated_at
    BEFORE UPDATE ON astronaut
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- environmental metrics must not be attached to an astronaut
CREATE FUNCTION check_measurement_metric_type() RETURNS trigger AS $$
DECLARE
    mt  smallint;
    ast uuid;
BEGIN
    SELECT metric_type INTO mt FROM metric WHERE id = NEW.metric_id;
    SELECT astronaut_id INTO ast FROM measurement WHERE id = NEW.measurement_id;
    IF mt = 2 AND ast IS NOT NULL THEN
        RAISE EXCEPTION 'metric_type=2 (environmental) cannot reference astronaut %', ast
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_measurement_metric_type
    BEFORE INSERT OR UPDATE ON measurement_value
    FOR EACH ROW EXECUTE FUNCTION check_measurement_metric_type();

-- denormalized timestamp: always mirror the parent session
CREATE FUNCTION sync_measurement_value_timestamp() RETURNS trigger AS $$
BEGIN
    SELECT recorded_at_utc INTO NEW.recorded_at_utc
    FROM measurement WHERE id = NEW.measurement_id;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_measurement_value_timestamp
    BEFORE INSERT OR UPDATE ON measurement_value
    FOR EACH ROW EXECUTE FUNCTION sync_measurement_value_timestamp();

-- keep children in sync when a session timestamp is corrected
CREATE FUNCTION cascade_measurement_timestamp() RETURNS trigger AS $$
BEGIN
    IF NEW.recorded_at_utc IS DISTINCT FROM OLD.recorded_at_utc THEN
        UPDATE measurement_value
           SET recorded_at_utc = NEW.recorded_at_utc
         WHERE measurement_id = NEW.id;
    END IF;
    RETURN NULL;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_measurement_timestamp_cascade
    AFTER UPDATE OF recorded_at_utc ON measurement
    FOR EACH ROW EXECUTE FUNCTION cascade_measurement_timestamp();

-- a reading's unit must share the metric's dimension
CREATE FUNCTION check_measurement_value_dimension() RETURNS trigger AS $$
DECLARE
    read_dim text;
    base_dim text;
BEGIN
    IF NEW.unit_id IS NULL THEN
        RETURN NEW;
    END IF;
    SELECT dimension INTO read_dim FROM unit WHERE id = NEW.unit_id;
    SELECT u.dimension INTO base_dim
    FROM metric m
    JOIN unit u ON u.id = m.base_unit_id
    WHERE m.id = NEW.metric_id;
    IF base_dim IS NOT NULL AND read_dim IS DISTINCT FROM base_dim THEN
        RAISE EXCEPTION 'unit dimension "%" does not match metric dimension "%"', read_dim, base_dim
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END $$ LANGUAGE plpgsql;

CREATE TRIGGER trg_measurement_value_dimension
    BEFORE INSERT OR UPDATE ON measurement_value
    FOR EACH ROW EXECUTE FUNCTION check_measurement_value_dimension();

-- ----------------------------------------------------------------------------
-- Views
-- ----------------------------------------------------------------------------
CREATE VIEW v_astronaut_age AS
SELECT id,
       date_part('year', age(CURRENT_DATE, birth_date))::smallint AS age
FROM astronaut;

COMMIT;
