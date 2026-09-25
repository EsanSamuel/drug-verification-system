-- name: CreateVerificationLog :one
INSERT INTO verification_logs (drug_unit_id, serial_number, verified, result, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListVerificationLogs :many
SELECT vl.id, vl.drug_unit_id, vl.serial_number, vl.verified, vl.result,
       vl.ip_address, vl.verified_at
FROM verification_logs vl
WHERE
    (sqlc.narg('serial_number')::text IS NULL OR vl.serial_number = sqlc.narg('serial_number')::text)
    AND (sqlc.narg('result')::verification_result IS NULL OR vl.result = sqlc.narg('result')::verification_result)
    AND (sqlc.narg('from_date')::timestamptz IS NULL OR vl.verified_at >= sqlc.narg('from_date')::timestamptz)
    AND (sqlc.narg('to_date')::timestamptz IS NULL OR vl.verified_at <= sqlc.narg('to_date')::timestamptz)
ORDER BY vl.verified_at DESC
LIMIT $1 OFFSET $2;

-- name: CountVerificationLogs :one
SELECT count(*)
FROM verification_logs vl
WHERE
    (sqlc.narg('serial_number')::text IS NULL OR vl.serial_number = sqlc.narg('serial_number')::text)
    AND (sqlc.narg('result')::verification_result IS NULL OR vl.result = sqlc.narg('result')::verification_result)
    AND (sqlc.narg('from_date')::timestamptz IS NULL OR vl.verified_at >= sqlc.narg('from_date')::timestamptz)
    AND (sqlc.narg('to_date')::timestamptz IS NULL OR vl.verified_at <= sqlc.narg('to_date')::timestamptz);

-- name: ListVerificationLogsByDrug :many
SELECT vl.id, vl.drug_unit_id, vl.serial_number, vl.verified, vl.result,
       vl.ip_address, vl.verified_at
FROM verification_logs vl
JOIN drug_units du ON vl.drug_unit_id = du.id
WHERE du.drug_id = $1
    AND (sqlc.narg('result')::verification_result IS NULL OR vl.result = sqlc.narg('result')::verification_result)
    AND (sqlc.narg('from_date')::timestamptz IS NULL OR vl.verified_at >= sqlc.narg('from_date')::timestamptz)
    AND (sqlc.narg('to_date')::timestamptz IS NULL OR vl.verified_at <= sqlc.narg('to_date')::timestamptz)
ORDER BY vl.verified_at DESC
LIMIT $2 OFFSET $3;

-- name: CountVerificationLogsByDrug :one
SELECT count(*)
FROM verification_logs vl
JOIN drug_units du ON vl.drug_unit_id = du.id
WHERE du.drug_id = $1
    AND (sqlc.narg('result')::verification_result IS NULL OR vl.result = sqlc.narg('result')::verification_result)
    AND (sqlc.narg('from_date')::timestamptz IS NULL OR vl.verified_at >= sqlc.narg('from_date')::timestamptz)
    AND (sqlc.narg('to_date')::timestamptz IS NULL OR vl.verified_at <= sqlc.narg('to_date')::timestamptz);
