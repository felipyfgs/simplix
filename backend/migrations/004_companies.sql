-- +goose Up
-- +goose StatementBegin

CREATE TABLE companies (
    id                UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name              VARCHAR(255) NOT NULL,
    domain            VARCHAR(255),
    phone             VARCHAR(50),
    website           VARCHAR(500),
    industry          VARCHAR(100),
    description       TEXT,
    custom_attributes JSONB        NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_companies_name ON companies(name);

ALTER TABLE contacts ADD COLUMN company_id UUID REFERENCES companies(id) ON DELETE SET NULL;
CREATE INDEX idx_contacts_company_id ON contacts(company_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

ALTER TABLE contacts DROP COLUMN IF EXISTS company_id;
DROP INDEX IF EXISTS idx_contacts_company_id;
DROP INDEX IF EXISTS idx_companies_name;
DROP TABLE IF EXISTS companies CASCADE;

-- +goose StatementEnd
