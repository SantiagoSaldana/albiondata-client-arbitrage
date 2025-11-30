package main

import (
	"os"
	"strings"
	"time"

	"github.com/ao-data/albiondata-client/client"
	"github.com/ao-data/albiondata-client/db"
	"github.com/ao-data/albiondata-client/gui"
	"github.com/ao-data/albiondata-client/log"
	"github.com/ao-data/albiondata-client/systray"

	"github.com/ao-data/go-githubupdate/updater"
)

var version string

func init() {
	client.ConfigGlobal.SetupFlags()
}

func main() {
	if client.ConfigGlobal.PrintVersion {
		log.Infof("Albion Data Client, version: %s", version)
		return
	}

	// Initialize database if enabled
	if client.ConfigGlobal.DatabaseEnabled {
		err := db.InitDB(client.ConfigGlobal.DatabasePath)
		if err != nil {
			log.Errorf("Failed to initialize database: %v", err)
			log.Error("Continuing without database support...")
		} else {
			defer db.Close()
			log.Infof("Database initialized at: %s", client.ConfigGlobal.DatabasePath)
		}
	}

	// Start GUI if enabled
	if client.ConfigGlobal.GUIEnabled {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf("GUI failed to initialize (may need to run from Windows desktop): %v", r)
					log.Info("Continuing without GUI. Database is still active.")
				}
			}()

			marketGUI := gui.NewGUI(client.ConfigGlobal.GUIRowsPerPage)
			if client.ConfigGlobal.GUIAutoRefreshSeconds > 0 {
				marketGUI.StartAutoRefresh(time.Duration(client.ConfigGlobal.GUIAutoRefreshSeconds) * time.Second)
			}

			// Run GUI (blocks until window closed)
			err := marketGUI.Run()
			if err != nil {
				log.Errorf("GUI error: %v", err)
			}
		}()
	}

	startUpdater()

	go systray.Run()

	c := client.NewClient(version)
	err := c.Run()
	if err != nil {
		log.Error(err)
		log.Error("The program encountered an error. Press any key to close this window.")
		var b = make([]byte, 1)
		_, _ = os.Stdin.Read(b)
	}

}

func startUpdater() {
	if version != "" && !strings.Contains(version, "dev") {
		u := updater.NewUpdater(
			version,
			"ao-data",
			"albiondata-client",
			"update-",
		)

		go func() {
			maxTries := 2
			for i := 0; i < maxTries; i++ {
				err := u.BackgroundUpdater()
				if err != nil {
					log.Error(err.Error())
					log.Info("Will try again in 60 seconds. You may need to run the client as Administrator.")
					// Sleep and hope the network connects
					time.Sleep(time.Second * 60)
				} else {
					break
				}
			}
		}()
	}
}
