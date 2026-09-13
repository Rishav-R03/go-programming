-- Create Enum For Order Status
CREATE TYPE order_status AS ENUM('PENDING','COMPLETED','FAILED');
-- Products Table
CREATE TABLE products(
    id BIGSERIAL PRIMARY KEY,
    sku VARCHAR(64) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    stock INT NOT NULL CHECK(stock >=0),
    version INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Orders table
CREATE TABLE orders(
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL CHECK(quantity > 0),
    status order_status NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index for rapid filtering of orders by status and date
CREATE INDEX idx_orders_status_created ON orders (status,created_at DESC);

-- SEED
INSERT INTO products(sku,name,stock) VALUES ('DEV-KEYBOARD-01','Mechanical Keyword',100);
