INSERT INTO restaurants(name, city)
VALUES
('Burger House', 'Bangalore'),
('Pizza World', 'Mumbai');

INSERT INTO customers(name, email)
VALUES
('John Doe', 'john@example.com'),
('Jane Doe', 'jane@example.com');

INSERT INTO menu_items(
    restaurant_id,
    item_name,
    price
)
VALUES
(1, 'Cheese Burger', 199),
(1, 'French Fries', 99),
(2, 'Farmhouse Pizza', 499);