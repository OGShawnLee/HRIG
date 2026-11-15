-- name: ListSpaceHazards :many
SELECT
    id,
    code,
    name,
    summary,
    clinical_impact,
    related_metric_codes,
    source,
    display_order
FROM space_hazard
ORDER BY display_order ASC, code ASC;

-- name: SearchSpaceHazards :many
SELECT
    id,
    code,
    name,
    summary,
    clinical_impact,
    related_metric_codes,
    source,
    display_order
FROM space_hazard
WHERE (
    sqlc.arg('q')::text = ''
    OR code ILIKE '%' || sqlc.arg('q')::text || '%'
    OR name ILIKE '%' || sqlc.arg('q')::text || '%'
    OR summary ILIKE '%' || sqlc.arg('q')::text || '%'
    OR clinical_impact ILIKE '%' || sqlc.arg('q')::text || '%'
    OR array_to_string(related_metric_codes, ' ') ILIKE '%' || sqlc.arg('q')::text || '%'
)
ORDER BY display_order ASC, code ASC;

-- name: ListMedicalProcedures :many
SELECT
    id,
    code,
    title,
    category,
    body_system,
    summary,
    when_to_use,
    escalation,
    estimated_minutes,
    environment,
    equipment,
    steps,
    warnings,
    related_metric_codes,
    hazard_codes,
    keywords,
    source,
    is_active,
    display_order
FROM medical_procedure
WHERE is_active = true
ORDER BY display_order ASC, code ASC;

-- name: SearchMedicalProcedures :many
SELECT
    id,
    code,
    title,
    category,
    body_system,
    summary,
    when_to_use,
    escalation,
    estimated_minutes,
    environment,
    equipment,
    steps,
    warnings,
    related_metric_codes,
    hazard_codes,
    keywords,
    source,
    is_active,
    display_order
FROM medical_procedure
WHERE is_active = true
  AND (
      sqlc.arg('q')::text = ''
      OR code ILIKE '%' || sqlc.arg('q')::text || '%'
      OR title ILIKE '%' || sqlc.arg('q')::text || '%'
      OR summary ILIKE '%' || sqlc.arg('q')::text || '%'
      OR when_to_use ILIKE '%' || sqlc.arg('q')::text || '%'
      OR category ILIKE '%' || sqlc.arg('q')::text || '%'
      OR body_system ILIKE '%' || sqlc.arg('q')::text || '%'
      OR array_to_string(keywords, ' ') ILIKE '%' || sqlc.arg('q')::text || '%'
      OR array_to_string(related_metric_codes, ' ') ILIKE '%' || sqlc.arg('q')::text || '%'
      OR array_to_string(hazard_codes, ' ') ILIKE '%' || sqlc.arg('q')::text || '%'
  )
  AND (sqlc.arg('category')::text = '' OR category = sqlc.arg('category')::text)
  AND (sqlc.arg('metric_code')::text = '' OR sqlc.arg('metric_code')::text = ANY(related_metric_codes))
ORDER BY display_order ASC, code ASC;
