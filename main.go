package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
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

const defaultDbStep = 2.0 // Typical dB step for volume up/down

// sendVolumeCommand sends a command to the /Volume endpoint with parameters
func sendVolumeCommand(playerUrl string, params map[string]string) (*VolumeStatus, error) {
	baseURL := fmt.Sprintf("%s/Volume", playerUrl)
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing base URL %s: %w", baseURL, err)
	}

	query := reqURL.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	reqURL.RawQuery = query.Encode()

	xmlBytes, err := getXML(reqURL.String())
	if err != nil {
		return nil, fmt.Errorf("volume command failed: %w", err)
	}

	var status VolumeStatus
	if err := xml.Unmarshal(xmlBytes, &status); err != nil {
		log.Printf("Failed to parse volume XML: %v\nXML: %s", err, string(xmlBytes))
		return nil, fmt.Errorf("XML parsing error: %w", err)
	}

	return &status, nil
}

// VolumeUp increases the volume by the default dB step
func VolumeUp(playerUrl string) (*VolumeStatus, error) {
	params := map[string]string{
		"db": fmt.Sprintf("%.1f", defaultDbStep),
	}
	return sendVolumeCommand(playerUrl, params)
}

// VolumeDown decreases the volume by the default dB step
func VolumeDown(playerUrl string) (*VolumeStatus, error) {
	params := map[string]string{
		"db": fmt.Sprintf("%.1f", -defaultDbStep),
	}
	return sendVolumeCommand(playerUrl, params)
}

// ToggleMute toggles the mute state
func ToggleMute(playerUrl string) (*VolumeStatus, error) {
	// Get current status first
	status, err := sendVolumeCommand(playerUrl, nil)
	if err != nil {
		return nil, err
	}

	// Toggle the mute state
	muteValue := "1"
	if status.Mute == 1 {
		muteValue = "0"
	}

	params := map[string]string{
		"mute": muteValue,
	}
	return sendVolumeCommand(playerUrl, params)
}

// createVolumeCommand creates a bitbar command for volume control operations
func createVolumeCommand(playerUrl string, params map[string]string) bitbar.Cmd {
	baseURL := fmt.Sprintf("%s/Volume", playerUrl)
	reqURL, _ := url.Parse(baseURL)

	query := reqURL.Query()
	for key, value := range params {
		query.Set(key, value)
	}
	reqURL.RawQuery = query.Encode()

	return bitbar.Cmd{
		Bash:     "curl",
		Params:   []string{"-sf", reqURL.String()},
		Terminal: BoolPointer(false),
		Refresh:  BoolPointer(true),
	}
}

// SetVolume sets the volume to a specific level (0-100)
func SetVolume(playerUrl string, level int) (*VolumeStatus, error) {
	if level < 0 || level > 100 {
		return nil, fmt.Errorf("invalid volume level %d (must be 0-100)", level)
	}

	params := map[string]string{
		"level": strconv.Itoa(level),
	}
	return sendVolumeCommand(playerUrl, params)
}
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

	if xmlBytes, err := getXML(statusUrl); err != nil {
		submenu.Line(err.Error()).Color("red").Length(MAX)
		log.Printf("Failed to get XML: %v", err)
	} else {
		var state StateXML
		xml.Unmarshal(xmlBytes, &state)
		c := fmt.Sprintf("%s/Pause?toggle=1", bluePlayerUrl)
		cmd := bitbar.Cmd{
			Bash:     "curl",
			Params:   []string{"-sf", c},
			Terminal: BoolPointer(false),
			Refresh:  BoolPointer(true),
		}
		if state.State == "connecting" {
			icon := ":powercord:"
			l1 := fmt.Sprintf("%s connecting", icon)
			app.StatusLine(l1).DropDown(false).Length(MAX)
		} else if state.State == "play" {
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
		} else if state.State == "stream" {
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
			// if t1 == "mpv" && t2 == "mpv" {
			// 	t1, t2, t3, t4 = mpv()
			// }
			if state.Service == "AirPlay" {
				if state.Mute == "0" {
					c = fmt.Sprintf("%s/Volume?mute=1", bluePlayerUrl)
					icon2 = ":speaker.zzz:"
				} else if state.Mute == "1" {
					c = fmt.Sprintf("%s/Volume?mute=0", bluePlayerUrl)
					icon2 = ":speaker.slash:"
				}
				cmd = bitbar.Cmd{
					Bash:     "curl",
					Params:   []string{"-sf", c},
					Terminal: BoolPointer(false),
					Refresh:  BoolPointer(true),
				}
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
		} else if state.State == "pause" {
			icon := ":pause.fill:"
			icon2 := ":play.fill:"
			l1 := fmt.Sprintf("%s %s", icon, state.Title1)
			s1 := fmt.Sprintf("%s %s: %s", icon2, state.ServiceName, state.Title1)
			app.StatusLine(l1).DropDown(false).Length(MAX)
			submenu.Line(s1).Length(MAX).Command(cmd)
		} else if state.State == "stop" {
			icon := ":stop.fill:"
			icon2 := ":play.fill:"
			c := fmt.Sprintf("%s/Play", bluePlayerUrl)
			cmd.Params = []string{"-sf", c}
			l1 := fmt.Sprintf("%s %s", icon, state.State)
			app.StatusLine(l1).DropDown(false).Length(MAX)
			if state.Service != "" {
				s1 := fmt.Sprintf("%s %s: %s", icon2, state.ServiceName, state.Title1)
				submenu.Line(s1).Length(MAX).Command(cmd)
			}
		}
	}
	if xmlBytes, err := getXML(presetsUrl); err != nil {
		submenu.Line(err.Error()).Color("red").Length(MAX)
		log.Printf("Failed to get XML: %v", err)
	} else {
		var presets Presets
		xml.Unmarshal(xmlBytes, &presets)
		for _, p := range presets.Preset {
			l := fmt.Sprintf("%s - %s", p.ID, p.Name)
			c := fmt.Sprintf("%s/Preset?id=%s", bluePlayerUrl, p.ID)
			cmd := bitbar.Cmd{
				Bash:     "curl",
				Params:   []string{"-sf", c},
				Terminal: BoolPointer(false),
				Refresh:  BoolPointer(true),
			}
			submenu.Line(l).Command(cmd)
		}
	}

	// Add volume controls section
	submenu.Line("--- Volume Controls ---").Alternate(true)

	// Get current volume status
	if xmlBytes, err := getXML(fmt.Sprintf("%s/Volume", bluePlayerUrl)); err != nil {
		submenu.Line("⚠️ Could not get volume").Color("red")
	} else {
		var volStatus VolumeStatus
		if err := xml.Unmarshal(xmlBytes, &volStatus); err != nil {
			submenu.Line("⚠️ Error parsing volume data").Color("red")
			log.Printf("Failed to parse volume XML: %v", err)
		} else {
			// Display volume information
			if volStatus.Mute == 1 {
				submenu.Line(fmt.Sprintf("🔇 Volume: Muted (%d%%)", volStatus.Level))
			} else {
				submenu.Line(fmt.Sprintf("🔊 Volume: %d%%", volStatus.Level))
			}

			// Add dB information if available
			if volStatus.Db != 0 {
				submenu.Line(fmt.Sprintf("🎛 Volume dB: %.1f dB", volStatus.Db)).Alternate(true)
			}

			// Volume controls using the helper function
			volumeUpCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"db": fmt.Sprintf("%.1f", defaultDbStep)})
			volumeDownCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"db": fmt.Sprintf("%.1f", -defaultDbStep)})

			submenu.Line("🔊 Volume Up").Command(volumeUpCmd)
			submenu.Line("🔉 Volume Down").Command(volumeDownCmd)

			// Mute controls
			muteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "1"})
			unmuteCmd := createVolumeCommand(bluePlayerUrl, map[string]string{"mute": "0"})

			if volStatus.Mute == 1 {
				submenu.Line("🔈 Unmute").Command(unmuteCmd)
			} else {
				submenu.Line("🔇 Mute").Command(muteCmd)
			}

			// Add volume preset buttons
			submenu.Line("--- Volume Presets ---").Alternate(true)
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
				submenu.Line(preset.Label).Command(presetCmd)
			}

			// Add fine volume controls
			submenu.Line("--- Fine Volume Control ---").Alternate(true)
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
				submenu.Line(step.Label).Command(fineCmd)
			}
		}
	}

	// Add audio quality info section if player is active
	if xmlBytes, err := getXML(statusUrl); err == nil {
		var state StateXML
		if err := xml.Unmarshal(xmlBytes, &state); err == nil && state.State != "stop" {
			submenu.Line("--- Audio Information ---").Alternate(true)

			if state.Quality != "" {
				submenu.Line(fmt.Sprintf("🎧 Quality: %s", state.Quality))
			}

			if state.StreamFormat != "" {
				submenu.Line(fmt.Sprintf("🎛 Format: %s", state.StreamFormat))
			}
		}
	}
	goto AppRender
AppRender:
	app.Render()
}

// mpv() - mpv is above providing essential 'now playing' info to anyone
// so we must read it from ipc socket (that must be configured and running)
// https://mpv.io/manual/master/#json-ipc
// func mpv() (string, string, string, string) {
// 	var t1, t2, t3, t4 string
// 	socketPath := fmt.Sprintf("%s/mpv_socket", TMP)
// 	conn := mpvipc.NewConnection(socketPath)
// 	err := conn.Open()
// 	if err != nil {
// 		log.Println(err.Error())
// 	}
// 	defer conn.Close()
// 	if prop, err := conn.Get("filtered-metadata"); err != nil {
// 		log.Println(err.Error())
// 	} else {
// 		// log.Printf("playing: %v", prop)
// 		m := prop.(map[string]interface{})
// 		// log.Printf("playing: %s", m["icy-title"])
// 		if m["icy-title"] != nil {
// 			t1 = m["icy-title"].(string)
// 		}
// 		if m["icy-name"] != nil {
// 			t2 = m["icy-name"].(string)
// 		}
// 		if m["title"] != nil {
// 			t1 = m["title"].(string)
// 		}
// 		if m["artist"] != nil {
// 			t2 = m["artist"].(string)
// 		}
// 		if m["Title"] != nil {
// 			t1 = m["Title"].(string)
// 		}
// 		if m["Artist"] != nil {
// 			t2 = m["Artist"].(string)
// 		}
// 		if m["TITLE"] != nil {
// 			t1 = m["TITLE"].(string)
// 		}
// 		if m["ARTIST"] != nil {
// 			t2 = m["ARTIST"].(string)
// 		}
// 	}
// 	if prop, err := conn.Get("volume"); err != nil {
// 		log.Println(err.Error())
// 	} else {
// 		// log.Printf("playing: %v", prop)
// 		dB := vol2db(prop.(float64))
// 		t3 = fmt.Sprintf("mpv: %.0f dB", dB)
// 	}
// 	if prop, err := conn.Get("audio-codec-name"); err != nil {
// 		log.Println(err.Error())
// 	} else {
// 		// log.Printf("playing: %v", prop)
// 		t4 += fmt.Sprintf("codec: %s", prop.(string))
// 	}
// 	if prop, err := conn.Get("audio-params/samplerate"); err != nil {
// 		log.Println(err.Error())
// 	} else {
// 		// log.Printf("playing: %v", prop)
// 		t4 += fmt.Sprintf(" samplerate: %.0f", prop.(float64))
// 	}
// 	return t1, t2, t3, t4
// }

// assumes samples are multiplied by (vol/100)^3
// https://github.com/mpv-player/mpv/blob/master/player/audio.c#L161
func vol2db(vol float64) float64 {
	return 60.0 * math.Log(vol/100.0) / math.Log(10.0)
}

// tweaked from: https://stackoverflow.com/a/42718113/1170664
func getXML(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}

func BoolPointer(b bool) *bool {
	return &b
}

type StateXML struct {
	Text    string `xml:",chardata"`
	Etag    string `xml:"etag,attr"`
	Actions struct {
		Text   string `xml:",chardata"`
		Action []struct {
			Text     string `xml:",chardata"`
			Name     string `xml:"name,attr"`
			URL      string `xml:"url,attr"`
			Icon     string `xml:"icon,attr"`
			State    string `xml:"state,attr"`
			AttrText string `xml:"text,attr"`
			Type     string `xml:"type,attr"`
		} `xml:"action"`
	} `xml:"actions,omitempty"`
	Album           string `xml:"album,omitempty"`
	Artist          string `xml:"artist,omitempty"`
	CanMovePlayback string `xml:"canMovePlayback"`
	CanSeek         string `xml:"canSeek"`
	CurrentImage    string `xml:"currentImage"`
	Cursor          string `xml:"cursor"`
	Db              string `xml:"db"`
	Image           string `xml:"image"`
	Indexing        string `xml:"indexing"`
	Mid             string `xml:"mid"`
	Mode            string `xml:"mode"`
	Mute            string `xml:"mute"`
	Name            string `xml:"name,omitempty"`
	Pid             string `xml:"pid"`
	PresetID        string `xml:"preset_id"`
	Prid            string `xml:"prid"`
	Quality         string `xml:"quality"`
	Repeat          string `xml:"repeat"`
	Service         string `xml:"service"`
	ServiceIcon     string `xml:"serviceIcon"`
	ServiceName     string `xml:"serviceName"`
	Shuffle         string `xml:"shuffle"`
	Sid             string `xml:"sid"`
	Sleep           string `xml:"sleep"`
	Song            string `xml:"song"`
	State           string `xml:"state"`
	StreamFormat    string `xml:"streamFormat"`
	StreamUrl       string `xml:"streamUrl"`
	SyncStat        string `xml:"syncStat"`
	Title1          string `xml:"title1"`
	Title2          string `xml:"title2"`
	Title3          string `xml:"title3"`
	Totlen          string `xml:"totlen,omitempty"`
	Volume          string `xml:"volume"`
	Secs            string `xml:"secs"`
}

type Presets struct {
	XMLName xml.Name `xml:"presets"`
	Text    string   `xml:",chardata"`
	Prid    string   `xml:"prid,attr"`
	Preset  []struct {
		Text  string `xml:",chardata"`
		URL   string `xml:"url,attr"`
		ID    string `xml:"id,attr"`
		Name  string `xml:"name,attr"`
		Image string `xml:"image,attr"`
	} `xml:"preset"`
}

// VolumeStatus represents the structure of the BluOS /Volume response XML
type VolumeStatus struct {
	XMLName    xml.Name `xml:"volume"`
	Db         float64  `xml:"db,attr"`         // Volume level in dB
	Mute       int      `xml:"mute,attr"`       // 1 if muted, 0 if not
	MuteDb     *float64 `xml:"muteDb,attr"`     // Volume level in dB before mute
	MuteVolume *int     `xml:"muteVolume,attr"` // Volume level before mute
	OffsetDb   float64  `xml:"offsetDb,attr"`   // Volume offset in dB
	Etag       string   `xml:"etag,attr"`       // Entity tag for caching
	Level      int      `xml:",chardata"`       // Current volume level (0-100)
}
