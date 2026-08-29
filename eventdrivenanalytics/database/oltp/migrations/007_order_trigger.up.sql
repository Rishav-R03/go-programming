CREATE TRIGGER trigger_order_created

AFTER INSERT
ON orders

FOR EACH ROW

EXECUTE FUNCTION notify_order_created();