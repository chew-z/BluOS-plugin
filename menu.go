package main

import (
	"encoding/xml"
	"fmt"
	"log"
	"strconv"

	"github.com/johnmccabe/go-bitbar"
)

// buildPlayerMenu builds the main menu structure based on player state and volume info
func buildPlayerMenu(app bitbar.Plugin, bluePlayerUrl string) {
	log.Printf("Building player menu for %s", bluePlayerUrl)
	statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
	presetsUrl := fmt.Sprintf("%s/Presets", bluePlayerUrl)
	volumeUrl := fmt.Sprintf("%s/Volume", bluePlayerUrl)

	submenu := app.NewSubMenu()
	
	// Process status data and create status bar
	createStatusDisplay(app, submenu, statusUrl, bluePlayerUrl)
	
	// Add radio presets section
	submenu.Line("--- Radio Presets ---")
	addRadioPresets(submenu, presetsUrl, bluePlayerUrl)
	
	// Add volume section
	submenu.Line("--- Volume Controls ---")
	volStatus := addVolumeInfo(submenu, volumeUrl)
	
	// Add volume presets
	submenu.Line("--- Volume Presets ---")
	addVolumePresets(submenu, bluePlayerUrl, volStatus)
	
	// Add mute toggle
	addMuteToggle(submenu, bluePlayerUrl, volStatus)
	
	// Add advanced options submenu
	advancedSubmenu := submenu.NewSubMenu()
	advancedSubmenu.Line("🎛️ Advanced Options").Font("Menlo-Bold")
	
	// Add fine volume controls to advanced submenu
	addFineVolumeControls(advancedSubmenu, bluePlayerUrl)
	
	// Add audio information to advanced submenu if player is active
	addAudioInfo(advancedSubmenu, statusUrl)

	log.Printf("Menu building completed")
}

// createStatusDisplay adds the status bar display based on player state
func createStatusDisplay(app bitbar.Plugin, submenu *bitbar.SubMenu, statusUrl, bluePlayerUrl string) {
	log.Printf("Creating status display")
	xmlBytes, err := getXML(statusUrl)
	if err != nil {
		submenu.Line(err.Error()).Color("red").Length(MAX)
		log.Printf("Failed to get XML: %v", err)
		return
	}
	
	var state StateXML
	if err := xml.Unmarshal(xmlBytes, &state); err != nil {
		submenu.Line("Error parsing status data").Color("red")
		log.Printf("Failed to parse status XML: %v", err)
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
		icon := ":powercord:"
		l1 := fmt.Sprintf("%s connecting", icon)
		app.StatusLine(l1).DropDown(false).Length(MAX)

	case "play":
		icon := ":music.note.list:"
		if state.Shuffle == "1" {
			icon = ":shuffle:"
		}
		icon2 := ":pause.fill:"
		l1 := fmt.Sprintf("%s %s", icon, state.Name)
		l2 := fmt.Sprintf("%s %s", icon, state.Album)
		l3 := fmt.Sprintf("%s %s", icon, state.Artist)
		s1 := fmt.Sprintf("%s %s: %s", icon2, state.ServiceName, state.Name)
		s2 := fmt.Sprintf("%s %s", state.Quality, state.StreamFormat)

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
			icon = ":music.note.tv:"
		} else if state.Service == "Capture" {
			icon = ":tv:"
			log.Printf("Found Capture service (HDMI ARC)")
		} else {
			icon = ":radio:"
		}
		icon2 := ":pause.fill:"

		t1 := state.Title1
		t2 := state.Title2
		t3 := state.Title3
		t4 := state.StreamFormat

		if state.Service == "AirPlay" {
			if state.Mute == "0" {
				c = fmt.Sprintf("%s/Volume?mute=1", bluePlayerUrl)
				icon2 = ":speaker.zzz:"
			} else if state.Mute == "1" {
				c = fmt.Sprintf("%s/Volume?mute=0", bluePlayerUrl)
				icon2 = ":speaker.slash:"
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
		icon := ":pause.fill:"
		icon2 := ":play.fill:"
		l1 := fmt.Sprintf("%s %s", icon, state.Title1)
		s1 := fmt.Sprintf("%s %s: %s", icon2, state.ServiceName, state.Title1)
		app.StatusLine(l1).DropDown(false).Length(MAX)
		submenu.Line(s1).Length(MAX).Command(cmd)

	case "stop":
		icon := ":stop.fill:"
		icon2 := ":play.fill:"
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
		icon := ":questionmark.circle:"
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
		// Use a radio icon 📻 for each preset
		l := fmt.Sprintf("📻 %s - %s", p.ID, p.Name)
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
func addVolumeInfo(submenu *bitbar.SubMenu, volumeUrl string) *VolumeStatus {
	log.Printf("Getting volume info")
	xmlBytes, err := getXML(volumeUrl)
	if err != nil {
		submenu.Line("⚠️ Could not get volume").Color("red")
		log.Printf("Failed to get volume XML: %v", err)
		return nil
	}
	
	var volStatus VolumeStatus
	if err := xml.Unmarshal(xmlBytes, &volStatus); err != nil {
		submenu.Line("⚠️ Error parsing volume data").Color("red")
		log.Printf("Failed to parse volume XML: %v", err)
		return nil
	}
	
	log.Printf("Current volume: %d%%, %.1f dB, Muted: %v", volStatus.Level, volStatus.Db, volStatus.Mute == 1)
	
	// Display volume information - dB as primary, percentage as alternate
	if volStatus.Mute == 1 {
		// For muted state, show in red
		submenu.Line(fmt.Sprintf("🔇 Volume: %.1f dB (Muted)", volStatus.Db)).Color("red")
		submenu.Line(fmt.Sprintf("🔇 Volume: %d%% (Muted)", volStatus.Level)).Alternate(true).Color("red")
	} else {
		// For active state, use color based on volume level
		volColor := ""  // default color
		if volStatus.Level > 85 {
			volColor = "red"
		} else if volStatus.Level > 60 {
			volColor = "orange"
		} else if volStatus.Level > 30 {
			volColor = "blue"
		} else {
			volColor = "green"
		}
		
		submenu.Line(fmt.Sprintf("🔊 Volume: %.1f dB", volStatus.Db)).Color(volColor)
		submenu.Line(fmt.Sprintf("🔊 Volume: %d%%", volStatus.Level)).Alternate(true).Color(volColor)
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
		{"📢 Max (100%)", 100},
		{"🔊 High (80%)", 80},
		{"🔉 Medium (50%)", 50},
		{"🔈 Low (20%)", 20},
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
		submenu.Line("🔈 Unmute").Command(unmuteCmd)
	} else {
		muteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "1"})
		submenu.Line("🔇 Mute").Command(muteCmd)
	}
}

// addFineVolumeControls adds fine volume adjustment controls
func addFineVolumeControls(submenu *bitbar.SubMenu, bluePlayerUrl string) {
	submenu.Line("--- Fine Volume Control ---")
	
	// 0.5dB adjustments as primary, 1dB as alternate
	upCmd05 := createVolumeCommand(bluePlayerUrl, map[string]string{"db": "0.5"})
	upCmd1 := createVolumeCommand(bluePlayerUrl, map[string]string{"db": "1.0"})
	downCmd05 := createVolumeCommand(bluePlayerUrl, map[string]string{"db": "-0.5"})
	downCmd1 := createVolumeCommand(bluePlayerUrl, map[string]string{"db": "-1.0"})
	
	submenu.Line("🔊 Volume Up (0.5dB)").Command(upCmd05)
	submenu.Line("🔊 Volume Up (1dB)").Command(upCmd1).Alternate(true)
	submenu.Line("🔉 Volume Down (0.5dB)").Command(downCmd05)
	submenu.Line("🔉 Volume Down (1dB)").Command(downCmd1).Alternate(true)
}

// addAudioInfo adds audio quality information when available
func addAudioInfo(submenu *bitbar.SubMenu, statusUrl string) {
	xmlBytes, err := getXML(statusUrl)
	if err != nil {
		return
	}
	
	var state StateXML
	if err := xml.Unmarshal(xmlBytes, &state); err != nil || state.State == "stop" {
		return
	}
	
	submenu.Line("--- Audio Information ---")
	
	if state.Quality != "" {
		submenu.Line(fmt.Sprintf("🎧 Quality: %s", state.Quality))
	}
	
	if state.StreamFormat != "" {
		submenu.Line(fmt.Sprintf("🎛 Format: %s", state.StreamFormat))
	}
}
