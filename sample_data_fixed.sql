-- Sample Data for Frappuccino Coffee Shop Database (Fixed Version)

-- First, let's check if we have menu items, if not create some basic ones
INSERT INTO menu_items (name, description, price, category, available, metadata, tags, available_sizes)
SELECT * FROM (VALUES
    ('Classic Americano', 'Rich espresso with hot water', 3.50, 'hot_coffee', true, '{"prep_time": 120}'::jsonb, ARRAY['coffee', 'espresso'], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Vanilla Latte', 'Smooth espresso with steamed milk and vanilla', 4.75, 'hot_coffee', true, '{"prep_time": 180}'::jsonb, ARRAY['coffee', 'latte', 'vanilla'], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Frappuccino Classic', 'Blended coffee with ice and cream', 5.50, 'cold_coffee', true, '{"prep_time": 240}'::jsonb, ARRAY['coffee', 'cold', 'blended'], ARRAY['medium', 'large']::item_size[]),
    ('Cafe Latte', 'Classic espresso with steamed milk', 4.25, 'hot_coffee', true, '{"prep_time": 150}'::jsonb, ARRAY['coffee', 'latte'], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Cappuccino', 'Espresso with steamed milk foam', 4.00, 'hot_coffee', true, '{"prep_time": 180}'::jsonb, ARRAY['coffee', 'cappuccino'], ARRAY['small', 'medium']::item_size[])
) as new_items(name, description, price, category, available, metadata, tags, available_sizes)
WHERE NOT EXISTS (SELECT 1 FROM menu_items WHERE menu_items.name = new_items.name);

-- Insert sample orders with correct enum values
INSERT INTO orders (customer_name, status, total_amount, special_instructions) VALUES
('Alice Johnson', 'closed', 8.25, '{"notes": "Extra hot, no foam"}'),
('Bob Smith', 'ready', 5.50, '{}'),
('Carol Davis', 'preparing', 12.75, '{"allergies": ["milk"], "substitutions": ["oat milk"]}'),
('David Wilson', 'pending', 7.20, '{}'),
('Emma Brown', 'closed', 15.50, '{"loyalty_discount": 0.10}');

-- Insert order items using a simpler approach with explicit menu item references
WITH order_data AS (
    SELECT o.id as order_id, o.customer_name, m.id as menu_item_id, m.name as menu_name, m.price
    FROM orders o
    CROSS JOIN menu_items m
    WHERE 
        (o.customer_name = 'Alice Johnson' AND m.name IN ('Classic Americano', 'Vanilla Latte')) OR
        (o.customer_name = 'Bob Smith' AND m.name = 'Frappuccino Classic') OR
        (o.customer_name = 'Carol Davis' AND m.name IN ('Cafe Latte', 'Cappuccino')) OR
        (o.customer_name = 'David Wilson' AND m.name = 'Classic Americano') OR
        (o.customer_name = 'Emma Brown' AND m.name IN ('Vanilla Latte', 'Cappuccino', 'Classic Americano'))
)
INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_time, customizations)
SELECT 
    order_id,
    menu_item_id,
    CASE 
        WHEN customer_name = 'Alice Johnson' THEN 1
        WHEN customer_name = 'Bob Smith' THEN 1
        WHEN customer_name = 'Carol Davis' AND menu_name = 'Cafe Latte' THEN 2
        WHEN customer_name = 'Carol Davis' AND menu_name = 'Cappuccino' THEN 1
        WHEN customer_name = 'David Wilson' THEN 2
        WHEN customer_name = 'Emma Brown' AND menu_name = 'Vanilla Latte' THEN 2
        WHEN customer_name = 'Emma Brown' AND menu_name = 'Cappuccino' THEN 1
        WHEN customer_name = 'Emma Brown' AND menu_name = 'Classic Americano' THEN 1
        ELSE 1
    END as quantity,
    price as price_at_time,
    CASE 
        WHEN customer_name = 'Alice Johnson' AND menu_name = 'Classic Americano' THEN '{"size": "large", "temperature": "extra_hot"}'
        WHEN customer_name = 'Alice Johnson' AND menu_name = 'Vanilla Latte' THEN '{"size": "medium", "milk": "whole"}'
        WHEN customer_name = 'Bob Smith' THEN '{"size": "large", "whipped_cream": true}'
        WHEN customer_name = 'Carol Davis' THEN '{"size": "medium", "milk": "oat"}'
        WHEN customer_name = 'David Wilson' THEN '{"size": "medium", "extra_shot": true}'
        WHEN customer_name = 'Emma Brown' AND menu_name = 'Vanilla Latte' THEN '{"size": "large", "extra_vanilla": true}'
        WHEN customer_name = 'Emma Brown' AND menu_name = 'Cappuccino' THEN '{"size": "medium"}'
        WHEN customer_name = 'Emma Brown' AND menu_name = 'Classic Americano' THEN '{"size": "small"}'
        ELSE '{}'
    END::jsonb as customizations
FROM order_data;

-- Add some basic inventory items if they don't exist
INSERT INTO inventory (name, quantity, unit, min_threshold, cost_per_unit)
SELECT * FROM (VALUES
    ('Coffee Beans - Arabica', 25.500, 'kg'::unit_type, 5.000, 12.50),
    ('Whole Milk', 15.000, 'liters'::unit_type, 3.000, 1.25),
    ('Vanilla Syrup', 5.000, 'liters'::unit_type, 1.000, 8.75)
) as new_inventory(name, quantity, unit, min_threshold, cost_per_unit)
WHERE NOT EXISTS (SELECT 1 FROM inventory WHERE inventory.name = new_inventory.name);

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
WHERE i.name = 'Coffee Beans - Arabica' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Coffee Beans - Arabica')

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
WHERE i.name = 'Whole Milk' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Whole Milk')

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
WHERE i.name = 'Coffee Beans - Arabica' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Coffee Beans - Arabica');
