CREATE TABLE order_items (

    order_item_id BIGSERIAL PRIMARY KEY,

    order_id BIGINT NOT NULL,

    item_id INT NOT NULL,

    quantity INT NOT NULL CHECK(quantity > 0),

    price NUMERIC(10,2) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_order_items_order
        FOREIGN KEY (order_id)
        REFERENCES orders(order_id)
        ON DELETE CASCADE,

    CONSTRAINT fk_order_items_item
        FOREIGN KEY (item_id)
        REFERENCES menu_items(item_id)
);