-- Sample Data for Frappuccino Coffee Shop Database (Fixed Version)

-- Insert sample orders with correct enum values
INSERT INTO orders (customer_name, status, total_amount, special_instructions) VALUES
('Alice Johnson', 'closed', 8.25, '{"notes": "Extra hot, no foam"}'),
('Bob Smith', 'ready', 5.50, '{}'),
('Carol Davis', 'preparing', 12.75, '{"allergies": ["milk"], "substitutions": ["oat milk"]}'),
('David Wilson', 'pending', 7.20, '{}'),
('Emma Brown', 'closed', 15.50, '{"loyalty_discount": 0.10}');

-- Insert order items using a simpler approach
INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_time, customizations)
SELECT 
    o.id as order_id,
    m.id as menu_item_id,
    CASE 
        WHEN o.customer_name = 'Alice Johnson' AND m.name = 'Classic Americano' THEN 1
        WHEN o.customer_name = 'Alice Johnson' AND m.name = 'Vanilla Latte' THEN 1
        WHEN o.customer_name = 'Bob Smith' AND m.name = 'Frappuccino Classic' THEN 1
        WHEN o.customer_name = 'Carol Davis' AND m.name = 'Cafe Latte' THEN 2
        WHEN o.customer_name = 'Carol Davis' AND m.name = 'Cappuccino' THEN 1
        WHEN o.customer_name = 'David Wilson' AND m.name = 'Classic Americano' THEN 2
        WHEN o.customer_name = 'Emma Brown' AND m.name = 'Vanilla Latte' THEN 2
        WHEN o.customer_name = 'Emma Brown' AND m.name = 'Cappuccino' THEN 1
        WHEN o.customer_name = 'Emma Brown' AND m.name = 'Classic Americano' THEN 1
    END as quantity,
    CASE 
        WHEN m.name = 'Classic Americano' THEN 3.50
        WHEN m.name = 'Vanilla Latte' THEN 4.75
        WHEN m.name = 'Frappuccino Classic' THEN 5.50
        WHEN m.name = 'Cafe Latte' THEN 4.25
        WHEN m.name = 'Cappuccino' THEN 4.00
    END as price_at_time,
    CASE 
        WHEN o.customer_name = 'Alice Johnson' AND m.name = 'Classic Americano' THEN '{"size": "large", "temperature": "extra_hot"}'
        WHEN o.customer_name = 'Alice Johnson' AND m.name = 'Vanilla Latte' THEN '{"size": "medium", "milk": "whole"}'
        WHEN o.customer_name = 'Bob Smith' AND m.name = 'Frappuccino Classic' THEN '{"size": "large", "whipped_cream": true}'
        WHEN o.customer_name = 'Carol Davis' AND m.name = 'Cafe Latte' THEN '{"size": "medium", "milk": "oat"}'
        WHEN o.customer_name = 'Carol Davis' AND m.name = 'Cappuccino' THEN '{"size": "small", "milk": "oat"}'
        WHEN o.customer_name = 'David Wilson' AND m.name = 'Classic Americano' THEN '{"size": "medium", "extra_shot": true}'
        WHEN o.customer_name = 'Emma Brown' AND m.name = 'Vanilla Latte' THEN '{"size": "large", "extra_vanilla": true}'
        WHEN o.customer_name = 'Emma Brown' AND m.name = 'Cappuccino' THEN '{"size": "medium"}'
        WHEN o.customer_name = 'Emma Brown' AND m.name = 'Classic Americano' THEN '{"size": "small"}'
        ELSE '{}'
    END::jsonb as customizations
FROM orders o
CROSS JOIN menu_items m
WHERE 
    (o.customer_name = 'Alice Johnson' AND m.name IN ('Classic Americano', 'Vanilla Latte')) OR
    (o.customer_name = 'Bob Smith' AND m.name = 'Frappuccino Classic') OR
    (o.customer_name = 'Carol Davis' AND m.name IN ('Cafe Latte', 'Cappuccino')) OR
    (o.customer_name = 'David Wilson' AND m.name = 'Classic Americano') OR
    (o.customer_name = 'Emma Brown' AND m.name IN ('Vanilla Latte', 'Cappuccino', 'Classic Americano'));

-- Insert some inventory transactions
INSERT INTO inventory_transactions (ingredient_id, transaction_type, quantity_change, quantity_before, quantity_after, reference_type, notes)
SELECT 
    i.id,
    'usage'::transaction_type,
    -2.500,
    i.quantity + 2.500,
    i.quantity,
    'order',
    'Coffee preparation for orders'
FROM inventory i 
WHERE i.name = 'Coffee Beans - Arabica'

UNION ALL

SELECT 
    i.id,
    'usage'::transaction_type,
    -5.000,
    i.quantity + 5.000,
    i.quantity,
    'order',
    'Milk used for lattes and cappuccinos'
FROM inventory i 
WHERE i.name = 'Whole Milk'

UNION ALL

SELECT 
    i.id,
    'purchase'::transaction_type,
    25.000,
    0.500,
    25.500,
    'supplier',
    'Weekly coffee bean delivery'
FROM inventory i 
WHERE i.name = 'Coffee Beans - Arabica';
