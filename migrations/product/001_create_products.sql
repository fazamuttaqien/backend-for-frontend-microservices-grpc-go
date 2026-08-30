CREATE TABLE
  IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    name VARCHAR(150) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price NUMERIC(19,4) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
  );

CREATE INDEX IF NOT EXISTS idx_products_created_at ON products (created_at DESC);
