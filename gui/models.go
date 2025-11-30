package gui

import (
	"fmt"
	"sort"

	"github.com/ao-data/albiondata-client/db"
	"github.com/lxn/walk"
)

// MarketOrderModel is the table model for displaying market orders
type MarketOrderModel struct {
	walk.TableModelBase
	walk.SorterBase
	sortColumn int
	sortOrder  walk.SortOrder
	orders     []*db.MarketOrderDB
}

// NewMarketOrderModel creates a new market order model
func NewMarketOrderModel() *MarketOrderModel {
	return &MarketOrderModel{
		orders: make([]*db.MarketOrderDB, 0),
	}
}

// RowCount returns the number of rows in the model
func (m *MarketOrderModel) RowCount() int {
	return len(m.orders)
}

// Value returns the value for a specific cell
func (m *MarketOrderModel) Value(row, col int) interface{} {
	if row >= len(m.orders) {
		return nil
	}

	order := m.orders[row]

	switch col {
	case 0:
		return order.OrderID
	case 1:
		return order.ItemID
	case 2:
		return order.LocationID
	case 3:
		return fmt.Sprintf("%d", order.QualityLevel)
	case 4:
		return fmt.Sprintf("%d", order.EnchantmentLevel)
	case 5:
		return formatNumber(order.Price)
	case 6:
		return order.Amount
	case 7:
		return formatAuctionType(order.AuctionType)
	case 8:
		return formatTime(order.CapturedAt)
	default:
		return nil
	}
}

// Sort sorts the model by the specified column
func (m *MarketOrderModel) Sort(col int, order walk.SortOrder) error {
	m.sortColumn = col
	m.sortOrder = order

	sort.SliceStable(m.orders, func(i, j int) bool {
		a, b := m.orders[i], m.orders[j]

		c := func(ls bool) bool {
			if m.sortOrder == walk.SortAscending {
				return ls
			}
			return !ls
		}

		switch col {
		case 0:
			return c(a.OrderID < b.OrderID)
		case 1:
			return c(a.ItemID < b.ItemID)
		case 2:
			return c(a.LocationID < b.LocationID)
		case 3:
			return c(a.QualityLevel < b.QualityLevel)
		case 4:
			return c(a.EnchantmentLevel < b.EnchantmentLevel)
		case 5:
			return c(a.Price < b.Price)
		case 6:
			return c(a.Amount < b.Amount)
		case 7:
			return c(a.AuctionType < b.AuctionType)
		case 8:
			return c(a.CapturedAt < b.CapturedAt)
		default:
			return false
		}
	})

	return m.SorterBase.Sort(col, order)
}

// SetOrders updates the orders in the model
func (m *MarketOrderModel) SetOrders(orders []*db.MarketOrderDB) {
	m.orders = orders
	m.PublishRowsReset()

	// Re-apply current sort
	if m.sortColumn >= 0 {
		m.Sort(m.sortColumn, m.sortOrder)
	}
}

// GetOrders returns the current orders
func (m *MarketOrderModel) GetOrders() []*db.MarketOrderDB {
	return m.orders
}

// Helper functions

func formatAuctionType(auctionType string) string {
	switch auctionType {
	case "offer":
		return "Sell"
	case "request":
		return "Buy"
	default:
		return auctionType
	}
}

func formatNumber(n int) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d,%03d,%03d", n/1000000, (n/1000)%1000, n%1000)
}

func formatTime(timestamp string) string {
	if len(timestamp) >= 19 {
		// Extract time portion from "2025-11-30 10:30:15"
		return timestamp[11:19]
	}
	return timestamp
}
