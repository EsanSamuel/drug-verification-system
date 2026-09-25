-- name: CreateDrug :one
INSERT INTO drugs (
    name, generic_name, manufacturer_id, batch_number,
    nafdac_number, manufacturing_date, expiry_date,
    quantity, status, created_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetDrug :one
SELECT d.*,
       m.name AS manufacturer_name
FROM drugs d
JOIN manufacturers m ON d.manufacturer_id = m.id
WHERE d.id = $1;

-- name: ListDrugs :many
SELECT d.*,
       m.name AS manufacturer_name
FROM drugs d
JOIN manufacturers m ON d.manufacturer_id = m.id
WHERE
    (sqlc.narg('name')::text IS NULL OR d.name ILIKE '%' || sqlc.narg('name')::text || '%')
    AND (sqlc.narg('generic_name')::text IS NULL OR d.generic_name ILIKE '%' || sqlc.narg('generic_name')::text || '%')
    AND (sqlc.narg('batch_number')::text IS NULL OR d.batch_number = sqlc.narg('batch_number')::text)
    AND (sqlc.narg('status')::drug_status IS NULL OR d.status = sqlc.narg('status')::drug_status)
    AND (sqlc.narg('manufacturer_id')::uuid IS NULL OR d.manufacturer_id = sqlc.narg('manufacturer_id')::uuid)
    AND (sqlc.narg('expiry_before')::date IS NULL OR d.expiry_date <= sqlc.narg('expiry_before')::date)
    AND (sqlc.narg('expiry_after')::date IS NULL OR d.expiry_date >= sqlc.narg('expiry_after')::date)
ORDER BY d.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountDrugs :one
SELECT count(*)
FROM drugs d
WHERE
    (sqlc.narg('name')::text IS NULL OR d.name ILIKE '%' || sqlc.narg('name')::text || '%')
    AND (sqlc.narg('generic_name')::text IS NULL OR d.generic_name ILIKE '%' || sqlc.narg('generic_name')::text || '%')
    AND (sqlc.narg('batch_number')::text IS NULL OR d.batch_number = sqlc.narg('batch_number')::text)
    AND (sqlc.narg('status')::drug_status IS NULL OR d.status = sqlc.narg('status')::drug_status)
    AND (sqlc.narg('manufacturer_id')::uuid IS NULL OR d.manufacturer_id = sqlc.narg('manufacturer_id')::uuid)
    AND (sqlc.narg('expiry_before')::date IS NULL OR d.expiry_date <= sqlc.narg('expiry_before')::date)
    AND (sqlc.narg('expiry_after')::date IS NULL OR d.expiry_date >= sqlc.narg('expiry_after')::date);

-- name: UpdateDrug :one
UPDATE drugs
SET
    name = $2,
    generic_name = $3,
    batch_number = $4,
    nafdac_number = $5,
    manufacturing_date = $6,
    expiry_date = $7,
    quantity = $8,
    status = $9
WHERE id = $1
RETURNING *;

-- name: UpdateDrugStatus :one
UPDATE drugs
SET status = $2
WHERE id = $1
RETURNING *;

-- name: GetDrugByManufacturerAndBatch :one
SELECT id FROM drugs
WHERE manufacturer_id = $1 AND batch_number = $2;

-- name: GetExpiredActiveDrugs :many
SELECT id, name, batch_number, expiry_date
FROM drugs
WHERE status = 'active' AND expiry_date < CURRENT_DATE
LIMIT 1000;
