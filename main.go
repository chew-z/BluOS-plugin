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
	myConfig, err = godotenv.Read(fmt.Sprintf("%s/.env", os.Getenv("SWIFTBAR_PLUGINS_PATH")))
	if err != nil {
		log.Fatalln("Error loading .env file")
	}
}

func main() {
	if m, e := strconv.Atoi(myConfig["MAX"]); e == nil {
		MAX = m
	}
	// blueWiFi := myConfig["BLUE_WIFI"]
	bluePlayerUrl := myConfig["BLUE_URL"]
	statusUrl := fmt.Sprintf("%s/Status", bluePlayerUrl)
	presetsUrl := fmt.Sprintf("%s/Presets", bluePlayerUrl)

	app := bitbar.New()
	submenu := app.NewSubMenu()

	// Process status data
	if xmlBytes, err := getXML(statusUrl); err != nil {
		submenu.Line(err.Error()).Color("red").Length(MAX)
		log.Printf("Failed to get XML: %v", err)
	} else {
		var state StateXML
		xml.Unmarshal(xmlBytes, &state)
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
		}
	}

	// Add presets
	if xmlBytes, err := getXML(presetsUrl); err != nil {
		submenu.Line(err.Error()).Color("red").Length(MAX)
		log.Printf("Failed to get XML: %v", err)
	} else {
		var presets Presets
		xml.Unmarshal(xmlBytes, &presets)
		
		// Create a submenu for presets
		presetsSubmenu := submenu.NewSubMenu()
		presetsSubmenu.Line("Radio Presets").Font("Menlo-Bold")
		
		for _, p := range presets.Preset {
			l := fmt.Sprintf("%s - %s", p.ID, p.Name)
			c := fmt.Sprintf("%s/Preset?id=%s", bluePlayerUrl, p.ID)
			cmd := createCommand(c)
			presetsSubmenu.Line(l).Command(cmd)
		}
	}

	// Add compact volume controls to main menu
	submenu.Line("--- Controls ---")
	
	// Get current volume status
	if xmlBytes, err := getXML(fmt.Sprintf("%s/Volume", bluePlayerUrl)); err != nil {
		submenu.Line("⚠️ Could not get volume").Color("red")
	} else {
		var volStatus VolumeStatus
		if err := xml.Unmarshal(xmlBytes, &volStatus); err != nil {
			submenu.Line("⚠️ Error parsing volume data").Color("red")
			log.Printf("Failed to parse volume XML: %v", err)
		} else {
			// Display current volume in main menu
			volumeText := ""
			if volStatus.Mute == 1 {
				volumeText = fmt.Sprintf("🔇 Muted (%d%%)", volStatus.Level)
			} else {
				volumeText = fmt.Sprintf("🔊 %d%%", volStatus.Level)
			}
			submenu.Line(volumeText)
			
			// Add basic volume controls to main menu
			volumeUpCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"db": fmt.Sprintf("%.1f", defaultDbStep)})
			volumeDownCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"db": fmt.Sprintf("%.1f", -defaultDbStep)})
			
			muteUnmuteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{})
			if volStatus.Mute == 1 {
				muteUnmuteCmd = createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "0"})
				submenu.Line("🔈 Unmute").Command(muteUnmuteCmd)
			} else {
				muteUnmuteCmd = createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "1"})
				submenu.Line("🔇 Mute").Command(muteUnmuteCmd)
			}
			
			// Create a volume submenu for detailed controls
			volumeSubmenu := submenu.NewSubMenu()
			volumeSubmenu.Line("Volume Controls").Font("Menlo-Bold")
			
			// Add current volume and dB info to submenu
			if volStatus.Mute == 1 {
				volumeSubmenu.Line(fmt.Sprintf("🔇 Volume: Muted (%d%%)", volStatus.Level))
			} else {
				volumeSubmenu.Line(fmt.Sprintf("🔊 Volume: %d%%", volStatus.Level))
			}
			
			if volStatus.Db != 0 {
				volumeSubmenu.Line(fmt.Sprintf("🎛 Volume dB: %.1f dB", volStatus.Db))
			}
			
			// Add volume up/down controls to submenu
			volumeSubmenu.Line("🔊 Volume Up").Command(volumeUpCmd)
			volumeSubmenu.Line("🔉 Volume Down").Command(volumeDownCmd)
			
			// Add volume presets to submenu
			volumeSubmenu.Line("--- Volume Presets ---")
			volumePresets := []struct {
				Label string
				Level int
			}{
				{"🔈 Low (20%)", 20},
				{"🔉 Medium (50%)", 50},
				{"🔊 High (80%)", 80},
				{"📢 Max (100%)", 100},
			}
			
			for _, preset := range volumePresets {
				presetCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"level": strconv.Itoa(preset.Level)})
				volumeSubmenu.Line(preset.Label).Command(presetCmd)
			}
			
			// Add fine volume controls to submenu
			volumeSubmenu.Line("--- Fine Volume Control ---")
			fineSteps := []struct {
				Label  string
				DbStep float64
			}{
				{"🔊 Volume Up (1dB)", 1.0},
				{"🔉 Volume Down (1dB)", -1.0},
				{"🔊 Volume Up (0.5dB)", 0.5},
				{"🔉 Volume Down (0.5dB)", -0.5},
			}
			
			for _, step := range fineSteps {
				fineCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"db": fmt.Sprintf("%.1f", step.DbStep)})
				volumeSubmenu.Line(step.Label).Command(fineCmd)
			}
		}
	}

	// Add audio quality info section if player is active
	if xmlBytes, err := getXML(statusUrl); err == nil {
		var state StateXML
		if err := xml.Unmarshal(xmlBytes, &state); err == nil && state.State != "stop" {
			// Create audio info submenu
			audioInfoSubmenu := submenu.NewSubMenu()
			audioInfoSubmenu.Line("Audio Information").Font("Menlo-Bold")
			
			if state.Quality != "" {
				audioInfoSubmenu.Line(fmt.Sprintf("🎧 Quality: %s", state.Quality))
			}
			
			if state.StreamFormat != "" {
				audioInfoSubmenu.Line(fmt.Sprintf("🎛 Format: %s", state.StreamFormat))
			}
		}
	}

	app.Render()
}
