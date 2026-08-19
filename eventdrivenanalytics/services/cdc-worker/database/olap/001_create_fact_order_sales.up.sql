CREATE TABLE fact_order_sales (

    fact_id BIGSERIAL PRIMARY KEY,

    order_id BIGINT UNIQUE NOT NULL,

    date_key INTEGER NOT NULL,

    restaurant_id INTEGER NOT NULL,

    customer_id INTEGER NOT NULL,

    restaurant_name VARCHAR(100) NOT NULL,

    restaurant_city VARCHAR(100) NOT NULL,

    customer_name VARCHAR(100) NOT NULL,

    total_item_count INTEGER NOT NULL,

    total_order_amount NUMERIC(12,2) NOT NULL,

    order_status VARCHAR(30) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_fact_restaurant
ON fact_order_sales(restaurant_id);

CREATE INDEX idx_fact_date
ON fact_order_sales(date_key);

CREATE INDEX idx_fact_restaurant_date
ON fact_order_sales(
    restaurant_id,
    date_key
);