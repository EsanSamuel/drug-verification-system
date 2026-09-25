-- 000001_init.up.sql
-- Drug Verification System — Initial Schema

-- Custom enum types
CREATE TYPE user_role AS ENUM ('pharmacist', 'admin');
CREATE TYPE drug_status AS ENUM ('active', 'expired', 'recalled', 'suspended');
CREATE TYPE drug_unit_status AS ENUM ('active', 'used', 'recalled', 'suspended');
CREATE TYPE verification_result AS ENUM ('verified', 'not_found', 'expired', 'recalled', 'suspended', 'invalid');

-- =============================================================================
-- Users
-- =============================================================================
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role user_role NOT NULL DEFAULT 'pharmacist',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_users_email ON users (email);

-- =============================================================================
-- Manufacturers
-- =============================================================================
CREATE TABLE manufacturers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    address TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_manufacturers_name ON manufacturers (name);

-- =============================================================================
-- Drugs
-- =============================================================================
CREATE TABLE drugs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    generic_name TEXT NOT NULL DEFAULT '',
    manufacturer_id UUID NOT NULL REFERENCES manufacturers(id) ON DELETE RESTRICT,
    batch_number TEXT NOT NULL,
    nafdac_number TEXT NOT NULL DEFAULT '',
    manufacturing_date DATE NOT NULL,
    expiry_date DATE NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    status drug_status NOT NULL DEFAULT 'active',
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_drugs_dates CHECK (manufacturing_date <= expiry_date)
);

-- Prevent duplicate batches from the same manufacturer
CREATE UNIQUE INDEX idx_drugs_manufacturer_batch ON drugs (manufacturer_id, batch_number);
CREATE INDEX idx_drugs_name ON drugs (name);
CREATE INDEX idx_drugs_generic_name ON drugs (generic_name);
CREATE INDEX idx_drugs_status ON drugs (status);
CREATE INDEX idx_drugs_expiry_date ON drugs (expiry_date);
CREATE INDEX idx_drugs_nafdac_number ON drugs (nafdac_number);

-- =============================================================================
-- Drug Units
-- =============================================================================
CREATE TABLE drug_units (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drug_id UUID NOT NULL REFERENCES drugs(id) ON DELETE RESTRICT,
    serial_number TEXT NOT NULL,
    status drug_unit_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX idx_drug_units_serial ON drug_units (serial_number);
CREATE INDEX idx_drug_units_drug_id ON drug_units (drug_id);

-- =============================================================================
-- Verification Logs
-- =============================================================================
CREATE TABLE verification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    drug_unit_id UUID REFERENCES drug_units(id) ON DELETE SET NULL,
    serial_number TEXT NOT NULL,
    verified BOOLEAN NOT NULL,
    result verification_result NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    verified_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_verification_logs_serial ON verification_logs (serial_number);
CREATE INDEX idx_verification_logs_drug_unit ON verification_logs (drug_unit_id);
CREATE INDEX idx_verification_logs_verified_at ON verification_logs (verified_at);
CREATE INDEX idx_verification_logs_result ON verification_logs (result);

-- =============================================================================
-- Updated-at trigger function
-- =============================================================================
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply to all tables with updated_at
CREATE TRIGGER set_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_manufacturers_updated_at
    BEFORE UPDATE ON manufacturers
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_drugs_updated_at
    BEFORE UPDATE ON drugs
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER set_drug_units_updated_at
    BEFORE UPDATE ON drug_units
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();
