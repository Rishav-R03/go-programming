CREATE TABLE restaurants (
    restaurant_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    city VARCHAR(50) NOT NULL, 
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
