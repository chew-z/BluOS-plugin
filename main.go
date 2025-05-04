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

	// Get BluOS device URL
	bluePlayerUrl := myConfig["BLUE_URL"]
	log.Printf("Using BluOS URL: %s", bluePlayerUrl)

	// Create BitBar app
	app := bitbar.New()

	// Try to contact the player
	statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
	stateXML, err := getXML(statusUrl)
	if err != nil {
		log.Printf("Failed to contact BluOS player: %v", err)

		// Show disconnected status
		submenu := app.NewSubMenu()
		app.StatusLine(":exclamationmark.triangle.fill: BluOS Disconnected").DropDown(false).Color("red")
		submenu.Line(":exclamationmark.triangle.fill: Player Disconnected").Color("red")
		submenu.Line("Check if your BluOS player is turned on").Color("gray")
		submenu.Line("Make sure you're on the same network").Color("gray")
		submenu.Line(fmt.Sprintf("Network: %s", myConfig["BLUE_WIFI"])).Color("gray")
		submenu.Line(fmt.Sprintf("URL: %s", bluePlayerUrl)).Color("gray")
	} else {
		// We're connected successfully
		log.Printf("Successfully connected to BluOS player (%d bytes received)", len(stateXML))

		// Use the modular menu builder from menu.go
		buildPlayerMenu(&app, bluePlayerUrl)
	}

	// Render the menu
	app.Render()
}
