package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/johnmccabe/go-bitbar"
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
	reqURL, err := url.Parse(baseURL)
	if err != nil {
		// For a command creation function, log the error but continue with a default
		log.Printf("Error parsing URL %s: %v", baseURL, err)
		return createCommand(baseURL) // Fallback to basic command
	}

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

// Db2vol converts dB to volume percentage (0-100)
// This is the inverse of the vol2db function and maintains compatibility
func Db2vol(db float64) float64 {
	return 100.0 * math.Pow(10.0, db/60.0)
}

// tweaked from: https://stackoverflow.com/a/42718113/1170664
func getXML(url string) ([]byte, error) {
	log.Printf("Fetching XML from: %s", url)

	// Set timeout for requests
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("Error connecting to %s: %v", url, err)
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Bad status code from %s: %d", url, resp.StatusCode)
		return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body from %s: %v", url, err)
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	log.Printf("Successfully retrieved %d bytes from %s", len(data), url)
	return data, nil
}

func BoolPointer(b bool) *bool {
	return &b
}

// createCommand is a helper to create curl commands for BluOS API endpoints
func createCommand(url string) bitbar.Cmd {
	return bitbar.Cmd{
		Bash:     "curl",
		Params:   []string{"-sf", url},
		Terminal: BoolPointer(false),
		Refresh:  BoolPointer(true),
	}
}
