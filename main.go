package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"

	"github.com/johnmccabe/go-bitbar"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

var (
	myConfig map[string]string
	osxCmd   = "/System/Library/PrivateFrameworks/Apple80211.framework/Versions/Current/Resources/airport"
	osxArgs  = "-I"
)

func init() {
	var err error
	myConfig, err = godotenv.Read("/Users/rrj/Projekty/SwiftBar/.env")
	if err != nil {
		log.Fatalln("Error loading .env file")
	}
}

func main() {
	// log.Println(myConfig)
	blueWiFi := myConfig["BLUE_WIFI"]
	bluePlayer := myConfig["BLUE_URL"]
	statusUrl := fmt.Sprintf("%s/Status", bluePlayer)
	presetsUrl := fmt.Sprintf("%s/Presets", bluePlayer)

	app := bitbar.New()
	submenu := app.NewSubMenu()
	if ssid := getSSID(); !strings.Contains(ssid, blueWiFi) {
		// app.StatusLine(ssid)
		app.StatusLine(":play.slash.fill:").Color("red")
		goto AppRender
	}
	if xmlBytes, err := getXML(statusUrl); err != nil {
		submenu.Line(err.Error()).Color("red").Length(25)
		log.Printf("Failed to get XML: %v", err)
	} else {
		var state StateXML
		var icon string
		xml.Unmarshal(xmlBytes, &state)
		if state.State == "play" {
			icon = ":play.fill:"
			l1 := fmt.Sprintf("%s %s", icon, state.Title1)
			l2 := fmt.Sprintf("%s %s", icon, state.Title2)
			l3 := fmt.Sprintf("%s %s", icon, state.Title3)
			app.StatusLine(l1).DropDown(false)
			app.StatusLine(l2).DropDown(false)
			app.StatusLine(l3).DropDown(false)
		} else if state.State == "stream" {
			icon = ":play:"
			l1 := fmt.Sprintf("%s %s", icon, state.Title1)
			l2 := fmt.Sprintf("%s %s", icon, state.Title2)
			app.StatusLine(l1).DropDown(false).Length(50)
			app.StatusLine(l2).DropDown(false)
		} else if state.State == "pause" {
			icon = ":pause.fill:"
			l1 := fmt.Sprintf("%s %s", icon, state.Title1)
			app.StatusLine(l1).DropDown(false).Length(50)
		} else if state.State == "stop" {
			icon = ":stop.fill:"
			l1 := fmt.Sprintf("%s %s", icon, state.State)
			app.StatusLine(l1).DropDown(false).Length(50)
		}
	}
	if xmlBytes, err := getXML(presetsUrl); err != nil {
		submenu.Line(err.Error()).Color("red").Length(25)
		log.Printf("Failed to get XML: %v", err)
	} else {
		var presets Presets
		xml.Unmarshal(xmlBytes, &presets)
		for _, p := range presets.Preset {
			l := fmt.Sprintf("%s - %s", p.ID, p.Name)
			submenu.Line(l)
			// log.Println(p.ID, p.Name)
		}
	}
	goto AppRender
AppRender:
	app.Render()
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

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}

func getSSID() string {

	cmd := exec.Command(osxCmd, osxArgs)
	stdout, _ := cmd.StdoutPipe()
	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		return "Could not get SSID"
	}
	defer cmd.Wait()

	var airport string
	if b, err := ioutil.ReadAll(stdout); err == nil {
		airport += (string(b) + "\n")
	}
	re := regexp.MustCompile(`[^B]SSID:\s.*`)
	name := strings.TrimPrefix(re.FindString(airport), " SSID: ")
	if len(name) <= 1 {
		return "Could not get SSID"
	} else {
		return name
	}
}

type StateXML struct {
	Text            string `xml:",chardata"`
	Etag            string `xml:"etag,attr"`
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
