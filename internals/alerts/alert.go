package alerts

import (
	"fmt"
	"time"

	"github.com/gen2brain/beeep"
)

func SendOutageAlert(duration time.Duration) {
	beeep.AppName = "WifiTracker"

	err := beeep.Notify("Wifi Down", fmt.Sprintf("You have had an outage that lasted %d seconds", duration), `bin\warning.png`)
	if err != nil {
		panic(err)
	}
}
