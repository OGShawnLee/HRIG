-- name: CreateMeasurement :one
INSERT INTO measurement (
    astronaut_id,
    device_id,
    environment,
    recorded_at_utc,
    note
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING id, astronaut_id, device_id, environment, recorded_at_utc, note;

-- name: CreateMeasurementValue :one
INSERT INTO measurement_value (
    measurement_id,
    metric_id,
    unit_id,
    recorded_at_utc,
    value_numeric,
    value_scale,
    value_text,
    status,
    note
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    COALESCE(sqlc.narg('status')::reading_status, 'normal'::reading_status),
    $8
)
RETURNING id, measurement_id, metric_id, unit_id, recorded_at_utc, value_numeric, value_scale, value_text, status, note;

-- name: CreateMediaAsset :one
INSERT INTO media_asset (
    measurement_id,
    kind,
    file_path,
    mime_type,
    findings,
    captured_at_utc
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING id, measurement_id, kind, file_path, mime_type, findings, captured_at_utc;

-- name: ListLatestReadingsByMetric :many
SELECT DISTINCT ON (mv.metric_id, m.astronaut_id)
    mv.id,
    mv.measurement_id,
    mv.metric_id,
    met.code AS metric_code,
    met.name AS metric_name,
    met.short_name AS metric_short_name,
    met.category AS metric_category,
    met.tier AS metric_tier,
    met.metric_type,
    met.value_type,
    met.display_unit_symbol,
    m.astronaut_id,
    a.full_name AS astronaut_name,
    m.device_id,
    d.code AS device_code,
    m.environment,
    mv.unit_id,
    u.code AS unit_code,
    u.symbol AS unit_symbol,
    mv.recorded_at_utc,
    mv.value_numeric,
    mv.value_scale,
    mv.value_text,
    mv.status,
    mv.note
FROM measurement_value mv
JOIN measurement m ON m.id = mv.measurement_id
JOIN metric met ON met.id = mv.metric_id
LEFT JOIN astronaut a ON a.id = m.astronaut_id
LEFT JOIN device d ON d.id = m.device_id
LEFT JOIN unit u ON u.id = mv.unit_id
WHERE met.is_active = true
  AND (sqlc.narg('astronaut_id')::uuid IS NULL OR m.astronaut_id = sqlc.narg('astronaut_id')::uuid OR m.astronaut_id IS NULL)
ORDER BY mv.metric_id, m.astronaut_id, mv.recorded_at_utc DESC, mv.id DESC;

-- name: ListTrendsByMetric :many
SELECT
    mv.id,
    mv.measurement_id,
    mv.metric_id,
    met.code AS metric_code,
    m.astronaut_id,
    m.device_id,
    m.environment,
    mv.unit_id,
    u.code AS unit_code,
    u.symbol AS unit_symbol,
    mv.recorded_at_utc,
    mv.value_numeric,
    mv.value_scale,
    mv.value_text,
    mv.status
FROM measurement_value mv
JOIN measurement m ON m.id = mv.measurement_id
JOIN metric met ON met.id = mv.metric_id
LEFT JOIN unit u ON u.id = mv.unit_id
WHERE met.code = sqlc.arg('metric_code')
  AND (sqlc.narg('astronaut_id')::uuid IS NULL OR m.astronaut_id = sqlc.narg('astronaut_id')::uuid)
  AND (sqlc.narg('since_utc')::timestamptz IS NULL OR mv.recorded_at_utc >= sqlc.narg('since_utc')::timestamptz)
ORDER BY mv.recorded_at_utc ASC;

-- name: ListHistory :many
SELECT
    m.id AS measurement_id,
    m.astronaut_id,
    a.full_name AS astronaut_name,
    m.device_id,
    d.code AS device_code,
    d.name AS device_name,
    m.environment,
    m.recorded_at_utc,
    m.note AS measurement_note,
    mv.id AS value_id,
    mv.metric_id,
    met.code AS metric_code,
    met.name AS metric_name,
    met.value_type,
    mv.unit_id,
    u.code AS unit_code,
    u.symbol AS unit_symbol,
    met.display_unit_symbol,
    mv.value_numeric,
    mv.value_scale,
    mv.value_text,
    mv.status,
    mv.note AS value_note,
    ma.id AS media_asset_id,
    ma.kind AS media_kind,
    ma.file_path AS media_file_path,
    ma.mime_type AS media_mime_type,
    ma.findings AS media_findings,
    ma.captured_at_utc AS media_captured_at_utc
FROM measurement m
JOIN measurement_value mv ON mv.measurement_id = m.id
JOIN metric met ON met.id = mv.metric_id
LEFT JOIN astronaut a ON a.id = m.astronaut_id
LEFT JOIN device d ON d.id = m.device_id
LEFT JOIN unit u ON u.id = mv.unit_id
LEFT JOIN media_asset ma ON ma.measurement_id = m.id
WHERE (sqlc.narg('astronaut_id')::uuid IS NULL OR m.astronaut_id = sqlc.narg('astronaut_id')::uuid)
ORDER BY m.recorded_at_utc DESC, m.id DESC, met.display_order ASC
LIMIT sqlc.arg('limit_rows') OFFSET sqlc.arg('offset_rows');
