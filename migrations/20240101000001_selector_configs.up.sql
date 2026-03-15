-- +goose Up
CREATE TABLE selector_configs (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform     TEXT NOT NULL,
    field        TEXT NOT NULL,
    rules        JSONB NOT NULL,
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    failure_rate NUMERIC(5,4) DEFAULT 0,
    last_success TIMESTAMPTZ,
    last_failure TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(platform, field)
);
CREATE INDEX idx_selector_configs_platform_active ON selector_configs(platform, active);

-- +goose Down
DROP TABLE IF EXISTS selector_configs;
