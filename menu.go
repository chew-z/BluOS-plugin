package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"

	"github.com/johnmccabe/go-bitbar"
)

// buildPlayerMenu builds the main menu structure based on player state and volume info
func buildPlayerMenu(app *bitbar.Plugin, bluePlayerUrl string) {
	log.Printf("Building player menu for %s", bluePlayerUrl)
	statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
	presetsUrl := fmt.Sprintf("%s/Presets", bluePlayerUrl)
	volumeUrl := fmt.Sprintf("%s/Volume", bluePlayerUrl)

	submenu := app.NewSubMenu()

	// Process status data and create status bar
	createStatusDisplay(app, submenu, statusUrl, bluePlayerUrl)

	// Add separator
	submenu.Line("---")

	// Add radio presets directly (no header)
	addRadioPresets(submenu, presetsUrl, bluePlayerUrl)

	// Add separator
	submenu.Line("---")

	// Add volume info (no header)
	volStatus := addVolumeInfo(submenu, volumeUrl)

	// Add volume presets (no header)
	addVolumePresets(submenu, bluePlayerUrl, volStatus)

	// Add mute toggle
	addMuteToggle(submenu, bluePlayerUrl, volStatus)

	log.Printf("Menu building completed")
}

// createStatusDisplay adds the status bar display based on player state
func createStatusDisplay(app *bitbar.Plugin, submenu *bitbar.SubMenu, statusUrl, bluePlayerUrl string) {
	log.Printf("Creating status display")
	xmlBytes, err := getXML(statusUrl)
	if err != nil {
		submenu.Line(err.Error()).Color("red").Length(MAX)
		log.Printf("Failed to get XML: %v", err)
		return
	}

	// Log the first part of the XML response for debugging
	if len(xmlBytes) > 0 {
		previewLength := 200
		if len(xmlBytes) < previewLength {
			previewLength = len(xmlBytes)
		}
		log.Printf("XML response preview: %s", string(xmlBytes[:previewLength]))
	}

	var state StateXML
	if err := xml.Unmarshal(xmlBytes, &state); err != nil {
		log.Printf("Failed to parse status XML: %v", err)

		// Try to display something even if XML parsing fails
		submenu.Line("XML parsing error - Limited display available").Color("orange")
		submenu.Line(fmt.Sprintf("Error: %v", err)).Color("red")

		// Add a raw view option that might be useful for debugging
		submenu.Line("Raw data available - device is connected").Color("blue")
		return
	}

	// Log player state information
	log.Printf("Player state: %s, Service: %s", state.State, state.Service)
	log.Printf("Titles: '%s', '%s', '%s'", state.Title1, state.Title2, state.Title3)

	c := fmt.Sprintf("%s/Pause?toggle=1", bluePlayerUrl)
	cmd := createCommand(c)

	// Handle different states
	switch state.State {
	case "connecting":
		icon := ":bolt.fill:"
		l1 := fmt.Sprintf("%s connecting", icon)
		app.StatusLine(l1).DropDown(false).Length(MAX)

	case "play":
		icon := ":play.circle.fill:"
		if state.Shuffle == "1" {
			icon = ":shuffle.circle.fill:"
		}
		icon2 := ":pause.circle.fill:"
		l1 := fmt.Sprintf("%s %s", icon, state.Name)
		l2 := fmt.Sprintf("%s %s", icon, state.Album)
		l3 := fmt.Sprintf("%s %s", icon, state.Artist)
		s1 := fmt.Sprintf("%s %s: %s", icon2, state.ServiceName, state.Name)
		s2 := fmt.Sprintf("%s %s", ":music.note.list:", state.Quality)

		app.StatusLine(l1).DropDown(false).Length(MAX)
		app.StatusLine(l2).DropDown(false).Length(MAX)
		app.StatusLine(l3).DropDown(false).Length(MAX)
		submenu.Line(s1).Command(cmd)
		submenu.Line(s2).Alternate(true)

	case "stream":
		var icon string
		if state.Service == "AirPlay" {
			icon = ":airplayaudio:"
		} else if state.Service == "Spotify" {
			icon = ":music.note.list:"
		} else if state.Service == "Capture" {
			icon = ":display:"
			log.Printf("Found Capture service (HDMI ARC)")
		} else {
			icon = ":radio.fill:"
		}
		icon2 := ":pause.circle.fill:"

		t1 := state.Title1
		t2 := state.Title2
		t3 := state.Title3
		t4 := state.StreamFormat

		if state.Service == "AirPlay" {
			if state.Mute == "0" {
				c = fmt.Sprintf("%s/Volume?mute=1", bluePlayerUrl)
				icon2 = ":speaker.wave.1.fill:"
			} else if state.Mute == "1" {
				c = fmt.Sprintf("%s/Volume?mute=0", bluePlayerUrl)
				icon2 = ":speaker.slash.fill:"
			}
			cmd = createCommand(c)
		}
		l1 := fmt.Sprintf("%s %s", icon, t1)
		l2 := fmt.Sprintf("%s %s", icon, t2)
		l3 := fmt.Sprintf("%s %s", icon, t3)
		app.StatusLine(l2).DropDown(false).Length(MAX)
		app.StatusLine(l1).DropDown(false).Length(MAX)
		if state.Service != "Spotify" {
			app.StatusLine(l3).DropDown(false).Length(MAX)
		}
		s1 := fmt.Sprintf("%s %s: %s", icon2, state.ServiceName, t3)
		s2 := t4
		submenu.Line(s1).Length(MAX).Command(cmd)
		submenu.Line(s2).Alternate(true)

		log.Printf("Created status display for stream state (service: %s)", state.Service)

	case "pause":
		icon := ":pause.circle.fill:"
		icon2 := ":play.circle.fill:"
		l1 := fmt.Sprintf("%s %s", icon, state.Title1)
		s1 := fmt.Sprintf("%s %s: %s", icon2, state.ServiceName, state.Title1)
		app.StatusLine(l1).DropDown(false).Length(MAX)
		submenu.Line(s1).Length(MAX).Command(cmd)

	case "stop":
		icon := ":stop.circle.fill:"
		icon2 := ":play.circle.fill:"
		c := fmt.Sprintf("%s/Play", bluePlayerUrl)
		cmd = createCommand(c)
		l1 := fmt.Sprintf("%s %s", icon, state.State)
		app.StatusLine(l1).DropDown(false).Length(MAX)
		if state.Service != "" {
			s1 := fmt.Sprintf("%s %s: %s", icon2, state.ServiceName, state.Title1)
			submenu.Line(s1).Length(MAX).Command(cmd)
		}
		log.Printf("Created status display for stop state")

	default:
		log.Printf("Unhandled player state: %s", state.State)

		// Add a generic display for unhandled states
		icon := ":questionmark.circle.fill:"
		l1 := fmt.Sprintf("%s %s", icon, state.State)
		app.StatusLine(l1).DropDown(false).Length(MAX)
		submenu.Line(fmt.Sprintf("State: %s", state.State))
		submenu.Line(fmt.Sprintf("Service: %s", state.Service))
		submenu.Line(fmt.Sprintf("Title: %s", state.Title1))
		log.Printf("Created generic status display")
	}
}

// addRadioPresets adds radio presets to the menu
func addRadioPresets(submenu *bitbar.SubMenu, presetsUrl, bluePlayerUrl string) {
	xmlBytes, err := getXML(presetsUrl)
	if err != nil {
		submenu.Line("⚠️ Error loading presets").Color("red")
		log.Printf("Failed to get presets XML: %v", err)
		return
	}

	var presets Presets
	if err := xml.Unmarshal(xmlBytes, &presets); err != nil {
		submenu.Line("⚠️ Error parsing presets").Color("red")
		log.Printf("Failed to parse presets XML: %v", err)
		return
	}

	// Add presets directly to the main menu
	log.Printf("Adding %d radio presets", len(presets.Preset))
	for _, p := range presets.Preset {
		// Use SF Symbol for each preset, matching the previous implementation
		l := fmt.Sprintf(":star.fill: %s - %s", p.ID, p.Name)
		c := fmt.Sprintf("%s/Preset?id=%s", bluePlayerUrl, p.ID)
		cmd := createCommand(c)
		submenu.Line(l).Command(cmd)
	}

	if len(presets.Preset) == 0 {
		submenu.Line("No presets found").Color("gray")
	}
}

// addVolumeInfo adds volume information to the menu
// Returns the parsed volume status for use in other sections
// getVolumeSymbol dynamically selects the appropriate SF Symbol for volume levels
func getVolumeSymbol(level int, isMuted bool) string {
	if isMuted {
		return ":speaker.slash.fill:"
	}

	switch {
	case level <= 0:
		return ":speaker.slash.fill:" // Muted or zero volume
	case level > 0 && level < 33:
		return ":speaker.wave.1.fill:" // Low volume
	case level >= 33 && level < 66:
		return ":speaker.wave.2.fill:" // Medium volume
	case level >= 66 && level < 100:
		return ":speaker.wave.3.fill:" // High volume
	default:
		return ":megaphone.fill:" // Max volume
	}
}

func addVolumeInfo(submenu *bitbar.SubMenu, volumeUrl string) *VolumeStatus {
	log.Printf("Getting volume info")
	xmlBytes, err := getXML(volumeUrl)
	if err != nil {
		submenu.Line("⚠️ Could not get volume").Color("red")
		log.Printf("Failed to get volume XML: %v", err)
		return nil
	}

	// Log the volume XML response for debugging
	if len(xmlBytes) > 0 {
		previewLength := 200
		if len(xmlBytes) < previewLength {
			previewLength = len(xmlBytes)
		}
		log.Printf("Volume XML response preview: %s", string(xmlBytes[:previewLength]))
	}

	var volStatus VolumeStatus
	if err := xml.Unmarshal(xmlBytes, &volStatus); err != nil {
		submenu.Line("⚠️ Error parsing volume data").Color("red")
		log.Printf("Failed to parse volume XML: %v", err)

		// Try to create a default volume object so the UI doesn't completely fail
		log.Printf("Creating default volume status object")
		return &VolumeStatus{
			Db:    -30.0, // Default reasonable value
			Mute:  0,
			Level: 50, // Default reasonable value
			Etag:  "unknown",
		}
	}

	log.Printf("Current volume: %d%%, %.1f dB, Muted: %v", volStatus.Level, volStatus.Db, volStatus.Mute == 1)

	// Determine volume symbol and color
	volumeSymbol := getVolumeSymbol(volStatus.Level, volStatus.Mute == 1)

	// Display volume information - dB as primary, percentage as alternate
	if volStatus.Mute == 1 {
		// For muted state, show in red
		submenu.Line(fmt.Sprintf("%s Volume: %.1f dB (Muted)", volumeSymbol, volStatus.Db)).Color("red")
		submenu.Line(fmt.Sprintf("%s Volume: %d%% (Muted)", volumeSymbol, volStatus.Level)).Alternate(true).Color("red")
	} else {
		// For active state, use color based on volume level
		var volColor string
		if volStatus.Level > 85 {
			volColor = "red"
		} else if volStatus.Level > 60 {
			volColor = "orange"
		} else if volStatus.Level > 30 {
			volColor = "blue"
		} else {
			volColor = "green"
		}

		// Main volume display
		submenu.Line(fmt.Sprintf("%s Volume: %.1f dB", volumeSymbol, volStatus.Db)).Color(volColor)

		// Alternate lines for volume and fine control
		submenu.Line(fmt.Sprintf("%s Volume: %d%%", volumeSymbol, volStatus.Level)).Alternate(true).Color(volColor)

		// Fine volume control as alternate lines
		submenu.Line(":speaker.wave.3.fill: Volume Up (1dB)").Command(
			createVolumeCommand(volumeUrl, map[string]string{"db": "1.0"}),
		).Alternate(true)
		submenu.Line(":speaker.wave.1.fill: Volume Down (1dB)").Command(
			createVolumeCommand(volumeUrl, map[string]string{"db": "-1.0"}),
		).Alternate(true)
	}

	return &volStatus
}

// addVolumePresets adds volume preset buttons to the menu
func addVolumePresets(submenu *bitbar.SubMenu, bluePlayerUrl string, volStatus *VolumeStatus) {
	if volStatus == nil {
		return
	}

	log.Printf("Adding volume presets")

	// Volume presets in descending order
	volumePresets := []struct {
		Label string
		Level int
	}{
		{":megaphone.fill: Max (100%)", 100},
		{":speaker.wave.3.fill: High (80%)", 80},
		{":speaker.wave.2.fill: Medium (50%)", 50},
		{":speaker.wave.1.fill: Low (20%)", 20},
	}

	// Highlight the current preset that's closest to the current volume
	currentVol := volStatus.Level
	for _, preset := range volumePresets {
		presetCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"level": strconv.Itoa(preset.Level)})
		line := submenu.Line(preset.Label).Command(presetCmd)

		// Highlight if this is the active preset (within 5%)
		if preset.Level-5 <= currentVol && currentVol <= preset.Level+5 {
			line.Color("blue")
		}
	}
}

// addMuteToggle adds the mute/unmute toggle button
func addMuteToggle(submenu *bitbar.SubMenu, bluePlayerUrl string, volStatus *VolumeStatus) {
	if volStatus == nil {
		return
	}

	if volStatus.Mute == 1 {
		unmuteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "0"})
		submenu.Line(":speaker.wave.2.fill: Unmute").Command(unmuteCmd)
	} else {
		muteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "1"})
		submenu.Line(":speaker.slash.fill: Mute").Command(muteCmd)
	}
}

// Note: The following functions have been removed:
// - addFineVolumeControls: fine volume controls are now integrated into addVolumeInfo
// - addAudioInfo: audio quality information has been removed for simplification
