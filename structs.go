package main

import (
	"encoding/xml"
)

// StateXML represents the structure of the BluOS /Status response XML
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

// Presets represents the structure of the BluOS /Presets response XML
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
