-- ============================================================================
-- HRIG · Demo data
-- 003_demo_seed.sql — three astronauts, three days and spacecraft telemetry
--
-- Run after 001_schema.sql and 002_seed.sql.
-- The IDs for demo rows are deterministic, so this file is safe to re-run.
-- ============================================================================

BEGIN;

-- --------------------------------------------------------------------------
-- Crew
-- --------------------------------------------------------------------------
INSERT INTO astronaut (
    id, full_name, birth_date, nationality_code, nationality_label, is_active
) VALUES
    ('11111111-1111-4111-8111-111111111111', 'Ana Torres',   '1988-04-12', 'MX', 'Mexico',  true),
    ('22222222-2222-4222-8222-222222222222', 'Liam Okafor',  '1991-09-27', 'NG', 'Nigeria', true),
    ('33333333-3333-4333-8333-333333333333', 'Sofia Petrov', '1986-01-08', 'RU', 'Russia',  true)
ON CONFLICT (id) DO UPDATE SET
    full_name         = EXCLUDED.full_name,
    birth_date        = EXCLUDED.birth_date,
    nationality_code  = EXCLUDED.nationality_code,
    nationality_label = EXCLUDED.nationality_label,
    is_active         = EXCLUDED.is_active;

-- A separate sensor makes the suit telemetry distinguishable from cabin data.
INSERT INTO device (code, name, kind)
VALUES ('suit_sensor_01', 'EVA suit environment sensor', 'environment_sensor')
ON CONFLICT (code) DO UPDATE SET
    name = EXCLUDED.name,
    kind = EXCLUDED.kind;

-- --------------------------------------------------------------------------
-- Measurement sessions: one physiological session per astronaut/day, plus
-- one cabin and one suit session per day.
-- --------------------------------------------------------------------------
WITH demo_days(day, recorded_at_utc) AS (
    VALUES
        (DATE '2026-11-13', TIMESTAMPTZ '2026-11-13 08:00:00+00'),
        (DATE '2026-11-14', TIMESTAMPTZ '2026-11-14 08:00:00+00'),
        (DATE '2026-11-15', TIMESTAMPTZ '2026-11-15 08:00:00+00')
), crew AS (
    SELECT id, full_name, row_number() OVER (ORDER BY full_name)::int AS crew_no
    FROM astronaut
    WHERE id IN (
        '11111111-1111-4111-8111-111111111111',
        '22222222-2222-4222-8222-222222222222',
        '33333333-3333-4333-8333-333333333333'
    )
), sessions AS (
    SELECT md5(format('hrig-demo:%s:%s:biomonitor', c.id, d.day))::uuid AS id,
           c.id AS astronaut_id,
           dev.id AS device_id,
           'iva'::environment_mode AS environment,
           d.recorded_at_utc,
           'Daily resting health check'::text AS note
    FROM crew c
    CROSS JOIN demo_days d
    CROSS JOIN (SELECT id FROM device WHERE code = 'biomonitor_01') dev
    UNION ALL
    SELECT md5(format('hrig-demo:cabin:%s', d.day))::uuid,
           NULL,
           dev.id,
           'iva'::environment_mode,
           d.recorded_at_utc,
           'Spacecraft cabin telemetry'::text
    FROM demo_days d
    CROSS JOIN (SELECT id FROM device WHERE code = 'cabin_sensor_01') dev
    UNION ALL
    SELECT md5(format('hrig-demo:suit:%s', d.day))::uuid,
           NULL,
           dev.id,
           'eva'::environment_mode,
           d.recorded_at_utc + INTERVAL '4 hours',
           'EVA suit telemetry'::text
    FROM demo_days d
    CROSS JOIN (SELECT id FROM device WHERE code = 'suit_sensor_01') dev
)
INSERT INTO measurement (id, astronaut_id, device_id, environment, recorded_at_utc, note)
SELECT id, astronaut_id, device_id, environment, recorded_at_utc, note
FROM sessions
ON CONFLICT (id) DO UPDATE SET
    astronaut_id    = EXCLUDED.astronaut_id,
    device_id       = EXCLUDED.device_id,
    environment     = EXCLUDED.environment,
    recorded_at_utc = EXCLUDED.recorded_at_utc,
    note            = EXCLUDED.note;

-- --------------------------------------------------------------------------
-- Measurement values. The metric and unit are resolved by code, never by
-- generated UUID, so the seed remains portable between databases.
-- --------------------------------------------------------------------------
WITH demo_days(day, recorded_at_utc) AS (
    VALUES
        (DATE '2026-11-13', TIMESTAMPTZ '2026-11-13 08:00:00+00'),
        (DATE '2026-11-14', TIMESTAMPTZ '2026-11-14 08:00:00+00'),
        (DATE '2026-11-15', TIMESTAMPTZ '2026-11-15 08:00:00+00')
), crew AS (
    SELECT id, row_number() OVER (ORDER BY full_name)::int AS crew_no
    FROM astronaut
    WHERE id IN (
        '11111111-1111-4111-8111-111111111111',
        '22222222-2222-4222-8222-222222222222',
        '33333333-3333-4333-8333-333333333333'
    )
), crew_readings AS (
    SELECT md5(format('hrig-demo:%s:%s:biomonitor', c.id, d.day))::uuid AS measurement_id,
           r.metric_code, r.value_numeric, r.value_scale
    FROM crew c
    CROSS JOIN demo_days d
    CROSS JOIN LATERAL (VALUES
        ('spo2',        (CASE c.crew_no WHEN 1 THEN 98.0 WHEN 2 THEN 97.0 ELSE 99.0 END - (d.day - DATE '2026-11-13') * 0.2)::numeric, NULL::smallint),
        ('resp_rate',   (CASE c.crew_no WHEN 1 THEN 15.0 WHEN 2 THEN 16.0 ELSE 14.0 END + (d.day - DATE '2026-11-13') * 0.3)::numeric, NULL::smallint),
        ('bp_systolic', (CASE c.crew_no WHEN 1 THEN 116.0 WHEN 2 THEN 121.0 ELSE 112.0 END + (d.day - DATE '2026-11-13'))::numeric, NULL::smallint),
        ('bp_diastolic',(CASE c.crew_no WHEN 1 THEN 76.0 WHEN 2 THEN 79.0 ELSE 72.0 END)::numeric, NULL::smallint),
        ('heart_rate',  (CASE c.crew_no WHEN 1 THEN 64.0 WHEN 2 THEN 69.0 ELSE 61.0 END + (d.day - DATE '2026-11-13'))::numeric, NULL::smallint),
        ('hrv_rmssd',   (CASE c.crew_no WHEN 1 THEN 52.0 WHEN 2 THEN 47.0 ELSE 58.0 END - (d.day - DATE '2026-11-13'))::numeric, NULL::smallint),
        ('core_temp',   (CASE c.crew_no WHEN 1 THEN 36.6 WHEN 2 THEN 36.7 ELSE 36.5 END)::numeric, NULL::smallint),
        ('radiation_rate', (0.18 + (d.day - DATE '2026-11-13') * 0.01)::numeric, NULL::smallint),
        ('radiation_dose_cum', (12.4 + (d.day - DATE '2026-11-13') * 0.35 + c.crew_no * 0.2)::numeric, NULL::smallint),
        ('sleep_hours', (7.2 - (d.day - DATE '2026-11-13') * 0.2 + c.crew_no * 0.1)::numeric, NULL::smallint),
        ('mood', NULL::numeric, (4 - ((d.day - DATE '2026-11-13') + c.crew_no) % 2)::smallint)
    ) AS r(metric_code, value_numeric, value_scale)
), cabin_readings AS (
    SELECT md5(format('hrig-demo:cabin:%s', d.day))::uuid,
           r.metric_code, r.value_numeric, NULL::smallint
    FROM demo_days d
    CROSS JOIN LATERAL (VALUES
        ('cabin_o2',       (20.9 - (d.day - DATE '2026-11-13') * 0.05)::numeric),
        ('cabin_ppco2',    (3.2 + (d.day - DATE '2026-11-13') * 0.1)::numeric),
        ('cabin_pressure', (101.3 + (d.day - DATE '2026-11-13') * 0.1)::numeric),
        ('cabin_temp',     (22.1 + (d.day - DATE '2026-11-13') * 0.2)::numeric)
    ) AS r(metric_code, value_numeric)
), suit_readings AS (
    SELECT md5(format('hrig-demo:suit:%s', d.day))::uuid,
           r.metric_code, r.value_numeric, NULL::smallint
    FROM demo_days d
    CROSS JOIN LATERAL (VALUES
        ('suit_pressure', (57.5 + (d.day - DATE '2026-11-13') * 0.1)::numeric),
        ('suit_temp',     (28.4 + (d.day - DATE '2026-11-13') * 0.3)::numeric)
    ) AS r(metric_code, value_numeric)
), readings AS (
    SELECT * FROM crew_readings
    UNION ALL SELECT * FROM cabin_readings
    UNION ALL SELECT * FROM suit_readings
)
INSERT INTO measurement_value (
    measurement_id, metric_id, unit_id, recorded_at_utc,
    value_numeric, value_scale
)
SELECT r.measurement_id,
       m.id,
       CASE WHEN m.value_type = 'scale' THEN u_score.id ELSE m.base_unit_id END,
       meas.recorded_at_utc,
       r.value_numeric,
       r.value_scale
FROM readings r
JOIN metric m ON m.code = r.metric_code
LEFT JOIN unit u_score ON u_score.code = '{score}'
JOIN measurement meas ON meas.id = r.measurement_id
ON CONFLICT (measurement_id, metric_id) DO UPDATE SET
    unit_id       = EXCLUDED.unit_id,
    recorded_at_utc = EXCLUDED.recorded_at_utc,
    value_numeric = EXCLUDED.value_numeric,
    value_scale   = EXCLUDED.value_scale;

COMMIT;
