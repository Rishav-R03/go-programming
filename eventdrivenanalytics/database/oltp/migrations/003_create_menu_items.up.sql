CREATE TABLE menu_items (

    item_id SERIAL PRIMARY KEY,

    restaurant_id INT NOT NULL,

    item_name VARCHAR(100) NOT NULL,

    price NUMERIC(10,2) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_menu_restaurant
    FOREIGN KEY (restaurant_id)
    REFERENCES restaurants(restaurant_id)
    ON DELETE CASCADE
);