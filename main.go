package main

import (
	"encoding/xml"
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
	
	// Add a fallback item in case everything else fails
	app.StatusLine("🔊 BluOS").DropDown(false)
	submenu := app.NewSubMenu()
	
	// Try to contact the player
	statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
	stateXML, err := getXML(statusUrl)
	if err != nil {
		log.Printf("Failed to contact BluOS player: %v", err)
		
		// Show disconnected status
		app.StatusLine("⚠️ BluOS Disconnected").DropDown(false).Color("red")
		submenu.Line("⚠️ Player Disconnected").Color("red")
		submenu.Line("Check if your BluOS player is turned on").Color("gray")
		submenu.Line("Make sure you're on the same network").Color("gray")
		submenu.Line(fmt.Sprintf("Network: %s", myConfig["BLUE_WIFI"])).Color("gray")
		submenu.Line(fmt.Sprintf("URL: %s", bluePlayerUrl)).Color("gray")
		
		// Add a refresh option
		submenu.Line("---")
		submenu.Line("🔄 Refresh").Refresh()
	} else {
		// We're connected successfully
		log.Printf("Successfully connected to BluOS player (%d bytes received)", len(stateXML))
		
		// Parse the state XML directly
		var state StateXML
		if err := xml.Unmarshal(stateXML, &state); err != nil {
			log.Printf("Error parsing player state: %v", err)
			submenu.Line("Error parsing player data").Color("red")
		} else {
			// Add basic player info directly to main menu
			log.Printf("Player state: %s, Service: %s", state.State, state.Service)
			
			// Set icon based on state
			var icon string
			switch state.State {
			case "play":
				icon = ":music.note.list:"
			case "stream":
				if state.Service == "Capture" {
					icon = ":tv:"
				} else {
					icon = ":radio:"
				}
			case "pause":
				icon = ":pause.fill:"
			case "stop":
				icon = ":stop.fill:"
			default:
				icon = ":questionmark.circle:"
			}
			
			app.StatusLine(fmt.Sprintf("%s %s", icon, state.Title1)).DropDown(false).Length(MAX)
			submenu.Line(fmt.Sprintf("State: %s", state.State))
			
			if state.Title1 != "" {
				submenu.Line(fmt.Sprintf("Title: %s", state.Title1))
			}
			
			if state.ServiceName != "" {
				submenu.Line(fmt.Sprintf("Service: %s", state.ServiceName))
			}
			
			// Add play/pause control
			if state.State == "play" || state.State == "stream" {
				pauseCmd := createCommand(fmt.Sprintf("%s/Pause", bluePlayerUrl))
				submenu.Line("⏸️ Pause").Command(pauseCmd)
			} else if state.State == "pause" {
				playCmd := createCommand(fmt.Sprintf("%s/Play", bluePlayerUrl))
				submenu.Line("▶️ Play").Command(playCmd)
			} else if state.State == "stop" {
				playCmd := createCommand(fmt.Sprintf("%s/Play", bluePlayerUrl))
				submenu.Line("▶️ Play").Command(playCmd)
			}
		}
		
		// Add radio presets
		submenu.Line("--- Radio Presets ---")
		presetsUrl := fmt.Sprintf("%s/Presets", bluePlayerUrl)
		presetXML, err := getXML(presetsUrl)
		if err == nil {
			var presets Presets
			if err := xml.Unmarshal(presetXML, &presets); err == nil {
				log.Printf("Found %d presets", len(presets.Preset))
				for _, p := range presets.Preset {
					l := fmt.Sprintf("📻 %s - %s", p.ID, p.Name)
					c := fmt.Sprintf("%s/Preset?id=%s", bluePlayerUrl, p.ID)
					cmd := createCommand(c)
					submenu.Line(l).Command(cmd)
				}
			}
		}
		
		// Add volume info
		submenu.Line("--- Volume Controls ---")
		volumeUrl := fmt.Sprintf("%s/Volume", bluePlayerUrl)
		volXML, err := getXML(volumeUrl)
		if err == nil {
			var volStatus VolumeStatus
			if err := xml.Unmarshal(volXML, &volStatus); err == nil {
				// Display volume information - dB as primary, percentage as alternate
				volColor := "orange"
				if volStatus.Mute == 1 {
					// For muted state, show in red
					volColor = "red"
					submenu.Line(fmt.Sprintf("🔇 Volume: %.1f dB (Muted)", volStatus.Db)).Color(volColor)
					submenu.Line(fmt.Sprintf("🔇 Volume: %d%% (Muted)", volStatus.Level)).Alternate(true).Color(volColor)
				} else {
					submenu.Line(fmt.Sprintf("🔊 Volume: %.1f dB", volStatus.Db)).Color(volColor)
					submenu.Line(fmt.Sprintf("🔊 Volume: %d%%", volStatus.Level)).Alternate(true).Color(volColor)
				}
				
				// Volume presets
				submenu.Line("--- Volume Presets ---")
				volumePresets := []struct {
					Label string
					Level int
				}{
					{"📢 Max (100%)", 100},
					{"🔊 High (80%)", 80},
					{"🔉 Medium (50%)", 50},
					{"🔈 Low (20%)", 20},
				}
				
				for _, preset := range volumePresets {
					presetCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"level": strconv.Itoa(preset.Level)})
					submenu.Line(preset.Label).Command(presetCmd)
				}
				
				// Add mute toggle
				if volStatus.Mute == 1 {
					unmuteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "0"})
					submenu.Line("🔈 Unmute").Command(unmuteCmd)
				} else {
					muteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "1"})
					submenu.Line("🔇 Mute").Command(muteCmd)
				}
			}
		}
	}
	
	// Make sure we have at least one line in case everything else fails
	submenu.Line("🔄 Refresh").Refresh()
	
	// Render the menu
	app.Render()
}
