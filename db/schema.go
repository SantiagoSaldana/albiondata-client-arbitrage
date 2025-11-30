package db

const schema = `
CREATE TABLE IF NOT EXISTS market_orders (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    order_id INTEGER NOT NULL,
    item_id TEXT NOT NULL,
    item_group_id TEXT,
    location_id TEXT NOT NULL,
    quality_level INTEGER,
    enchantment_level INTEGER,
    price INTEGER,
    amount INTEGER,
    auction_type TEXT,
    expires TEXT,
    captured_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(order_id, location_id, auction_type) ON CONFLICT REPLACE
);

CREATE INDEX IF NOT EXISTS idx_item_location ON market_orders(item_id, location_id);
CREATE INDEX IF NOT EXISTS idx_captured_at ON market_orders(captured_at DESC);
CREATE INDEX IF NOT EXISTS idx_auction_type ON market_orders(auction_type);
CREATE INDEX IF NOT EXISTS idx_item_id ON market_orders(item_id);
CREATE INDEX IF NOT EXISTS idx_location_id ON market_orders(location_id);
`
