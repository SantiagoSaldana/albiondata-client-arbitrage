package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/ao-data/albiondata-client/lib"
	_ "modernc.org/sqlite"
)

// MarketOrderDB represents a market order stored in the database
type MarketOrderDB struct {
	ID               int
	OrderID          int
	ItemID           string
	ItemGroupID      string
	LocationID       string
	QualityLevel     int
	EnchantmentLevel int
	Price            int
	Amount           int
	AuctionType      string
	Expires          string
	CapturedAt       string
}

var (
	DB *sql.DB
	mu sync.Mutex
)

// InitDB initializes the SQLite database
func InitDB(dbPath string) error {
	mu.Lock()
	defer mu.Unlock()

	var err error
	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	DB.SetMaxOpenConns(25)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(5 * time.Minute)

	// Create schema
	_, err = DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// InsertMarketOrder inserts or updates a market order in the database
func InsertMarketOrder(order *lib.MarketOrder) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `
		INSERT INTO market_orders (
			order_id, item_id, item_group_id, location_id,
			quality_level, enchantment_level, price, amount,
			auction_type, expires
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := DB.Exec(query,
		order.ID,
		order.ItemID,
		order.GroupTypeId,
		order.LocationID,
		order.QualityLevel,
		order.EnchantmentLevel,
		order.Price,
		order.Amount,
		order.AuctionType,
		order.Expires,
	)

	return err
}

// GetRecentOrders retrieves the most recent market orders
func GetRecentOrders(limit int) ([]*MarketOrderDB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, order_id, item_id, item_group_id, location_id,
		       quality_level, enchantment_level, price, amount,
		       auction_type, expires, captured_at
		FROM market_orders
		ORDER BY captured_at DESC
		LIMIT ?
	`

	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanOrders(rows)
}

// GetOrdersByFilter retrieves market orders based on filter criteria
func GetOrdersByFilter(itemID, location, auctionType string, limit int) ([]*MarketOrderDB, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, order_id, item_id, item_group_id, location_id,
		       quality_level, enchantment_level, price, amount,
		       auction_type, expires, captured_at
		FROM market_orders
		WHERE 1=1
	`
	args := []interface{}{}

	if itemID != "" {
		query += " AND item_id LIKE ?"
		args = append(args, "%"+itemID+"%")
	}
	if location != "" {
		query += " AND location_id = ?"
		args = append(args, location)
	}
	if auctionType != "" {
		query += " AND auction_type = ?"
		args = append(args, auctionType)
	}

	query += " ORDER BY captured_at DESC LIMIT ?"
	args = append(args, limit)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanOrders(rows)
}

// GetOrderCount returns the total number of orders in the database
func GetOrderCount() (int, error) {
	if DB == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM market_orders").Scan(&count)
	return count, err
}

// GetOrderCountByType returns order counts by auction type
func GetOrderCountByType() (buy, sell int, err error) {
	if DB == nil {
		return 0, 0, fmt.Errorf("database not initialized")
	}

	err = DB.QueryRow(`
		SELECT
			COALESCE(SUM(CASE WHEN auction_type = 'request' THEN 1 ELSE 0 END), 0) as buy_count,
			COALESCE(SUM(CASE WHEN auction_type = 'offer' THEN 1 ELSE 0 END), 0) as sell_count
		FROM market_orders
	`).Scan(&buy, &sell)

	return buy, sell, err
}

// GetUniqueLocations returns all unique location IDs
func GetUniqueLocations() ([]string, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	rows, err := DB.Query(`
		SELECT DISTINCT location_id
		FROM market_orders
		ORDER BY location_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []string
	for rows.Next() {
		var loc string
		if err := rows.Scan(&loc); err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, rows.Err()
}

// scanOrders scans SQL rows into MarketOrderDB structs
func scanOrders(rows *sql.Rows) ([]*MarketOrderDB, error) {
	var orders []*MarketOrderDB

	for rows.Next() {
		order := &MarketOrderDB{}
		err := rows.Scan(
			&order.ID,
			&order.OrderID,
			&order.ItemID,
			&order.ItemGroupID,
			&order.LocationID,
			&order.QualityLevel,
			&order.EnchantmentLevel,
			&order.Price,
			&order.Amount,
			&order.AuctionType,
			&order.Expires,
			&order.CapturedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

// Close closes the database connection
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if DB != nil {
		return DB.Close()
	}
	return nil
}
