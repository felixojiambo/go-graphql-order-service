-- migrations/001_create_schema.up.sql

-- Enable uuid-ossp or pgcrypto for gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Customers
CREATE TABLE customers (
                           id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           name         TEXT NOT NULL,
                           email        TEXT NOT NULL UNIQUE,
                           created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                           updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Categories (self-referential for hierarchy)
CREATE TABLE categories (
                            id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            name         TEXT NOT NULL,
                            parent_id    UUID REFERENCES categories(id) ON DELETE SET NULL,
                            created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                            updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Products
CREATE TABLE products (
                          id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          name         TEXT NOT NULL,
                          description  TEXT,
                          price        NUMERIC(10,2) NOT NULL CHECK (price >= 0),
                          category_id  UUID NOT NULL REFERENCES categories(id),
                          created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                          updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_products_category ON products(category_id);

-- Orders
CREATE TABLE orders (
                        id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        customer_id   UUID NOT NULL REFERENCES customers(id),
                        total_amount  NUMERIC(12,2) NOT NULL CHECK (total_amount >= 0),
                        status        TEXT NOT NULL DEFAULT 'pending',
                        created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                        updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_orders_customer ON orders(customer_id);

-- Order Items
CREATE TABLE order_items (
                             id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             order_id     UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
                             product_id   UUID NOT NULL REFERENCES products(id),
                             quantity     INT NOT NULL CHECK (quantity > 0),
                             unit_price   NUMERIC(10,2) NOT NULL CHECK (unit_price >= 0),
                             created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_order_items_order ON order_items(order_id);
