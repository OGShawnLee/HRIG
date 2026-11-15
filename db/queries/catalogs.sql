-- name: ListUnits :many
SELECT
    id,
    code,
    symbol,
    dimension,
    decimals,
    is_active
FROM unit
ORDER BY code ASC;

-- name: GetActiveUnitByCode :one
SELECT
    id,
    code,
    symbol,
    dimension,
    decimals,
    is_active
FROM unit
WHERE code = $1
  AND is_active = true;

-- name: ListActiveMetrics :many
SELECT
    m.id,
    m.code,
    m.name,
    m.short_name,
    m.category,
    m.tier,
    m.metric_type,
    m.value_type,
    m.base_unit_id,
    u.code AS base_unit_code,
    u.symbol AS base_unit_symbol,
    u.dimension AS base_unit_dimension,
    u.decimals AS base_unit_decimals,
    m.display_unit_symbol,
    m.scale_min,
    m.scale_max,
    m.display_order,
    m.is_active,
    mt.warning_min,
    mt.warning_max,
    mt.critical_min,
    mt.critical_max,
    mt.classification_mode
FROM metric m
LEFT JOIN unit u ON u.id = m.base_unit_id
LEFT JOIN metric_threshold mt ON mt.metric_id = m.id
WHERE m.is_active = true
ORDER BY m.display_order ASC, m.code ASC;

-- name: GetActiveMetricByCode :one
SELECT
    m.id,
    m.code,
    m.name,
    m.short_name,
    m.category,
    m.tier,
    m.metric_type,
    m.value_type,
    m.base_unit_id,
    u.code AS base_unit_code,
    u.symbol AS base_unit_symbol,
    u.dimension AS base_unit_dimension,
    u.decimals AS base_unit_decimals,
    m.display_unit_symbol,
    m.scale_min,
    m.scale_max,
    m.display_order,
    m.is_active,
    mt.warning_min,
    mt.warning_max,
    mt.critical_min,
    mt.critical_max,
    mt.classification_mode
FROM metric m
LEFT JOIN unit u ON u.id = m.base_unit_id
LEFT JOIN metric_threshold mt ON mt.metric_id = m.id
WHERE m.code = $1
  AND m.is_active = true;

-- name: GetThresholdByMetricCode :one
SELECT
    mt.id,
    mt.metric_id,
    m.code AS metric_code,
    mt.warning_min,
    mt.warning_max,
    mt.critical_min,
    mt.critical_max,
    mt.classification_mode,
    mt.source,
    mt.updated_at
FROM metric_threshold mt
JOIN metric m ON m.id = mt.metric_id
WHERE m.code = $1;

-- name: ListActiveDevices :many
SELECT
    id,
    code,
    name,
    kind,
    is_active
FROM device
WHERE is_active = true
ORDER BY code ASC;

-- name: GetActiveDeviceByCode :one
SELECT
    id,
    code,
    name,
    kind,
    is_active
FROM device
WHERE code = $1
  AND is_active = true;
