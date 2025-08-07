-- Sample Data for Frappuccino Coffee Shop Database (Expanded Version)

-- First, let's check if we have menu items, if not create some basic ones
INSERT INTO menu_items (name, description, price, category, available, metadata, tags, allergens, available_sizes)
SELECT * FROM (VALUES
    ('Classic Americano', 'Rich espresso with hot water', 3.50, 'hot_coffee', true, '{"prep_time": 120, "caffeine_level": "high"}'::jsonb, ARRAY['coffee', 'espresso'], ARRAY[]::text[], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Vanilla Latte', 'Smooth espresso with steamed milk and vanilla', 4.75, 'hot_coffee', true, '{"prep_time": 180, "caffeine_level": "medium"}'::jsonb, ARRAY['coffee', 'latte', 'vanilla'], ARRAY['milk'], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Frappuccino Classic', 'Blended coffee with ice and cream', 5.50, 'cold_coffee', true, '{"prep_time": 240, "caffeine_level": "medium"}'::jsonb, ARRAY['coffee', 'cold', 'blended'], ARRAY['milk'], ARRAY['medium', 'large']::item_size[]),
    ('Cafe Latte', 'Classic espresso with steamed milk', 4.25, 'hot_coffee', true, '{"prep_time": 150, "caffeine_level": "medium"}'::jsonb, ARRAY['coffee', 'latte'], ARRAY['milk'], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Cappuccino', 'Espresso with steamed milk foam', 4.00, 'hot_coffee', true, '{"prep_time": 180, "caffeine_level": "high"}'::jsonb, ARRAY['coffee', 'cappuccino'], ARRAY['milk'], ARRAY['small', 'medium']::item_size[]),
    ('Caramel Macchiato', 'Vanilla syrup, steamed milk, espresso, caramel drizzle', 5.25, 'hot_coffee', true, '{"prep_time": 200, "caffeine_level": "medium"}'::jsonb, ARRAY['coffee', 'caramel', 'sweet'], ARRAY['milk'], ARRAY['medium', 'large']::item_size[]),
    ('Mocha', 'Rich chocolate and espresso with steamed milk', 4.95, 'hot_coffee', true, '{"prep_time": 190, "caffeine_level": "medium"}'::jsonb, ARRAY['coffee', 'chocolate', 'sweet'], ARRAY['milk'], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Iced Coffee', 'Cold brew coffee served over ice', 2.95, 'cold_coffee', true, '{"prep_time": 60, "caffeine_level": "high"}'::jsonb, ARRAY['coffee', 'cold', 'refreshing'], ARRAY[]::text[], ARRAY['medium', 'large']::item_size[]),
    ('Green Tea Latte', 'Matcha green tea with steamed milk', 4.50, 'tea', true, '{"prep_time": 160, "caffeine_level": "low"}'::jsonb, ARRAY['tea', 'matcha', 'healthy'], ARRAY['milk'], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Chai Tea Latte', 'Spiced tea blend with steamed milk', 4.25, 'tea', true, '{"prep_time": 170, "caffeine_level": "low"}'::jsonb, ARRAY['tea', 'spiced', 'warming'], ARRAY['milk'], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Hot Chocolate', 'Rich chocolate drink with steamed milk', 3.75, 'non_coffee', true, '{"prep_time": 140, "caffeine_level": "none"}'::jsonb, ARRAY['chocolate', 'sweet', 'comfort'], ARRAY['milk'], ARRAY['small', 'medium', 'large']::item_size[]),
    ('Blueberry Muffin', 'Fresh baked muffin with real blueberries', 2.85, 'pastry', true, '{"prep_time": 30, "contains_gluten": true}'::jsonb, ARRAY['bakery', 'sweet', 'breakfast'], ARRAY['gluten', 'eggs'], ARRAY['medium']::item_size[]),
    ('Croissant', 'Buttery, flaky French pastry', 3.25, 'pastry', true, '{"prep_time": 45, "contains_gluten": true}'::jsonb, ARRAY['bakery', 'buttery', 'breakfast'], ARRAY['gluten', 'butter'], ARRAY['medium']::item_size[]),
    ('Avocado Toast', 'Multigrain bread with fresh avocado spread', 6.50, 'food', true, '{"prep_time": 300, "healthy": true}'::jsonb, ARRAY['healthy', 'breakfast', 'vegetarian'], ARRAY['gluten'], ARRAY['medium']::item_size[]),
    ('Turkey Sandwich', 'Sliced turkey with lettuce, tomato on sourdough', 8.95, 'food', true, '{"prep_time": 360, "protein_rich": true}'::jsonb, ARRAY['lunch', 'protein', 'savory'], ARRAY['gluten'], ARRAY['medium']::item_size[])
) as new_items(name, description, price, category, available, metadata, tags, allergens, available_sizes)
WHERE NOT EXISTS (SELECT 1 FROM menu_items WHERE menu_items.name = new_items.name);

-- Add comprehensive inventory items
INSERT INTO inventory (name, quantity, unit, min_threshold, cost_per_unit)
SELECT * FROM (VALUES
    ('Coffee Beans - Arabica', 25.500, 'kg'::unit_type, 5.000, 12.50),
    ('Coffee Beans - Robusta', 15.000, 'kg'::unit_type, 3.000, 10.75),
    ('Whole Milk', 45.000, 'liters'::unit_type, 8.000, 1.25),
    ('Oat Milk', 20.000, 'liters'::unit_type, 4.000, 2.85),
    ('Almond Milk', 15.000, 'liters'::unit_type, 3.000, 3.25),
    ('Vanilla Syrup', 12.000, 'liters'::unit_type, 2.000, 8.75),
    ('Caramel Syrup', 8.000, 'liters'::unit_type, 2.000, 9.25),
    ('Chocolate Syrup', 10.000, 'liters'::unit_type, 2.500, 7.95),
    ('Matcha Powder', 2.500, 'kg'::unit_type, 0.500, 45.00),
    ('Chai Tea Blend', 5.000, 'kg'::unit_type, 1.000, 18.50),
    ('Whipped Cream', 8.000, 'liters'::unit_type, 2.000, 4.50),
    ('Sugar', 50.000, 'kg'::unit_type, 10.000, 2.25),
    ('Cinnamon', 1.000, 'kg'::unit_type, 0.200, 12.00),
    ('Cocoa Powder', 3.000, 'kg'::unit_type, 0.750, 15.25),
    ('Paper Cups - Small', 500.000, 'pieces'::unit_type, 100.000, 0.12),
    ('Paper Cups - Medium', 750.000, 'pieces'::unit_type, 150.000, 0.15),
    ('Paper Cups - Large', 400.000, 'pieces'::unit_type, 80.000, 0.18),
    ('Plastic Lids', 800.000, 'pieces'::unit_type, 200.000, 0.08),
    ('Napkins', 2000.000, 'pieces'::unit_type, 500.000, 0.02),
    ('Stirring Sticks', 1500.000, 'pieces'::unit_type, 300.000, 0.01)
) as new_inventory(name, quantity, unit, min_threshold, cost_per_unit)
WHERE NOT EXISTS (SELECT 1 FROM inventory WHERE inventory.name = new_inventory.name);

-- Insert expanded sample orders
INSERT INTO orders (customer_name, status, total_amount, special_instructions) VALUES
('Alice Johnson', 'closed', 8.25, '{"notes": "Extra hot, no foam"}'),
('Bob Smith', 'ready', 5.50, '{}'),
('Carol Davis', 'preparing', 12.75, '{"allergies": ["milk"], "substitutions": ["oat milk"]}'),
('David Wilson', 'pending', 7.20, '{}'),
('Emma Brown', 'closed', 15.50, '{"loyalty_discount": 0.10}'),
('Frank Miller', 'closed', 18.95, '{"payment_method": "card", "tip": 3.00}'),
('Grace Lee', 'ready', 6.75, '{"pickup_time": "2025-08-07T09:30:00Z"}'),
('Henry Wong', 'preparing', 11.20, '{"decaf_requested": true}'),
('Iris Chen', 'pending', 4.50, '{"mobile_order": true}'),
('Jack Thompson', 'closed', 22.40, '{"office_delivery": true, "floor": 5}'),
('Kate Rodriguez', 'cancelled', 9.85, '{"cancellation_reason": "customer_request"}'),
('Luis Garcia', 'ready', 13.60, '{"extra_shot": true, "light_foam": true}'),
('Maya Patel', 'preparing', 16.25, '{"birthday_order": true, "special_message": "Happy Birthday!"}'),
('Noah Kim', 'pending', 7.95, '{"student_discount": 0.15}'),
('Olivia Taylor', 'closed', 19.75, '{"corporate_account": "TechCorp_001"}');

-- Insert comprehensive order items with more variety
WITH order_data AS (
    SELECT o.id as order_id, o.customer_name, m.id as menu_item_id, m.name as menu_name, m.price
    FROM orders o
    CROSS JOIN menu_items m
    WHERE 
        (o.customer_name = 'Alice Johnson' AND m.name IN ('Classic Americano', 'Vanilla Latte')) OR
        (o.customer_name = 'Bob Smith' AND m.name = 'Frappuccino Classic') OR
        (o.customer_name = 'Carol Davis' AND m.name IN ('Cafe Latte', 'Cappuccino')) OR
        (o.customer_name = 'David Wilson' AND m.name = 'Classic Americano') OR
        (o.customer_name = 'Emma Brown' AND m.name IN ('Vanilla Latte', 'Cappuccino', 'Classic Americano')) OR
        (o.customer_name = 'Frank Miller' AND m.name IN ('Caramel Macchiato', 'Mocha', 'Blueberry Muffin', 'Croissant')) OR
        (o.customer_name = 'Grace Lee' AND m.name IN ('Iced Coffee', 'Avocado Toast')) OR
        (o.customer_name = 'Henry Wong' AND m.name IN ('Green Tea Latte', 'Chai Tea Latte')) OR
        (o.customer_name = 'Iris Chen' AND m.name = 'Hot Chocolate') OR
        (o.customer_name = 'Jack Thompson' AND m.name IN ('Classic Americano', 'Cafe Latte', 'Turkey Sandwich', 'Croissant', 'Blueberry Muffin')) OR
        (o.customer_name = 'Kate Rodriguez' AND m.name IN ('Vanilla Latte', 'Mocha')) OR
        (o.customer_name = 'Luis Garcia' AND m.name IN ('Cappuccino', 'Caramel Macchiato')) OR
        (o.customer_name = 'Maya Patel' AND m.name IN ('Frappuccino Classic', 'Hot Chocolate', 'Blueberry Muffin')) OR
        (o.customer_name = 'Noah Kim' AND m.name IN ('Iced Coffee', 'Turkey Sandwich')) OR
        (o.customer_name = 'Olivia Taylor' AND m.name IN ('Mocha', 'Chai Tea Latte', 'Avocado Toast', 'Croissant'))
),
order_items_data AS (
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
        WHEN customer_name = 'Frank Miller' AND menu_name IN ('Blueberry Muffin', 'Croissant') THEN 1
        WHEN customer_name = 'Frank Miller' AND menu_name IN ('Caramel Macchiato', 'Mocha') THEN 2
        WHEN customer_name = 'Grace Lee' THEN 1
        WHEN customer_name = 'Henry Wong' THEN 1
        WHEN customer_name = 'Iris Chen' THEN 1
        WHEN customer_name = 'Jack Thompson' AND menu_name IN ('Classic Americano', 'Cafe Latte') THEN 2
        WHEN customer_name = 'Jack Thompson' AND menu_name = 'Turkey Sandwich' THEN 1
        WHEN customer_name = 'Jack Thompson' AND menu_name IN ('Croissant', 'Blueberry Muffin') THEN 1
        WHEN customer_name = 'Kate Rodriguez' THEN 1
        WHEN customer_name = 'Luis Garcia' THEN 2
        WHEN customer_name = 'Maya Patel' AND menu_name = 'Frappuccino Classic' THEN 1
        WHEN customer_name = 'Maya Patel' AND menu_name IN ('Hot Chocolate', 'Blueberry Muffin') THEN 2
        WHEN customer_name = 'Noah Kim' THEN 1
        WHEN customer_name = 'Olivia Taylor' AND menu_name IN ('Mocha', 'Chai Tea Latte') THEN 2
        WHEN customer_name = 'Olivia Taylor' AND menu_name IN ('Avocado Toast', 'Croissant') THEN 1
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
        WHEN customer_name = 'Frank Miller' AND menu_name = 'Caramel Macchiato' THEN '{"size": "large", "extra_caramel": true}'
        WHEN customer_name = 'Frank Miller' AND menu_name = 'Mocha' THEN '{"size": "medium", "whipped_cream": true}'
        WHEN customer_name = 'Grace Lee' AND menu_name = 'Iced Coffee' THEN '{"size": "large", "extra_ice": true}'
        WHEN customer_name = 'Grace Lee' AND menu_name = 'Avocado Toast' THEN '{"extra_avocado": true, "add_tomato": true}'
        WHEN customer_name = 'Henry Wong' THEN '{"size": "medium", "decaf": true}'
        WHEN customer_name = 'Iris Chen' THEN '{"size": "small", "marshmallows": true}'
        WHEN customer_name = 'Jack Thompson' AND menu_name IN ('Classic Americano', 'Cafe Latte') THEN '{"size": "large"}'
        WHEN customer_name = 'Kate Rodriguez' THEN '{"size": "medium"}'
        WHEN customer_name = 'Luis Garcia' THEN '{"size": "medium", "extra_shot": true, "light_foam": true}'
        WHEN customer_name = 'Maya Patel' AND menu_name = 'Frappuccino Classic' THEN '{"size": "large", "extra_whipped_cream": true, "birthday_special": true}'
        WHEN customer_name = 'Maya Patel' AND menu_name = 'Hot Chocolate' THEN '{"size": "medium", "extra_chocolate": true}'
        WHEN customer_name = 'Noah Kim' AND menu_name = 'Iced Coffee' THEN '{"size": "medium", "sugar_free": true}'
        WHEN customer_name = 'Olivia Taylor' AND menu_name = 'Mocha' THEN '{"size": "large", "oat_milk": true}'
        WHEN customer_name = 'Olivia Taylor' AND menu_name = 'Chai Tea Latte' THEN '{"size": "medium", "extra_spice": true}'
        ELSE '{}'
    END::jsonb as customizations
FROM order_data
)
INSERT INTO order_items (order_id, menu_item_id, quantity, price_at_time, customizations)
SELECT order_id, menu_item_id, quantity, price_at_time, customizations FROM order_items_data;

-- Add menu item ingredients relationships
INSERT INTO menu_item_ingredients (menu_item_id, ingredient_id, required_quantity, unit)
SELECT m.id, i.id, 
    CASE 
        WHEN m.name = 'Classic Americano' AND i.name = 'Coffee Beans - Arabica' THEN 0.025
        WHEN m.name = 'Vanilla Latte' AND i.name = 'Coffee Beans - Arabica' THEN 0.020
        WHEN m.name = 'Vanilla Latte' AND i.name = 'Whole Milk' THEN 0.200
        WHEN m.name = 'Vanilla Latte' AND i.name = 'Vanilla Syrup' THEN 0.030
        WHEN m.name = 'Cafe Latte' AND i.name = 'Coffee Beans - Arabica' THEN 0.020
        WHEN m.name = 'Cafe Latte' AND i.name = 'Whole Milk' THEN 0.180
        WHEN m.name = 'Cappuccino' AND i.name = 'Coffee Beans - Arabica' THEN 0.018
        WHEN m.name = 'Cappuccino' AND i.name = 'Whole Milk' THEN 0.150
        WHEN m.name = 'Caramel Macchiato' AND i.name = 'Coffee Beans - Arabica' THEN 0.020
        WHEN m.name = 'Caramel Macchiato' AND i.name = 'Whole Milk' THEN 0.190
        WHEN m.name = 'Caramel Macchiato' AND i.name = 'Vanilla Syrup' THEN 0.025
        WHEN m.name = 'Caramel Macchiato' AND i.name = 'Caramel Syrup' THEN 0.020
        WHEN m.name = 'Mocha' AND i.name = 'Coffee Beans - Arabica' THEN 0.020
        WHEN m.name = 'Mocha' AND i.name = 'Whole Milk' THEN 0.170
        WHEN m.name = 'Mocha' AND i.name = 'Chocolate Syrup' THEN 0.040
        WHEN m.name = 'Green Tea Latte' AND i.name = 'Matcha Powder' THEN 0.008
        WHEN m.name = 'Green Tea Latte' AND i.name = 'Whole Milk' THEN 0.200
        WHEN m.name = 'Chai Tea Latte' AND i.name = 'Chai Tea Blend' THEN 0.012
        WHEN m.name = 'Chai Tea Latte' AND i.name = 'Whole Milk' THEN 0.180
        WHEN m.name = 'Hot Chocolate' AND i.name = 'Cocoa Powder' THEN 0.025
        WHEN m.name = 'Hot Chocolate' AND i.name = 'Whole Milk' THEN 0.220
        WHEN m.name = 'Hot Chocolate' AND i.name = 'Sugar' THEN 0.015
    END as required_quantity,
    CASE 
        WHEN i.name LIKE '%Milk%' OR i.name LIKE '%Syrup%' THEN 'liters'::unit_type
        WHEN i.name LIKE '%Powder%' OR i.name LIKE '%Beans%' OR i.name LIKE '%Blend%' OR i.name = 'Sugar' THEN 'kg'::unit_type
        ELSE 'grams'::unit_type
    END as unit
FROM menu_items m
CROSS JOIN inventory i
WHERE (
    (m.name = 'Classic Americano' AND i.name = 'Coffee Beans - Arabica') OR
    (m.name = 'Vanilla Latte' AND i.name IN ('Coffee Beans - Arabica', 'Whole Milk', 'Vanilla Syrup')) OR
    (m.name = 'Cafe Latte' AND i.name IN ('Coffee Beans - Arabica', 'Whole Milk')) OR
    (m.name = 'Cappuccino' AND i.name IN ('Coffee Beans - Arabica', 'Whole Milk')) OR
    (m.name = 'Caramel Macchiato' AND i.name IN ('Coffee Beans - Arabica', 'Whole Milk', 'Vanilla Syrup', 'Caramel Syrup')) OR
    (m.name = 'Mocha' AND i.name IN ('Coffee Beans - Arabica', 'Whole Milk', 'Chocolate Syrup')) OR
    (m.name = 'Green Tea Latte' AND i.name IN ('Matcha Powder', 'Whole Milk')) OR
    (m.name = 'Chai Tea Latte' AND i.name IN ('Chai Tea Blend', 'Whole Milk')) OR
    (m.name = 'Hot Chocolate' AND i.name IN ('Cocoa Powder', 'Whole Milk', 'Sugar'))
)
AND NOT EXISTS (
    SELECT 1 FROM menu_item_ingredients mii 
    WHERE mii.menu_item_id = m.id AND mii.ingredient_id = i.id
);

-- Insert comprehensive inventory transactions
INSERT INTO inventory_transactions (ingredient_id, transaction_type, quantity_change, quantity_before, quantity_after, reference_type, reference_id, notes)
SELECT 
    i.id,
    'usage'::transaction_type,
    -2.500,
    i.quantity + 2.500,
    i.quantity,
    'order',
    o.id,
    'Coffee preparation for orders'
FROM inventory i 
CROSS JOIN orders o
WHERE i.name = 'Coffee Beans - Arabica' AND o.customer_name IN ('Alice Johnson', 'David Wilson', 'Emma Brown')
AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Coffee Beans - Arabica')

UNION ALL

SELECT 
    i.id,
    'usage'::transaction_type,
    -5.000,
    i.quantity + 5.000,
    i.quantity,
    'order',
    o.id,
    'Milk used for lattes and cappuccinos'
FROM inventory i 
CROSS JOIN orders o
WHERE i.name = 'Whole Milk' AND o.customer_name IN ('Carol Davis', 'Frank Miller', 'Grace Lee')
AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Whole Milk')

UNION ALL

SELECT 
    i.id,
    'purchase'::transaction_type,
    50.000,
    i.quantity - 50.000,
    i.quantity,
    'supplier',
    NULL,
    'Weekly coffee bean delivery from Premium Coffee Co'
FROM inventory i 
WHERE i.name = 'Coffee Beans - Arabica' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Coffee Beans - Arabica')

UNION ALL

SELECT 
    i.id,
    'purchase'::transaction_type,
    30.000,
    i.quantity - 30.000,
    i.quantity,
    'supplier',
    NULL,
    'Fresh milk delivery from Local Dairy Farm'
FROM inventory i 
WHERE i.name = 'Whole Milk' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Whole Milk')

UNION ALL

SELECT 
    i.id,
    'usage'::transaction_type,
    -1.200,
    i.quantity + 1.200,
    i.quantity,
    'order',
    NULL,
    'Vanilla syrup used in specialty drinks'
FROM inventory i 
WHERE i.name = 'Vanilla Syrup' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Vanilla Syrup')

UNION ALL

SELECT 
    i.id,
    'purchase'::transaction_type,
    500.000,
    i.quantity - 500.000,
    i.quantity,
    'supplier',
    NULL,
    'Monthly cup supply restock'
FROM inventory i 
WHERE i.name = 'Paper Cups - Medium' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Paper Cups - Medium')

UNION ALL

SELECT 
    i.id,
    'waste'::transaction_type,
    -2.000,
    i.quantity + 2.000,
    i.quantity,
    'quality_control',
    NULL,
    'Expired milk disposal'
FROM inventory i 
WHERE i.name = 'Oat Milk' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Oat Milk')

UNION ALL

SELECT 
    i.id,
    'adjustment'::transaction_type,
    -0.500,
    i.quantity + 0.500,
    i.quantity,
    'inventory_count',
    NULL,
    'Stock count adjustment - spillage'
FROM inventory i 
WHERE i.name = 'Caramel Syrup' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Caramel Syrup')

UNION ALL

SELECT 
    i.id,
    'purchase'::transaction_type,
    1.000,
    i.quantity - 1.000,
    i.quantity,
    'supplier',
    NULL,
    'Premium matcha powder restock'
FROM inventory i 
WHERE i.name = 'Matcha Powder' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Matcha Powder')

UNION ALL

SELECT 
    i.id,
    'usage'::transaction_type,
    -200.000,
    i.quantity + 200.000,
    i.quantity,
    'daily_operations',
    NULL,
    'Daily cup usage for orders'
FROM inventory i 
WHERE i.name = 'Paper Cups - Large' AND EXISTS (SELECT 1 FROM inventory WHERE name = 'Paper Cups - Large');

-- Insert order status history for completed orders
INSERT INTO order_status_history (order_id, old_status, new_status, changed_by, reason)
SELECT 
    o.id,
    'pending'::order_status,
    'preparing'::order_status,
    'barista_jenny',
    'Order accepted and preparation started'
FROM orders o
WHERE o.status IN ('preparing', 'ready', 'closed')

UNION ALL

SELECT 
    o.id,
    'preparing'::order_status,
    'ready'::order_status,
    'barista_mike',
    'Order completed and ready for pickup'
FROM orders o
WHERE o.status IN ('ready', 'closed')

UNION ALL

SELECT 
    o.id,
    'ready'::order_status,
    'closed'::order_status,
    'cashier_sarah',
    'Order picked up by customer'
FROM orders o
WHERE o.status = 'closed'

UNION ALL

SELECT 
    o.id,
    'pending'::order_status,
    'cancelled'::order_status,
    'manager_alex',
    'Customer requested cancellation'
FROM orders o
WHERE o.status = 'cancelled';

-- Insert price history for some menu items (simulating price changes)
INSERT INTO price_history (menu_item_id, old_price, new_price, changed_by, reason)
SELECT 
    m.id,
    m.price - 0.25,
    m.price,
    'manager_alex',
    'Seasonal price adjustment due to ingredient cost increase'
FROM menu_items m
WHERE m.name IN ('Vanilla Latte', 'Caramel Macchiato', 'Mocha')

UNION ALL

SELECT 
    m.id,
    m.price - 0.50,
    m.price,
    'owner_jessica',
    'Premium ingredient upgrade - organic coffee beans'
FROM menu_items m
WHERE m.name IN ('Classic Americano', 'Cappuccino')

UNION ALL

SELECT 
    m.id,
    m.price + 0.15,
    m.price,
    'manager_alex',
    'Promotional pricing reduction for new menu item'
FROM menu_items m
WHERE m.name IN ('Green Tea Latte', 'Chai Tea Latte')

UNION ALL

SELECT 
    m.id,
    m.price - 0.35,
    m.price,
    'corporate_pricing',
    'Market competition adjustment'
FROM menu_items m
WHERE m.name IN ('Iced Coffee', 'Hot Chocolate');
