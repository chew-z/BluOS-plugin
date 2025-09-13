package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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

func main() {
	// Initialize configuration
	if m, e := strconv.Atoi(myConfig["MAX"]); e == nil {
		MAX = m
	}

	// Get BluOS device URL (try discovery first, fall back to config)
	bluePlayerUrl, err := getBluOSPlayerURL(myConfig["BLUE_URL"])
	if err != nil {
		log.Printf("Failed to determine BluOS player URL: %v", err)

		// Create error menu
		app := bitbar.New()
		submenu := app.NewSubMenu()
		app.StatusLine(":exclamationmark.triangle.fill: BluOS Not Found").DropDown(false).Color("red")
		submenu.Line(":exclamationmark.triangle.fill: No BluOS Device Found").Color("red")
		submenu.Line("Auto-discovery failed and no BLUE_URL configured").Color("gray")
		submenu.Line("---")
		submenu.Line("Troubleshooting:").Color("gray")
		submenu.Line("• Ensure BluOS device is powered on").Color("gray")
		submenu.Line("• Check you're on the same Wi-Fi network").Color("gray")
		submenu.Line("• Set BLUE_URL in .env if discovery fails").Color("gray")
		submenu.Line(fmt.Sprintf("Network: %s", myConfig["BLUE_WIFI"])).Color("gray")
		app.Render()
		return
	}
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
