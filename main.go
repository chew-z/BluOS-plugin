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
	ssid     string
)

func init() {
	var err error
	myConfig, err = godotenv.Read("/Users/rrj/Projekty/SwiftBar/.env")
	if err != nil {
		log.Fatalln("Error loading .env file")
	}
}

func main() {
	log.Println(myConfig)
	wifi := myConfig["WIFI"]
	bluePlayer := myConfig["BLUE"]
	statusUrl := fmt.Sprintf("%s/Status", bluePlayer)
	app := bitbar.New()
	submenu := app.NewSubMenu()
	if ssid := getSSID(); !strings.Contains(ssid, wifi) {
		app.StatusLine(ssid)
		goto AppRender
	}
	if xmlBytes, err := getXML(statusUrl); err != nil {
		submenu.Line(err.Error()).Color("red").Length(25)
		log.Printf("Failed to get XML: %v", err)
	} else {
		var state StateXML
		xml.Unmarshal(xmlBytes, &state)
		l1 := fmt.Sprintf("[%s] %s", state.State, state.Title1)
		l2 := fmt.Sprintf("%s", state.Title2)
		l3 := fmt.Sprintf("%s", state.Title3)

		app.StatusLine(l1)
		app.StatusLine(l2)
		app.StatusLine(l3)

		m := fmt.Sprintf("[%s] %s - %s", state.Song, state.Secs, state.Service)
		submenu.Line(m)
		a := fmt.Sprintf("[%s] %s", state.StreamFormat, state.ServiceName)
		submenu.Line(a).Alternate(true)
		// submenu.Line(m).Href(quote.webURL).Color(color)
		// submenu.Line(a).Alternate(true).Href(quote.webURL).Color(color)
		// log.Printf("[%s] %s -- %s [%s]\n", state.State, state.Title2, state.ServiceName, state.StreamFormat)
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
	stdout, err := cmd.StdoutPipe()
	panicIf(err)
	// start the command after having set up the pipe
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	defer cmd.Wait()
	var str string

	if b, err := ioutil.ReadAll(stdout); err == nil {
		str += (string(b) + "\n")
	}
	r := regexp.MustCompile(`s*SSID: (.+)s*`)
	name := r.FindAllStringSubmatch(str, -1)
	// log.Println(name[0][0])
	if len(name[0][0]) <= 1 {
		return "Could not get SSID"
	} else {
		return name[0][0]
	}
}

func panicIf(err error) {
	if err != nil {
		panic(err)
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
