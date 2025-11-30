package gui

import (
	"fmt"
	"time"

	"github.com/ao-data/albiondata-client/db"
	"github.com/ao-data/albiondata-client/log"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// MarketOrdersGUI represents the main GUI window
type MarketOrdersGUI struct {
	mainWindow     *walk.MainWindow
	tableView      *walk.TableView
	model          *MarketOrderModel
	filterItem     *walk.LineEdit
	filterLocation *walk.ComboBox
	filterType     *walk.ComboBox
	statsLabel     *walk.Label
	lastUpdated    *walk.Label
	refreshTicker  *time.Ticker
	stopRefresh    chan bool
	rowsPerPage    int
}

var globalGUI *MarketOrdersGUI

// NewGUI creates a new MarketOrdersGUI instance
func NewGUI(rowsPerPage int) *MarketOrdersGUI {
	if rowsPerPage <= 0 {
		rowsPerPage = 100
	}

	gui := &MarketOrdersGUI{
		model:       NewMarketOrderModel(),
		stopRefresh: make(chan bool),
		rowsPerPage: rowsPerPage,
	}
	globalGUI = gui
	return gui
}

// Run starts the GUI
func (g *MarketOrdersGUI) Run() error {
	err := g.createWindow()
	if err != nil {
		return err
	}

	// Initial data load
	g.Refresh()

	// Show window
	g.mainWindow.Run()

	return nil
}

// createWindow creates the main window UI
func (g *MarketOrdersGUI) createWindow() (err error) {
	// Catch any panics from Walk library
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("GUI initialization failed: %v", r)
		}
	}()

	// Build filter location items
	locations := []string{"All"}
	dbLocations, _ := db.GetUniqueLocations()
	locations = append(locations, dbLocations...)

	// Create main window
	mw := MainWindow{
		AssignTo: &g.mainWindow,
		Title:    "Albion Market Orders Tracker",
		MinSize:  Size{Width: 1000, Height: 600},
		Size:     Size{Width: 1200, Height: 700},
		Layout:   VBox{},
		Children: []Widget{
			// Filter section
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{Text: "Item:"},
					LineEdit{
						AssignTo: &g.filterItem,
						OnEditingFinished: func() {
							g.Refresh()
						},
					},
					Label{Text: "Location:"},
					ComboBox{
						AssignTo:      &g.filterLocation,
						Model:         locations,
						CurrentIndex:  0,
						OnCurrentIndexChanged: func() {
							g.Refresh()
						},
					},
					Label{Text: "Type:"},
					ComboBox{
						AssignTo:      &g.filterType,
						Model:         []string{"All", "Buy", "Sell"},
						CurrentIndex:  0,
						OnCurrentIndexChanged: func() {
							g.Refresh()
						},
					},
					PushButton{
						Text: "Refresh",
						OnClicked: func() {
							g.Refresh()
						},
					},
					PushButton{
						Text: "Clear Filters",
						OnClicked: func() {
							g.filterItem.SetText("")
							g.filterLocation.SetCurrentIndex(0)
							g.filterType.SetCurrentIndex(0)
							g.Refresh()
						},
					},
				},
			},

			// Statistics section
			Composite{
				Layout: HBox{},
				Children: []Widget{
					Label{
						AssignTo:  &g.statsLabel,
						Text:      "Loading...",
						TextColor: walk.RGB(0, 100, 0),
					},
					HSpacer{},
					Label{
						AssignTo: &g.lastUpdated,
						Text:     "Last Updated: -",
					},
				},
			},

			// Table view
			TableView{
				AssignTo:         &g.tableView,
				AlternatingRowBG: true,
				ColumnsOrderable: true,
				MultiSelection:   false,
				Columns: []TableViewColumn{
					{Title: "Order ID", Width: 80},
					{Title: "Item ID", Width: 150},
					{Title: "Location", Width: 100},
					{Title: "Quality", Width: 60},
					{Title: "Enchant", Width: 60},
					{Title: "Price", Width: 100},
					{Title: "Amount", Width: 80},
					{Title: "Type", Width: 60},
					{Title: "Time", Width: 90},
				},
				Model: g.model,
			},
		},
	}

	err = mw.Create()
	if err != nil {
		return err
	}

	return nil
}

// Refresh updates the data in the GUI
func (g *MarketOrdersGUI) Refresh() {
	if g.mainWindow == nil {
		return
	}

	g.mainWindow.Synchronize(func() {
		// Get filter values
		itemFilter := ""
		if g.filterItem != nil {
			itemFilter = g.filterItem.Text()
		}

		locationFilter := ""
		if g.filterLocation != nil && g.filterLocation.CurrentIndex() > 0 {
			locationFilter = g.filterLocation.Text()
		}

		typeFilter := ""
		if g.filterType != nil {
			switch g.filterType.CurrentIndex() {
			case 1:
				typeFilter = "request"
			case 2:
				typeFilter = "offer"
			}
		}

		// Fetch orders from database
		var orders []*db.MarketOrderDB
		var err error

		if itemFilter == "" && locationFilter == "" && typeFilter == "" {
			orders, err = db.GetRecentOrders(g.rowsPerPage)
		} else {
			orders, err = db.GetOrdersByFilter(itemFilter, locationFilter, typeFilter, g.rowsPerPage)
		}

		if err != nil {
			log.Errorf("Failed to fetch orders: %v", err)
			return
		}

		// Update model
		g.model.SetOrders(orders)

		// Update statistics
		g.UpdateStats()

		// Update last updated time
		if g.lastUpdated != nil {
			g.lastUpdated.SetText(fmt.Sprintf("Last Updated: %s", time.Now().Format("15:04:05")))
		}
	})
}

// UpdateStats updates the statistics display
func (g *MarketOrdersGUI) UpdateStats() {
	if g.statsLabel == nil {
		return
	}

	totalCount, err := db.GetOrderCount()
	if err != nil {
		log.Errorf("Failed to get order count: %v", err)
		return
	}

	buyCount, sellCount, err := db.GetOrderCountByType()
	if err != nil {
		log.Errorf("Failed to get order counts by type: %v", err)
		return
	}

	statsText := fmt.Sprintf("Total Orders: %d  |  Buy Orders: %d  |  Sell Orders: %d  |  Showing: %d",
		totalCount, buyCount, sellCount, g.model.RowCount())

	g.statsLabel.SetText(statsText)
}

// StartAutoRefresh starts automatic refresh of the data
func (g *MarketOrdersGUI) StartAutoRefresh(interval time.Duration) {
	if interval <= 0 {
		return
	}

	g.refreshTicker = time.NewTicker(interval)

	go func() {
		for {
			select {
			case <-g.refreshTicker.C:
				g.Refresh()
			case <-g.stopRefresh:
				g.refreshTicker.Stop()
				return
			}
		}
	}()
}

// StopAutoRefresh stops automatic refresh
func (g *MarketOrdersGUI) StopAutoRefresh() {
	if g.refreshTicker != nil {
		g.stopRefresh <- true
	}
}

// Show shows the window
func (g *MarketOrdersGUI) Show() {
	if g.mainWindow != nil {
		g.mainWindow.Show()
		g.mainWindow.SetFocus()
	}
}

// Hide hides the window
func (g *MarketOrdersGUI) Hide() {
	if g.mainWindow != nil {
		g.mainWindow.Hide()
	}
}

// GetGlobalGUI returns the global GUI instance
func GetGlobalGUI() *MarketOrdersGUI {
	return globalGUI
}
