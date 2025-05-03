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
	myConfig, err = godotenv.Read(fmt.Sprintf("%s/.env", os.Getenv("SWIFTBAR_PLUGINS_PATH")))
	if err != nil {
		log.Fatalln("Error loading .env file")
	}
}

func main() {
	// Initialize configuration
	if m, e := strconv.Atoi(myConfig["MAX"]); e == nil {
		MAX = m
	}
	
	// Get BluOS device URL
	bluePlayerUrl := myConfig["BLUE_URL"]
	
	// Create BitBar app
	app := bitbar.New()
	
	// Build the menu structure
	createMainMenu(app, bluePlayerUrl)
	
	// Render the menu
	app.Render()
}
