-- name: ListActiveAstronauts :many
SELECT
    a.id,
    a.full_name,
    a.birth_date,
    v.age,
    a.nationality_code,
    a.nationality_label,
    a.is_active,
    a.created_at,
    a.updated_at
FROM astronaut a
JOIN v_astronaut_age v ON v.id = a.id
WHERE a.is_active = true
ORDER BY a.full_name ASC;

-- name: GetAstronautByID :one
SELECT
    a.id,
    a.full_name,
    a.birth_date,
    v.age,
    a.nationality_code,
    a.nationality_label,
    a.is_active,
    a.created_at,
    a.updated_at
FROM astronaut a
JOIN v_astronaut_age v ON v.id = a.id
WHERE a.id = $1;

-- name: ListActiveAstronautsByFullName :many
SELECT
    id,
    full_name,
    birth_date,
    nationality_code,
    nationality_label,
    is_active,
    created_at,
    updated_at
FROM astronaut
WHERE full_name = $1
  AND is_active = true
ORDER BY created_at ASC, id ASC;
