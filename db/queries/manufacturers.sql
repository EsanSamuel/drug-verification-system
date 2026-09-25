-- name: CreateManufacturer :one
INSERT INTO manufacturers (name, address)
VALUES ($1, $2)
RETURNING id, name, address, created_at, updated_at;

-- name: GetManufacturer :one
SELECT id, name, address, created_at, updated_at
FROM manufacturers
WHERE id = $1;

-- name: ListManufacturers :many
SELECT id, name, address, created_at, updated_at
FROM manufacturers
ORDER BY name ASC
LIMIT $1 OFFSET $2;

-- name: CountManufacturers :one
SELECT count(*) FROM manufacturers;

-- name: UpdateManufacturer :one
UPDATE manufacturers
SET name = $2, address = $3
WHERE id = $1
RETURNING id, name, address, created_at, updated_at;

-- name: DeleteManufacturer :exec
DELETE FROM manufacturers WHERE id = $1;

-- name: GetManufacturerByName :one
SELECT id, name, address, created_at, updated_at
FROM manufacturers
WHERE name = $1;
