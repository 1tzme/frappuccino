-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
-- Enable trigram extension for text search
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

CREATE TYPE order_status AS ENUM ('pending', 'preparing',  'ready', 'closed', 'cancelled');

CREATE TYPE item_size AS ENUM ('small', 'medium', 'large', 'extra_large');

CREATE TYPE unit_type AS ENUM ('grams', 'ml', 'pieces', 'kg', 'liters');

CREATE TYPE transaction_type AS ENUM ('purchase', 'usage', 'waste', 'adjustment', 'return');

CREATE TABLE inventory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    quantity DECIMAL(10,3) NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    unit unit_type NOT NULL,
    min_threshold DECIMAL(10,3) NOT NULL DEFAULT 0 CHECK (min_threshold >= 0),
    cost_per_unit DECIMAL(10,2) NOT NULL DEFAULT 0 CHECK (cost_per_unit >= 0),
    last_updated TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE menu_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL,
    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
    available BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB DEFAULT '{}',
    tags TEXT[] DEFAULT '{}',
    allergens TEXT[] DEFAULT '{}',
    available_sizes item_size[] DEFAULT ARRAY['medium']::item_size[],
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE menu_item_ingredients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    menu_item_id UUID NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    ingredient_id UUID NOT NULL REFERENCES inventory(id) ON DELETE RESTRICT,
    required_quantity DECIMAL(10,3) NOT NULL CHECK (required_quantity > 0),
    unit unit_type NOT NULL,
    UNIQUE(menu_item_id, ingredient_id)
);

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_name VARCHAR(255) NOT NULL,
    special_instructions JSONB DEFAULT '{}',
    status order_status NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0 CHECK (total_amount >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE order_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    menu_item_id UUID NOT NULL REFERENCES menu_items(id) ON DELETE RESTRICT,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    price_at_time DECIMAL(10,2) NOT NULL CHECK (price_at_time >= 0),
    customizations JSONB DEFAULT '{}'
);

CREATE TABLE order_status_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    old_status order_status,
    new_status order_status NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_by VARCHAR(255) DEFAULT 'system',
    reason TEXT
);

CREATE TABLE price_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    menu_item_id UUID NOT NULL REFERENCES menu_items(id) ON DELETE CASCADE,
    old_price DECIMAL(10,2),
    new_price DECIMAL(10,2) NOT NULL CHECK (new_price >= 0),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    changed_by VARCHAR(255) DEFAULT 'system',
    reason TEXT
);

CREATE TABLE inventory_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ingredient_id UUID NOT NULL REFERENCES inventory(id) ON DELETE CASCADE,
    transaction_type transaction_type NOT NULL,
    quantity_change DECIMAL(10,3) NOT NULL,
    quantity_before DECIMAL(10,3) NOT NULL,
    quantity_after DECIMAL(10,3) NOT NULL,
    transaction_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reference_type VARCHAR(50),
    reference_id UUID,
    notes TEXT
);

-- INDEXES
CREATE INDEX idx_orders_customer_name ON orders(customer_name);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
CREATE INDEX idx_orders_updated_at ON orders(updated_at);

CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_menu_item_id ON order_items(menu_item_id);

CREATE INDEX idx_menu_items_category ON menu_items(category);
CREATE INDEX idx_menu_items_available ON menu_items(available);
CREATE INDEX idx_menu_items_price ON menu_items(price);

CREATE INDEX idx_inventory_name ON inventory(name);
CREATE INDEX idx_inventory_quantity ON inventory(quantity);
CREATE INDEX idx_inventory_last_updated ON inventory(last_updated);

CREATE INDEX idx_menu_items_name_trgm ON menu_items USING gin(name gin_trgm_ops);
CREATE INDEX idx_menu_items_description_trgm ON menu_items USING gin(description gin_trgm_ops);
CREATE INDEX idx_orders_customer_trgm ON orders USING gin(customer_name gin_trgm_ops);

CREATE INDEX idx_order_status_history_order_changed ON order_status_history(order_id, changed_at);
CREATE INDEX idx_price_history_item_changed ON price_history(menu_item_id, changed_at);
CREATE INDEX idx_inventory_transactions_ingredient_date ON inventory_transactions(ingredient_id, transaction_date);

CREATE INDEX idx_menu_items_tags ON menu_items USING gin(tags);
CREATE INDEX idx_menu_items_allergens ON menu_items USING gin(allergens);
CREATE INDEX idx_menu_items_metadata ON menu_items USING gin(metadata);
CREATE INDEX idx_orders_special_instructions ON orders USING gin(special_instructions);
CREATE INDEX idx_order_items_customizations ON order_items USING gin(customizations);

-- TRIGGERS
-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_orders_updated_at BEFORE UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_menu_items_updated_at BEFORE UPDATE ON menu_items
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Function to track order status changes
CREATE OR REPLACE FUNCTION track_order_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status != NEW.status THEN
        INSERT INTO order_status_history (order_id, old_status, new_status, reason)
        VALUES (NEW.id, OLD.status, NEW.status, 'Status updated');
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for order status tracking
CREATE TRIGGER track_order_status AFTER UPDATE ON orders
    FOR EACH ROW EXECUTE FUNCTION track_order_status_change();

-- Function to track price changes
CREATE OR REPLACE FUNCTION track_price_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.price != NEW.price THEN
        INSERT INTO price_history (menu_item_id, old_price, new_price, reason)
        VALUES (NEW.id, OLD.price, NEW.price, 'Price updated');
    END IF;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for price tracking
CREATE TRIGGER track_price_changes AFTER UPDATE ON menu_items
    FOR EACH ROW EXECUTE FUNCTION track_price_change();

-- Function to update inventory last_updated
CREATE OR REPLACE FUNCTION update_inventory_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.last_updated = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger for inventory updates
CREATE TRIGGER update_inventory_timestamp BEFORE UPDATE ON inventory
    FOR EACH ROW EXECUTE FUNCTION update_inventory_timestamp();