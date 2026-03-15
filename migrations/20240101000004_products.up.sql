-- +goose Up
CREATE TABLE products (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    platform    TEXT NOT NULL,
    product_url TEXT NOT NULL UNIQUE,
    product_id  TEXT NOT NULL,
    active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(platform, product_id)
);
CREATE INDEX idx_products_platform_active ON products(platform, active);

-- +goose Down
DROP TABLE IF EXISTS products;
