CREATE OR REPLACE FUNCTION notify_order_created()
RETURNS trigger
AS $$
DECLARE
    payload JSON;
BEGIN

    SELECT json_build_object(
        'order_id', NEW.order_id,
        'customer_id', NEW.customer_id,
        'restaurant_id', NEW.restaurant_id,
        'status', NEW.status,
        'created_at', NEW.created_at
    )
    INTO payload;

    PERFORM pg_notify(
        'order_events',
        payload::TEXT
    );

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;