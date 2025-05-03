package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

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
	submenu := app.NewSubMenu()
	
	// Try to contact the player
	statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
	stateXML, err := getXML(statusUrl)
	if err != nil {
		log.Printf("Failed to contact BluOS player: %v", err)
		
		// Show disconnected status
		app.StatusLine(":exclamationmark.triangle.fill: BluOS Disconnected").DropDown(false).Color("red")
		submenu.Line(":exclamationmark.triangle.fill: Player Disconnected").Color("red")
		submenu.Line("Check if your BluOS player is turned on").Color("gray")
		submenu.Line("Make sure you're on the same network").Color("gray")
		submenu.Line(fmt.Sprintf("Network: %s", myConfig["BLUE_WIFI"])).Color("gray")
		submenu.Line(fmt.Sprintf("URL: %s", bluePlayerUrl)).Color("gray")
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
				icon = ":play.circle.fill:"
			case "stream":
				if state.Service == "Capture" {
					icon = ":display:"
				} else if state.Service == "AirPlay" {
					icon = ":airplayaudio:"
				} else if state.Service == "Spotify" {
					icon = ":music.note.list:"
				} else {
					icon = ":radio.fill:"
				}
			case "pause":
				icon = ":pause.circle.fill:"
			case "stop":
				icon = ":stop.circle.fill:"
			default:
				icon = ":questionmark.circle.fill:"
			}
			
			// Collect potential titles for menu bar
			potentialTitles := []string{
				state.Title1,    // First title
				state.Title2,    // Second title (album, artist, etc.)
				state.Title3,    // Third title (optional additional info)
			}

			// Filter out empty or whitespace-only titles
			var nonEmptyTitles []string
			for _, title := range potentialTitles {
				trimmedTitle := strings.TrimSpace(title)
				if trimmedTitle != "" {
					nonEmptyTitles = append(nonEmptyTitles, trimmedTitle)
				}
			}

			// Fallback if no titles are available
			if len(nonEmptyTitles) == 0 {
				nonEmptyTitles = []string{fmt.Sprintf("%s %s", state.State, state.ServiceName)}
			}

			// Display menu bar lines
			for i, title := range nonEmptyTitles {
				if i == 0 {
					// First line gets the icon
					app.StatusLine(fmt.Sprintf("%s %s", icon, title)).Length(MAX).DropDown(false)
				} else {
					// Subsequent lines
					app.StatusLine(title).Length(MAX).DropDown(false)
				}
			}
			
			// Add play/pause control with SF Symbol icons
			if state.State == "play" || state.State == "stream" {
				pauseCmd := createCommand(fmt.Sprintf("%s/Pause", bluePlayerUrl))
				submenu.Line(":pause.circle.fill: Pause").Command(pauseCmd)
			} else if state.State == "pause" {
				playCmd := createCommand(fmt.Sprintf("%s/Play", bluePlayerUrl))
				submenu.Line(":play.circle.fill: Play").Command(playCmd)
			} else if state.State == "stop" {
				playCmd := createCommand(fmt.Sprintf("%s/Play", bluePlayerUrl))
				submenu.Line(":play.circle.fill: Play").Command(playCmd)
			}
			
			// Combine State and Service as a single alternate line
			var alternateInfo string
			if state.ServiceName != "" {
				alternateInfo = fmt.Sprintf("State: %s - Service: %s", state.State, state.ServiceName)
			} else {
				alternateInfo = fmt.Sprintf("State: %s", state.State)
			}
			submenu.Line(alternateInfo).Alternate(true)
			
			// Fine volume control as alternate lines
			submenu.Line(":speaker.wave.3.fill: Volume Up 1dB").Alternate(true).Command(
				createVolumeCommand(bluePlayerUrl, map[string]string{"db": "1.0"}),
			)
			submenu.Line(":speaker.wave.1.fill: Volume Down 1dB").Alternate(true).Command(
				createVolumeCommand(bluePlayerUrl, map[string]string{"db": "-1.0"}),
			)
			submenu.Line(":speaker.wave.2.fill: Volume Up 0.5dB").Alternate(true).Command(
				createVolumeCommand(bluePlayerUrl, map[string]string{"db": "0.5"}),
			)
			submenu.Line(":speaker.wave.2.fill: Volume Down 0.5dB").Alternate(true).Command(
				createVolumeCommand(bluePlayerUrl, map[string]string{"db": "-0.5"}),
			)
		}
		
		// Add radio presets
		submenu.Line("---")
		submenu.Line(":radio.fill: Radio Presets")
		presetsUrl := fmt.Sprintf("%s/Presets", bluePlayerUrl)
		presetXML, err := getXML(presetsUrl)
		if err == nil {
			var presets Presets
			if err := xml.Unmarshal(presetXML, &presets); err == nil {
				log.Printf("Found %d presets", len(presets.Preset))
				for _, p := range presets.Preset {
					l := fmt.Sprintf(":star.fill: %s - %s", p.ID, p.Name)
					c := fmt.Sprintf("%s/Preset?id=%s", bluePlayerUrl, p.ID)
					cmd := createCommand(c)
					submenu.Line(l).Command(cmd)
				}
			}
		}
		
		// Add volume info
		submenu.Line("---")
		submenu.Line(":speaker.wave.2.fill: Volume Controls")
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
					submenu.Line(fmt.Sprintf(":speaker.slash.fill: Volume: %.1f dB (Muted)", volStatus.Db)).Color(volColor)
					submenu.Line(fmt.Sprintf(":speaker.slash.fill: Volume: %d%% (Muted)", volStatus.Level)).Alternate(true).Color(volColor)
				} else {
					submenu.Line(fmt.Sprintf(":speaker.wave.2.fill: Volume: %.1f dB", volStatus.Db)).Color(volColor)
					submenu.Line(fmt.Sprintf(":speaker.wave.2.fill: Volume: %d%%", volStatus.Level)).Alternate(true).Color(volColor)
				}
				
				// Volume presets
				submenu.Line("---")
				submenu.Line(":list.star: Volume Presets")
				volumePresets := []struct {
					Label string
					Level int
				}{
					{":arrow.up.circle.fill: Max (100%)", 100},
					{":arrow.up.circle: High (80%)", 80},
					{":arrow.down.circle: Medium (50%)", 50},
					{":arrow.down.circle.fill: Low (20%)", 20},
				}
				
				for _, preset := range volumePresets {
					presetCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"level": strconv.Itoa(preset.Level)})
					submenu.Line(preset.Label).Command(presetCmd)
				}
				
				// Add mute toggle
				if volStatus.Mute == 1 {
					unmuteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "0"})
					submenu.Line(":speaker.wave.2.fill: Unmute").Command(unmuteCmd)
				} else {
					muteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "1"})
					submenu.Line(":speaker.slash.fill: Mute").Command(muteCmd)
				}
			}
		}
	}
	
	// Render the menu
	app.Render()
}
