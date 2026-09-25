-- name: CreateDrugUnit :one
INSERT INTO drug_units (drug_id, serial_number, status)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetDrugUnitBySerial :one
SELECT du.*,
       d.name AS drug_name,
       d.generic_name AS drug_generic_name,
       d.batch_number AS drug_batch_number,
       d.nafdac_number AS drug_nafdac_number,
       d.manufacturing_date AS drug_manufacturing_date,
       d.expiry_date AS drug_expiry_date,
       d.status AS drug_status,
       m.name AS manufacturer_name
FROM drug_units du
JOIN drugs d ON du.drug_id = d.id
JOIN manufacturers m ON d.manufacturer_id = m.id
WHERE du.serial_number = $1;

-- name: ListDrugUnitsByDrug :many
SELECT id, drug_id, serial_number, status, created_at, updated_at
FROM drug_units
WHERE drug_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountDrugUnitsByDrug :one
SELECT count(*)
FROM drug_units
WHERE drug_id = $1;

-- name: UpdateDrugUnitStatus :one
UPDATE drug_units
SET status = $2
WHERE id = $1
RETURNING *;

-- name: GetDrugUnit :one
SELECT id, drug_id, serial_number, status, created_at, updated_at
FROM drug_units
WHERE id = $1;
