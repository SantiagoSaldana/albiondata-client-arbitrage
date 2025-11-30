// +build windows

package systray

import (
	"fmt"
	"os"

	"github.com/ao-data/albiondata-client/client"
	"github.com/ao-data/albiondata-client/gui"

	"github.com/ao-data/albiondata-client/icon"
	"github.com/getlantern/systray"
	"github.com/gonutz/w32"
)

var consoleHidden bool

func hideConsole() {
	console := w32.GetConsoleWindow()
	if console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, w32.SW_HIDE)
		}
	}

	consoleHidden = true
}

func showConsole() {
	console := w32.GetConsoleWindow()
	if console != 0 {
		_, consoleProcID := w32.GetWindowThreadProcessId(console)
		if w32.GetCurrentProcessId() == consoleProcID {
			w32.ShowWindowAsync(console, w32.SW_SHOW)
		}
	}

	consoleHidden = false
}

func GetActionTitle() string {
	if consoleHidden {
		return "Show Console"
	} else {
		return "Hide Console"
	}
}

func Run() {
	systray.Run(onReady, onExit)
}

func onExit() {

}

func onReady() {
	// Don't hide the console automatically
	// Unless started from the scheduled task or with the parameter
	// People think it is crashing
	if client.ConfigGlobal.Minimize {
		hideConsole()
	}
	systray.SetIcon(icon.Data)
	systray.SetTitle("Albion Data Client")
	systray.SetTooltip("Albion Data Client")

	// Add GUI menu item if enabled
	var mShowGUI *systray.MenuItem
	if client.ConfigGlobal.GUIEnabled {
		mShowGUI = systray.AddMenuItem("Show Market Orders", "Open the market orders window")
		systray.AddSeparator()
	}

	mConHideShow := systray.AddMenuItem(GetActionTitle(), "Show/Hide Console")
	mQuit := systray.AddMenuItem("Quit", "Close the Albion Data Client")

	func() {
		for {
			select {
			case <-mQuit.ClickedCh:
				fmt.Println("Requesting quit")
				systray.Quit()
				os.Exit(0)
				fmt.Println("Finished quitting")

			case <-mConHideShow.ClickedCh:
				if consoleHidden == true {
					showConsole()
					mConHideShow.SetTitle(GetActionTitle())
				} else {
					hideConsole()
					mConHideShow.SetTitle(GetActionTitle())
				}

			case <-func() chan struct{} {
				if mShowGUI != nil {
					return mShowGUI.ClickedCh
				}
				// Return a channel that never sends if GUI is disabled
				ch := make(chan struct{})
				return ch
			}():
				// Show the GUI window
				if g := gui.GetGlobalGUI(); g != nil {
					g.Show()
				}
			}
		}
	}()
}
