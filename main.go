package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/johnmccabe/go-bitbar"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

var (
	MAX      = 40
	myConfig map[string]string
	TMP      = os.Getenv("TMPDIR")
)

func init() {
	var err error
	envPath := fmt.Sprintf("%s/.env", os.Getenv("SWIFTBAR_PLUGINS_PATH"))
	log.Printf("Loading env file from: %s", envPath)

	myConfig, err = godotenv.Read(envPath)
	if err != nil {
		log.Fatalln("Error loading .env file:", err)
	}

	// Log all config values for debugging
	for key, value := range myConfig {
		log.Printf("Config: %s = %s", key, value)
	}
}

// isDeviceReachable performs a simple network check to see if the device is reachable,
// even if the main API might be having issues
func isDeviceReachable(url string) bool {
	// Try a few different endpoints to increase chances of success
	endpoints := []string{"/Status", "/Volume", "/"}
	
	for _, endpoint := range endpoints {
		fullURL := fmt.Sprintf("%s%s", url, endpoint)
		log.Printf("Checking device reachability via: %s", fullURL)
		
		// Use a simple check with long timeout for reachability testing
		client := &http.Client{
			Timeout: 15 * time.Second, // Longer timeout for reachability check
		}
		
		resp, err := client.Get(fullURL)
		if err == nil {
			resp.Body.Close() // Don't forget to close the body
			log.Printf("Device is reachable via %s (status: %d)", endpoint, resp.StatusCode)
			return true
		}
		
		log.Printf("Failed to reach device via %s: %v", endpoint, err)
	}
	
	log.Printf("Device appears to be completely unreachable")
	return false
}

func main() {
	// Initialize configuration
	if m, e := strconv.Atoi(myConfig["MAX"]); e == nil {
		MAX = m
	}

	// Get BluOS device URL
	bluePlayerUrl := myConfig["BLUE_URL"]
	log.Printf("Using BluOS URL: %s", bluePlayerUrl)

	// Create BitBar app
	app := bitbar.New()

	// Try to contact the player
	statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
	stateXML, err := getXML(statusUrl)
	if err != nil {
		log.Printf("Failed to get BluOS status XML: %v", err)
		
		// Check if the device is reachable at all
		if isDeviceReachable(bluePlayerUrl) {
			// Device is reachable but API might be having issues
			submenu := app.NewSubMenu()
			app.StatusLine(":exclamationmark.circle.fill: BluOS Issues").DropDown(false).Color("orange")
			submenu.Line(":exclamationmark.circle.fill: Player API Issues").Color("orange")
			submenu.Line("Device is reachable but API is not responding properly").Color("gray")
			submenu.Line("The player might be updating or rebooting").Color("gray")
			submenu.Line("Try again in a few minutes").Color("gray")
			submenu.Line(fmt.Sprintf("URL: %s", bluePlayerUrl)).Color("gray")
			submenu.Line("---")
			submenu.Line("Attempt Manual Refresh").Command(createCommand(statusUrl))
		} else {
			// Device appears to be completely offline
			submenu := app.NewSubMenu()
			app.StatusLine(":exclamationmark.triangle.fill: BluOS Disconnected").DropDown(false).Color("red")
			submenu.Line(":exclamationmark.triangle.fill: Player Disconnected").Color("red")
			submenu.Line("Check if your BluOS player is turned on").Color("gray")
			submenu.Line("Make sure you're on the same network").Color("gray")
			submenu.Line(fmt.Sprintf("Network: %s", myConfig["BLUE_WIFI"])).Color("gray")
			submenu.Line(fmt.Sprintf("URL: %s", bluePlayerUrl)).Color("gray")
			submenu.Line("---")
			submenu.Line("Attempt Manual Refresh").Command(createCommand(statusUrl))
		}
	} else {
		// We're connected successfully
		log.Printf("Successfully connected to BluOS player (%d bytes received)", len(stateXML))

		// Use the modular menu builder from menu.go
		buildPlayerMenu(&app, bluePlayerUrl)
	}

	// Render the menu
	app.Render()
}
