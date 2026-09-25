-- 000001_init.down.sql
-- Reverse of the initial schema

DROP TRIGGER IF EXISTS set_drug_units_updated_at ON drug_units;
DROP TRIGGER IF EXISTS set_drugs_updated_at ON drugs;
DROP TRIGGER IF EXISTS set_manufacturers_updated_at ON manufacturers;
DROP TRIGGER IF EXISTS set_users_updated_at ON users;
DROP FUNCTION IF EXISTS trigger_set_updated_at();

DROP TABLE IF EXISTS verification_logs;
DROP TABLE IF EXISTS drug_units;
DROP TABLE IF EXISTS drugs;
DROP TABLE IF EXISTS manufacturers;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS verification_result;
DROP TYPE IF EXISTS drug_unit_status;
DROP TYPE IF EXISTS drug_status;
DROP TYPE IF EXISTS user_role;
