CREATE INDEX idx_fact_status_date 
ON fact_order_sales (order_status, created_at);

CREATE INDEX idx_fact_restaurant_date 
ON fact_order_sales (restaurant_id, created_at);

CREATE INDEX idx_fact_restaurant_analytics 
ON fact_order_sales (restaurant_id, restaurant_name) 
INCLUDE (total_order_amount, total_item_count);